// Copyright 2017 The kubecfg authors
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

	"github.com/ksonnet/ksonnet/pkg/actions"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	vApplyComponent = "apply-components"
	vApplyCreate    = "apply-create"
	vApplyGcTag     = "apply-gc-tag"
	vApplyDryRun    = "apply-dry-run"
	vApplySkipGc    = "apply-skip-gc"

	applyShortDesc = "Apply local Kubernetes manifests (components) to remote clusters"
	applyLong      = `
The ` + "`apply`" + `command uses local manifest(s) to update (and optionally create)
Kubernetes resources on a remote cluster. This cluster is determined by the
mandatory ` + "`<env-name>`" + ` argument.

The manifests themselves correspond to the components of your app, and reside
in your app's ` + "`components/`" + ` directory. When applied, the manifests are fully
expanded using the parameters of the specified environment.

By default, all component manifests are applied. To apply a subset of components,
use the ` + "`--component` " + `flag, as seen in the examples below.

Note that this command needs to be run *within* a ksonnet app directory.

### Related Commands

* ` + "`ks diff` " + `— ` + diffShortDesc + `
* ` + "`ks delete` " + `— ` + deleteShortDesc + `

### Syntax
`

	applyExample = `
# Create or update all resources described in the ksonnet application, specifically
# the ones running in the 'dev' environment. This command works in any subdirectory
# of the app.
#
# This essentially deploys all components in the 'components/' directory.
ks apply dev

# Similar to the previous command, but does not immediately execute. Use this to
# see a preview of the cluster-changing actions.
ks apply dev --dry-run

# Create or update the single 'guestbook-ui' component of a ksonnet app, specifically
# the instance running in the 'dev' environment.
#
# This essentially deploys 'components/guestbook-ui.jsonnet'.
ks apply dev -c guestbook-ui

# Create or update multiple components in a ksonnet application (e.g. 'guestbook-ui'
# and 'nginx-depl') for the 'dev' environment. Does not create resources that are
# not already present on the cluster.
#
# This essentially deploys 'components/guestbook-ui.jsonnet' and
# 'components/nginx-depl.jsonnet'.
ks apply dev -c guestbook-ui -c nginx-depl --create false
`
)

func newApplyCmd(a app.App) *cobra.Command {
	applyClientConfig := client.NewDefaultClientConfig(a)

	applyCmd := &cobra.Command{
		Use:   "apply <env-name> [-c <component-name>] [--dry-run]",
		Short: applyShortDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			var envName string
			if len(args) == 1 {
				envName = args[0]
			}

			m := map[string]interface{}{
				actions.OptionApp:            a,
				actions.OptionClientConfig:   applyClientConfig,
				actions.OptionComponentNames: viper.GetStringSlice(vApplyComponent),
				actions.OptionCreate:         viper.GetBool(vApplyCreate),
				actions.OptionDryRun:         viper.GetBool(vApplyDryRun),
				actions.OptionEnvName:        envName,
				actions.OptionGcTag:          viper.GetString(vApplyGcTag),
				actions.OptionSkipGc:         viper.GetBool(vApplySkipGc),
			}

			fmt.Println("extract jsonnet flag")
			if err := extractJsonnetFlags(a, "apply"); err != nil {
				return errors.Wrap(err, "handle jsonnet flags")
			}

			return runAction(actionApply, m)
		},
		Long:    applyLong,
		Example: applyExample,
	}

	applyClientConfig.BindClientGoFlags(applyCmd)
	bindJsonnetFlags(applyCmd, "apply")

	applyCmd.Flags().StringSliceP(flagComponent, shortComponent, nil, "Name of a specific component (multiple -c flags accepted, allows YAML, JSON, and Jsonnet)")
	viper.BindPFlag(vApplyComponent, applyCmd.Flags().Lookup(flagComponent))

	applyCmd.Flags().Bool(flagCreate, true, "Option to create resources if they do not already exist on the cluster")
	viper.BindPFlag(vApplyCreate, applyCmd.Flags().Lookup(flagCreate))

	applyCmd.Flags().Bool(flagSkipGc, false, "Option to skip garbage collection, even with --"+flagGcTag+" specified")
	viper.BindPFlag(vApplySkipGc, applyCmd.Flags().Lookup(flagSkipGc))

	applyCmd.Flags().String(flagGcTag, "", "A tag that's (1) added to all updated objects (2) used to garbage collect existing objects that are no longer in the manifest")
	viper.BindPFlag(vApplyGcTag, applyCmd.Flags().Lookup(flagGcTag))

	applyCmd.Flags().Bool(flagDryRun, false, "Option to preview the list of operations without changing the cluster state")
	viper.BindPFlag(vApplyDryRun, applyCmd.Flags().Lookup(flagDryRun))

	return applyCmd
}
