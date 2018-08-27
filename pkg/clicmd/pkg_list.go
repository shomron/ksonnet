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

package clicmd

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/ksonnet/ksonnet/pkg/actions"
	"github.com/spf13/cobra"
)

const (
	vPkgListInstalled = "pkg-list-installed"
	vPkgListOutput    = "pkg-list-output"
)

var (
	pkgListLong = `
The ` + "`list`" + ` command outputs a table that describes all *known* packages (not
necessarily downloaded, but available from existing registries). This includes
the following info:

1. Library name
2. Registry name
3. Installed status — an asterisk indicates 'installed'

### Related Commands

* ` + "`ks pkg install` " + `— ` + pkgShortDesc["install"] + `
* ` + "`ks pkg describe` " + `— ` + pkgShortDesc["describe"] + `
* ` + "`ks registry describe` " + `— ` + regShortDesc["describe"] + `

### Syntax
`
)

func newPkgListCmd() *cobra.Command {
	pkgListCmd := &cobra.Command{
		Use:   "list",
		Short: pkgShortDesc["list"],
		Long:  pkgListLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("Command 'pkg list' does not take arguments")
			}

			m := map[string]interface{}{
				actions.OptionInstalled:     viper.GetBool(vPkgListInstalled),
				actions.OptionOutput:        viper.GetString(vPkgListOutput),
				actions.OptionTLSSkipVerify: viper.GetBool(flagTLSSkipVerify),
			}

			return runAction(actionPkgList, m)
		},
	}

	addCmdOutput(pkgListCmd, vPkgListOutput)
	pkgListCmd.Flags().Bool(flagInstalled, false, "Only list installed packages")
	viper.BindPFlag(vPkgListInstalled, pkgListCmd.Flags().Lookup(flagInstalled))

	return pkgListCmd
}
