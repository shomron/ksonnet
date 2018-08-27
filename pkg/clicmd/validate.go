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
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"

	"github.com/ksonnet/ksonnet/pkg/actions"
	"github.com/ksonnet/ksonnet/pkg/client"
)

const (
	vValidateComponent = "validate-component"
	valShortDesc       = "Check generated component manifests against the server's API"
)

var (
	validateLong = `
The ` + "`validate`" + ` command checks that an application or file is compliant with the
server APIs Kubernetes specification. Note that this command actually communicates
*with* the server for the specified ` + "`<env-name>`" + `, so it only works if your
$KUBECONFIG specifies a valid kubeconfig file.

When NO component is specified (no ` + "`-c`" + ` flag), this command checks all of
the files in the ` + "`components/`" + ` directory. This is the same as what would
get deployed to your cluster with ` + "`ks apply <env-name>`" + `.

When a component IS specified via the ` + "`-c`" + ` flag, this command only checks
the manifest for that particular component.

### Related Commands

* ` + "`ks show` " + `— ` + showShortDesc + `
* ` + "`ks apply` " + `— ` + applyShortDesc + `

### Syntax
`
	validateExample = `
# Validate all resources described in the ksonnet app, against the server
# specified by the 'dev' environment.
# NOTE: Make sure your current $KUBECONFIG matches the 'dev' cluster info
ksonnet validate dev

# Validate resources from the 'redis' component only, against the server specified
# by the 'prod' environment
# NOTE: Make sure your current $KUBECONFIG matches the 'prod' cluster info
ksonnet validate prod -c redis
`
)

func newValidateCmd(fs afero.Fs) *cobra.Command {
	validateClientConfig := client.NewDefaultClientConfig()

	validateCmd := &cobra.Command{
		Use:     "validate <env-name> [-c <component-name>]",
		Short:   valShortDesc,
		Long:    validateLong,
		Example: validateExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			var envName string
			if len(args) == 1 {
				envName = args[0]
			}

			m := map[string]interface{}{
				actions.OptionEnvName:        envName,
				actions.OptionModule:         "",
				actions.OptionComponentNames: viper.GetStringSlice(vValidateComponent),
				actions.OptionClientConfig:   validateClientConfig,
			}

			if err := extractJsonnetFlags(fs, "validate"); err != nil {
				return errors.Wrap(err, "handle jsonnet flags")
			}

			return runAction(actionValidate, m)
		},
	}

	addEnvCmdFlags(validateCmd)
	bindJsonnetFlags(validateCmd, "validate")
	validateClientConfig.BindClientGoFlags(validateCmd)

	viper.BindPFlag(vValidateComponent, validateCmd.Flag(flagComponent))

	return validateCmd
}
