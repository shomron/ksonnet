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
	"github.com/ksonnet/ksonnet/pkg/actions"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	upgradeShortDesc  = "Upgrade ks configuration"
	flagUpgradeDryRun = "dry-run"
)

var (
	upgradeLong = `
The upgrade command upgrades a ksonnet application to the latest version.

### Syntax
`
	upgradeExample = `
# Upgrade ksonnet application in dry-run mode to see the changes to be performed by the
# upgrade process.
ks upgrade --dry-run

# Upgrade ksonnet application. This will update app.yaml to apiVersion 0.1.0
# and migrate environment spec.json files to ` + "`" + `app.yaml` + "`" + `.
ks upgrade
`
)

func newUpgradeCmd() *cobra.Command {
	upgradeCmd := &cobra.Command{
		Use:     "upgrade [--dry-run]",
		Short:   upgradeShortDesc,
		Long:    upgradeLong,
		Example: upgradeExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			dryRun, err := cmd.Flags().GetBool(flagUpgradeDryRun)
			if err != nil {
				return err
			}

			m := map[string]interface{}{
				actions.OptionDryRun:           dryRun,
				actions.OptionTLSSkipVerify:    viper.GetBool(flagTLSSkipVerify),
				actions.OptionSkipCheckUpgrade: true,
			}

			return runAction(actionUpgrade, m)
		},
	}

	upgradeCmd.Flags().Bool(flagUpgradeDryRun, false, "Dry-run upgrade process. Prints out changes.")

	return upgradeCmd
}
