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
	"github.com/ksonnet/ksonnet/pkg/actions"
	"github.com/ksonnet/ksonnet/pkg/client"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	vDeleteComponent   = "delete-components"
	vDeleteGracePeriod = "delete-grace-period"

	deleteShortDesc = "Remove component-specified Kubernetes resources from remote clusters"
	deleteLong      = `
The ` + "`delete`" + ` command removes Kubernetes resources (described in local
*component* manifests) from a cluster. This cluster is determined by the mandatory
` + "`<env-name>`" + `argument.

An entire ksonnet application can be removed from a cluster, or just its specific
components.

**This command can be considered the inverse of the ` + "`ks apply`" + ` command.**

### Related Commands

* ` + "`ks diff` " + `— Compare manifests, based on environment or location (local or remote)
* ` + "`ks apply` " + `— ` + applyShortDesc + `

### Syntax
`
	deleteExample = `# Delete resources from the 'dev' environment, based on ALL of the manifests in your
# ksonnet app's 'components/' directory. This command works in any subdirectory
# of the app.
ks delete dev

# Delete resources described by the 'nginx' component. $KUBECONFIG is overridden by
# the CLI-specified './kubeconfig', so these changes are deployed to the current
# context's cluster (not the 'default' environment)
ks delete --kubeconfig=./kubeconfig -c nginx`
)

func newDeleteCmd(fs afero.Fs) *cobra.Command {
	deleteClientConfig := client.NewDefaultClientConfig()

	deleteCmd := &cobra.Command{
		Use:     "delete [env-name] [-c <component-name>]",
		Short:   deleteShortDesc,
		Long:    deleteLong,
		Example: deleteExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			var envName string
			if len(args) == 1 {
				envName = args[0]
			}

			m := map[string]interface{}{
				actions.OptionClientConfig:   deleteClientConfig,
				actions.OptionComponentNames: viper.GetStringSlice(vDeleteComponent),
				actions.OptionEnvName:        envName,
				actions.OptionGracePeriod:    viper.GetInt64(vDeleteGracePeriod),
			}

			if err := extractJsonnetFlags(fs, "delete"); err != nil {
				return errors.Wrap(err, "handle jsonnet flags")
			}

			return runAction(actionDelete, m)
		},
	}

	deleteClientConfig.BindClientGoFlags(deleteCmd)
	bindJsonnetFlags(deleteCmd, "delete")

	deleteCmd.Flags().StringSliceP(flagComponent, shortComponent, nil, "Name of a specific component (multiple -c flags accepted, allows YAML, JSON, and Jsonnet)")
	viper.BindPFlag(vDeleteComponent, deleteCmd.Flags().Lookup(flagComponent))

	deleteCmd.Flags().Int64(flagGracePeriod, -1, "Number of seconds given to resources to terminate gracefully. A negative value is ignored")
	viper.BindPFlag(vDeleteGracePeriod, deleteCmd.Flags().Lookup(flagGracePeriod))

	return deleteCmd
}
