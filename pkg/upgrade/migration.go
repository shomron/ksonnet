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

package upgrade

import (
	"regexp"

	"github.com/blang/semver"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/registry"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Migration interface {
	// Match returns true if the migration can upgrade the specified version.
	Match(string) bool

	Migrate() (string, error)
}

type migration struct {
	name string

	// Match returns true if the migration can upgrade the specified version.
	matchFn func(string) bool

	migrateFn func() (string, error)
}

func (m *migration) Match(ver string) bool {
	return m.matchFn(ver)
}

func (m *migration) Migrate() (string, error) {
	return m.migrateFn()
}

// MatchFunc returns true if a version matches a compatible range.
type MatchFunc func(string) bool

// SemverMatcher implements MatchFunc using semver ranges.
func SemverMatcher(versionRange string) MatchFunc {
	match := semver.MustParseRange(versionRange)
	return func(ver string) bool {
		sv, err := semver.Parse(ver)
		if err != nil {
			return false
		}
		return match(sv)
	}
}

// RunMigrations runs all matching migrations sequentially
func RunMigrations(a app.App, ver string) error {
	log := log.WithField("action", "upgrade.RunMigrations")

	// Migrations should be ordered, oldest to newest.
	var migrations = []migration{
		migration{
			name:    "0.1.0 -> 0.2.0",
			matchFn: SemverMatcher("0.1.0"),
			migrateFn: func() (string, error) {
				m020 := app020Migration{a}
				return m020.Migrate()
			},
		},
	}

	current := ver
	var err error
	for _, m := range migrations {
		if !m.Match(current) {
			continue
		}

		log.Debugf("Migrating %v", m.name)
		current, err = m.Migrate()
		if err != nil {
			return errors.Wrapf(err, "migrating %v", m.name)
		}
	}

	return nil
}

// Migrate to 020 from previous versions
type app020Migration struct {
	app app.App
}

func (m *app020Migration) Migrate() (string, error) {
	if m == nil {
		return "", errors.Errorf("nil receiver")
	}

	if err := m.migrateVendorCache(); err != nil {

	}

	// Upgraded to 0.2.0
	return "0.2.0", nil
}

var removeVersionPattern = regexp.MustCompile("(.*)@.*$")

// Upgrade unversioned vendor cache to versioned vendor cache
// e.g. vendor/<registry>/<pkg> -> vendor/<registry>/<pkg>@<version>
func (m *app020Migration) migrateVendorCache() error {
	if m == nil {
		return errors.Errorf("nil receiver")
	}
	a := m.app
	if a == nil {
		return errors.Errorf("nil app")
	}
	fs := a.Fs()
	if fs == nil {
		return errors.Errorf("nil filesystem interface")
	}

	pm := registry.NewPackageManager(a)
	pkgs, err := pm.Packages()
	if err != nil {
		return errors.Wrapf(err, "resolving packages")
	}

	for _, p := range pkgs {
		if p.Version() == "" {
			// Skip unversioned packages
			continue
		}

		versioned := p.Path()

		ok, err := afero.Exists(fs, versioned)
		if err != nil {
			return errors.Wrapf(err, "checking package: %v", p)
		}
		if ok {
			// Already upgraded
			continue
		}

		// Check for unversioned path
		unversioned := string(removeVersionPattern.ReplaceAll([]byte(versioned), []byte("$1")))
		ok, err = afero.Exists(fs, unversioned)
		if err != nil {
			return errors.Wrapf(err, "checking path: %v", unversioned)
		}
		if !ok {
			// Nothing to upgrade, the package is simply missing
			continue
		}

		// Ok, time to upgrade -> unversioned -> versioned
		if err := fs.Rename(unversioned, versioned); err != nil {
			return errors.Wrapf(err, "renaming %v -> %v", unversioned, versioned)
		}
	}
	return nil
}
