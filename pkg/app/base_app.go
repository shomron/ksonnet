// Copyright 2018 The ksonnet authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package app

import (
	"path/filepath"
	"sync"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type baseApp struct {
	root string
	fs   afero.Fs

	config    *Spec
	overrides *Override

	mu sync.Mutex
}

func newBaseApp(fs afero.Fs, root string) *baseApp {
	return &baseApp{
		fs:     fs,
		root:   root,
		config: &Spec{},
		overrides: &Override{
			Environments: EnvironmentConfigs{},
			Registries:   RegistryConfigs{},
		},
	}
}

func (ba *baseApp) CurrentEnvironment() string {
	currentPath := filepath.Join(ba.root, currentEnvName)
	data, err := afero.ReadFile(ba.fs, currentPath)
	if err != nil {
		return ""
	}

	return string(data)
}

func (ba *baseApp) SetCurrentEnvironment(name string) error {
	currentPath := filepath.Join(ba.root, currentEnvName)
	return afero.WriteFile(ba.fs, currentPath, []byte(name), DefaultFilePermissions)
}

func (ba *baseApp) configPath() string {
	return filepath.Join(ba.root, "app.yaml")
}

func (ba *baseApp) overridePath() string {
	return filepath.Join(ba.root, "app.override.yaml")
}

func (ba *baseApp) save() error {
	log := log.WithField("action", "baseApp.save")
	log.Debugf("called [%p]", ba)

	ba.mu.Lock()
	defer ba.mu.Unlock()

	log.Debugf("saving app version %v", ba.config.APIVersion)
	configData, err := yaml.Marshal(ba.config)
	if err != nil {
		return errors.Wrap(err, "convert application configuration to YAML")
	}

	if err = afero.WriteFile(ba.fs, ba.configPath(), configData, DefaultFilePermissions); err != nil {
		return errors.Wrapf(err, "write %s", ba.configPath())
	}

	if err = removeOverride(ba.fs, ba.root); err != nil {
		return errors.Wrap(err, "clean overrides")
	}

	if ba.overrides.IsDefined() {
		return SaveOverride(defaultYAMLEncoder, ba.fs, ba.root, ba.overrides)
	}

	return nil
}

func (ba *baseApp) load() error {
	log := log.WithField("action", "baseApp.load")
	log.Debugf("called [%p]", ba)

	ba.mu.Lock()
	defer ba.mu.Unlock()

	configData, err := afero.ReadFile(ba.fs, ba.configPath())
	if err != nil {
		return errors.Wrapf(err, "read %s", ba.configPath())
	}

	var config Spec
	if err = yaml.Unmarshal(configData, &config); err != nil {
		return errors.Wrapf(err, "unmarshal application YAML config")
	}

	exists, err := afero.Exists(ba.fs, ba.overridePath())
	if err != nil {
		return err
	}

	if len(config.Environments) == 0 {
		config.Environments = EnvironmentConfigs{}
	}

	if len(config.Registries) == 0 {
		config.Registries = RegistryConfigs{}
	}

	override := Override{
		Environments: EnvironmentConfigs{},
		Registries:   RegistryConfigs{},
	}
	if exists {
		overrideData, err := afero.ReadFile(ba.fs, ba.overridePath())
		if err != nil {
			return errors.Wrapf(err, "read %s", ba.overridePath())
		}
		if err = yaml.Unmarshal(overrideData, &override); err != nil {
			return errors.Wrapf(err, "unmarshal override YAML config")
		}

		if err = override.Validate(); err != nil {
			return errors.Wrap(err, "validating override")
		}

		if len(override.Environments) == 0 {
			override.Environments = EnvironmentConfigs{}
		}

		if len(override.Registries) == 0 {
			override.Registries = RegistryConfigs{}
		}

		for k := range override.Registries {
			override.Registries[k].isOverride = true
		}

		for k := range override.Environments {
			override.Environments[k].isOverride = true
		}

	}

	ba.overrides = &override
	ba.config = &config

	return ba.config.validate()
}

func (ba *baseApp) AddRegistry(newReg *RegistryConfig, isOverride bool) error {
	if err := ba.load(); err != nil {
		return errors.Wrap(err, "load configuration")
	}

	if newReg.Name == "" {
		return ErrRegistryNameInvalid
	}

	if isOverride {
		_, exists := ba.overrides.Registries[newReg.Name]
		if exists {
			return ErrRegistryExists
		}

		newReg.isOverride = true

		ba.overrides.Registries[newReg.Name] = newReg
		return ba.save()
	}

	_, exists := ba.config.Registries[newReg.Name]
	if exists {
		return ErrRegistryExists
	}

	newReg.isOverride = false
	ba.config.Registries[newReg.Name] = newReg

	return ba.save()
}

func (ba *baseApp) UpdateLib(name string, libSpec *LibraryConfig) error {
	if err := ba.load(); err != nil {
		return errors.Wrap(err, "load configuration")
	}

	ba.config.Libraries[name] = libSpec
	return ba.save()
}

// UpdateRegistry updates a registry spec and persists in app[.override].yaml
func (ba *baseApp) UpdateRegistry(spec *RegistryConfig) error {
	if err := ba.load(); err != nil {
		return errors.Wrap(err, "load configuration")
	}

	if spec.Name == "" {
		return ErrRegistryNameInvalid
	}

	// Figure out where the registry is defined (app or override)
	var ok, okOverride bool
	if ba.config != nil {
		_, ok = ba.config.Registries[spec.Name]
	}
	if ba.overrides != nil {
		_, okOverride = ba.overrides.Registries[spec.Name]
	}

	if !ok && !okOverride {
		return errors.Errorf("registry not found: %v", spec.Name)
	}

	if ok && okOverride {
		return errors.Errorf("registry %v found in both app.yaml and app.override.yaml", spec.Name)
	}

	if ok {
		ba.config.Registries[spec.Name] = spec
	} else {
		ba.overrides.Registries[spec.Name] = spec
	}

	return ba.save()
}

func (ba *baseApp) Fs() afero.Fs {
	return ba.fs
}

func (ba *baseApp) Root() string {
	return ba.root
}

func (ba *baseApp) EnvironmentParams(envName string) (string, error) {
	if envName == "" {
		return "", errors.New("environment name is blank")
	}
	envParamsPath := filepath.Join(ba.Root(), EnvironmentDirName, envName, "params.libsonnet")
	b, err := afero.ReadFile(ba.Fs(), envParamsPath)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read params for environment %s", envName)
	}

	return string(b), nil
}

func (ba *baseApp) VendorPath() string {
	return filepath.Join(ba.Root(), "vendor")
}
