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

package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ksonnet/ksonnet/pkg/parts"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// Local is a package based on vendored contents.
type Local struct {
	pkg
	config *parts.Spec
}

var _ Package = (*Local)(nil)

// NewLocal creates an instance of Local.
func NewLocal(a app.App, name, registryName string, version string, installChecker InstallChecker) (*Local, error) {
	if installChecker == nil {
		installChecker = &DefaultInstallChecker{App: a}
	}

	versionedDir := buildPath(a, registryName, name, version)
	partsPath := filepath.Join(versionedDir, "parts.yaml")
	b, err := afero.ReadFile(a.Fs(), partsPath)
	if err != nil {
		return nil, errors.Wrap(err, "reading package configuration")
	}

	config, err := parts.Unmarshal(b)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling package configuration")
	}

	return &Local{
		pkg: pkg{
			registryName: registryName,
			name:         name,
			version:      version,

			a:              a,
			installChecker: installChecker,
		},
		config: config,
	}, nil
}

// Description returns the description for the package. The description
// is ready from the package's parts.yaml.
func (l *Local) Description() string {
	return l.config.Description
}

// Prototypes returns prototypes for this package. Prototypes are defined in the
// package's `prototypes` directory.
func (l *Local) Prototypes() (prototype.Prototypes, error) {
	var prototypes prototype.Prototypes

	protoPath := filepath.Join(l.Path(), "prototypes")
	exists, err := afero.DirExists(l.a.Fs(), protoPath)
	if err != nil {
		return nil, err
	}

	if !exists {
		return prototypes, nil
	}

	err = afero.Walk(l.a.Fs(), protoPath, func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() || filepath.Ext(path) != ".jsonnet" {
			return nil
		}

		data, err := afero.ReadFile(l.a.Fs(), path)
		if err != nil {
			return err
		}

		spec, err := prototype.DefaultBuilder(string(data))
		if err != nil {
			return err
		}

		prototypes = append(prototypes, spec)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return prototypes, nil
}

// buildPath returns local directory for vendoring a package.
func buildPath(a app.App, registry string, name string, version string) string {
	if a == nil || registry == "" || name == "" || version == "" {
		return ""
	}

	// Construct package path: `vendor/<registry>/<pkg>@<version>`
	versionedDir := fmt.Sprintf("%v@%v", name, version)
	path := filepath.Join(a.VendorPath(), registry, versionedDir)
	return path
}

// Path returns local directory for vendoring the package.
func (l *Local) Path() string {
	if l == nil {
		return ""
	}
	if l.a == nil {
		return ""
	}

	return buildPath(l.a, l.registryName, l.name, l.version)
}
