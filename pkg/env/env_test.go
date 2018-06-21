// Copyright 2018 The kubecfg authors
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

package env

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/pkg"
	pmocks "github.com/ksonnet/ksonnet/pkg/pkg/mocks"
	rmocks "github.com/ksonnet/ksonnet/pkg/registry/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddJPaths(t *testing.T) {
	withJsonnetPaths(func() {
		AddJPaths("/vendor")
		require.Equal(t, []string{"/vendor"}, componentJPaths)
	})
}

func TestAddExtVar(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	testCases := []struct {
		name string
		args args
	}{
		{
			name: "add a key and value",
			args: args{
				key:   "key",
				value: "value",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withJsonnetPaths(func() {
				AddExtVar(tc.args.key, tc.args.value)
				require.Equal(t, tc.args.value, componentExtVars[tc.args.key])
			})
		})
	}
}

func TestAddExtVarFile(t *testing.T) {
	type args struct {
		key  string
		file string
	}
	testCases := []struct {
		name          string
		args          args
		expectedValue string
		stagePath     string
		isErr         bool
	}{
		{
			name: "add a key and value",
			args: args{
				key:  "key",
				file: "/app/value.txt",
			},
			expectedValue: "value",
			stagePath:     "/app/value.txt",
		},
		{
			name: "add a key and value",
			args: args{
				key:  "key",
				file: "/app/value.txt",
			},
			isErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
				withJsonnetPaths(func() {
					if tc.stagePath != "" {
						test.StageFile(t, fs, "value.txt", tc.stagePath)
					}
					err := AddExtVarFile(a, tc.args.key, tc.args.file)
					if tc.isErr {
						require.Error(t, err)
						return
					}
					require.NoError(t, err)
					require.Equal(t, tc.expectedValue, componentExtVars[tc.args.key])

				})
			})
		})
	}
}

func TestAddTlaVar(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	testCases := []struct {
		name          string
		args          args
		expectedKey   string
		expectedValue string
	}{
		{
			name: "add a key and value",
			args: args{
				key:   "key",
				value: "value",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			withJsonnetPaths(func() {
				AddTlaVar(tc.args.key, tc.args.value)
				require.Equal(t, tc.args.value, componentTlaVars[tc.args.key])
			})
		})
	}
}

func TestAddTlaVarFile(t *testing.T) {
	type args struct {
		key  string
		file string
	}
	testCases := []struct {
		name          string
		args          args
		expectedValue string
		stagePath     string
		isErr         bool
	}{
		{
			name: "add a key and value",
			args: args{
				key:  "key",
				file: "/app/value.txt",
			},
			expectedValue: "value",
			stagePath:     "/app/value.txt",
		},
		{
			name: "add a key and value",
			args: args{
				key:  "key",
				file: "/app/value.txt",
			},
			isErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
				withJsonnetPaths(func() {
					if tc.stagePath != "" {
						test.StageFile(t, fs, "value.txt", tc.stagePath)
					}
					err := AddTlaVarFile(a, tc.args.key, tc.args.file)
					if tc.isErr {
						require.Error(t, err)
						return
					}
					require.NoError(t, err)
					require.Equal(t, tc.expectedValue, componentTlaVars[tc.args.key])

				})
			})
		})
	}
}

func withJsonnetPaths(fn func()) {
	ogComponentJPaths := componentJPaths
	ogComponentExtVars := componentExtVars
	ogComponentTlaVars := componentTlaVars

	defer func() {
		componentJPaths = ogComponentJPaths
		componentExtVars = ogComponentExtVars
		componentTlaVars = ogComponentTlaVars
	}()

	fn()
}

func TestEvaluate(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		envSpec := &app.EnvironmentConfig{
			Path: "default",
			Destination: &app.EnvironmentDestinationSpec{
				Server:    "http://example.com",
				Namespace: "default",
			},
		}
		a.On("Environment", "default").Return(envSpec, nil)

		test.StageFile(t, fs, "main.jsonnet", "/app/environments/default/main.jsonnet")

		components, err := ioutil.ReadFile(filepath.FromSlash("testdata/evaluate/components.jsonnet"))
		require.NoError(t, err)

		got, err := Evaluate(a, "default", string(components), "")
		require.NoError(t, err)

		test.AssertOutput(t, "evaluate/out.jsonnet", got)
	})
}

func TestMainFile(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		envSpec := &app.EnvironmentConfig{}
		a.On("Environment", "default").Return(envSpec, nil)

		test.StageFile(t, fs, "main.jsonnet", "/app/environments/main.jsonnet")

		got, err := MainFile(a, "default")
		require.NoError(t, err)

		test.AssertOutput(t, "main.jsonnet", got)
	})
}

func Test_upgradeArray(t *testing.T) {
	snippet, err := ioutil.ReadFile(filepath.FromSlash("testdata/upgradeArray/in.jsonnet"))
	require.NoError(t, err)

	got, err := upgradeArray(string(snippet))
	require.NoError(t, err)

	test.AssertOutput(t, "upgradeArray/out.jsonnet", got)
}

func Test_buildPackagePaths(t *testing.T) {
	makePackage := func(registry string, name string, version string, installed bool) pkg.Package {
		p := new(pmocks.Package)
		p.On("Name").Return(name)
		p.On("RegistryName").Return(registry)
		p.On("Version").Return(version)
		p.On("IsInstalled").Return(installed)
		p.On("Path").Return(
			filepath.Join("vendor", registry, fmt.Sprintf("%s@%s", name, version)),
		)
		return p
	}

	// Rig a package manager to return a fixed set of packages for the environment
	r := "incubator"
	e := &app.EnvironmentConfig{Name: "default"}
	pkgByName := map[string]pkg.Package{
		"nginx": makePackage(r, "nginx", "1.2.3", true),
		"mysql": makePackage(r, "mysql", "00112233ff", true),
	}
	packages := make([]pkg.Package, 0, len(pkgByName))
	for _, p := range pkgByName {
		packages = append(packages, p)
	}
	pm := new(rmocks.PackageManager)
	pm.On("PackagesForEnv", e).Return(packages, nil)

	// Stage some packages to copy
	fs := afero.NewMemMapFs()
	for _, p := range packages {
		test.StageDir(t, fs, filepath.Join("packages", p.Name()), p.Path())
	}

	results, err := buildPackagePaths(pm, e)
	require.NoError(t, err)

	assert.Equal(t, len(pkgByName), len(results), "result length")
	for name, path := range results {
		p, ok := pkgByName[name]
		assert.True(t, ok, "unexpected package: %v", name)
		if p != nil {
			assert.Equal(t, path, p.Path(), "package %v vendor path mismatch", name)
		}
	}

}
