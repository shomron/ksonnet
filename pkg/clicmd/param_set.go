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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	vParamSetEnv          = "param-set-env"
	vParamSetAsString     = "param-set-as-string"
	vParamSetResolveImage = "param-set-resolve-image"

	paramSetLong = `
The ` + "`set`" + ` command sets component or environment parameters such as replica count
or name. Parameters are set individually, one at a time. All of these changes are
reflected in the ` + "`params.libsonnet`" + ` files.

For more details on how parameters are organized, see ` + "`ks param --help`" + `.

*(If you need to customize multiple parameters at once, we suggest that you modify
your ksonnet application's ` + " `components/params.libsonnet` " + `file directly. Likewise,
for greater customization of environment parameters, we suggest modifying the
` + " `environments/:name/params.libsonnet` " + `file.)*

### Related Commands

* ` + "`ks param diff` " + `— ` + paramShortDesc["diff"] + `
* ` + "`ks apply` " + `— ` + applyShortDesc + `

### Syntax
`
	paramSetExample = `
# Update the replica count of the 'guestbook' component to 4.
ks param set guestbook replicas 4

# Update the replica count of the 'guestbook' component to 2, but only for the
# 'dev' environment
ks param set guestbook replicas 2 --env=dev`
)

func newParamSetCmd() *cobra.Command {
	paramSetCmd := &cobra.Command{
		Use:     "set <component-name> <param-key> <param-value>",
		Short:   paramShortDesc["set"],
		Long:    paramSetLong,
		Example: paramSetExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			var path string
			var value string

			switch len(args) {
			default:
				return errors.New("invalid arguments for 'param set'")
			case 3:
				name = args[0]
				path = args[1]
				value = args[2]
			case 2:
				path = args[0]
				value = args[1]
			}

			m := map[string]interface{}{
				actions.OptionName:         name,
				actions.OptionPath:         path,
				actions.OptionValue:        value,
				actions.OptionEnvName:      viper.GetString(vParamSetEnv),
				actions.OptionAsString:     viper.GetBool(vParamSetAsString),
				actions.OptionResolveImage: viper.GetBool(vParamSetResolveImage),
			}

			return runAction(actionParamSet, m)
		},
	}

	paramSetCmd.Flags().String(flagEnv, "", "Specify environment to set parameters for")
	viper.BindPFlag(vParamSetEnv, paramSetCmd.Flags().Lookup(flagEnv))

	paramSetCmd.Flags().Bool(flagAsString, false, "Force value to be interpreted as string")
	viper.BindPFlag(vParamSetAsString, paramSetCmd.Flags().Lookup(flagAsString))

	paramSetCmd.Flags().Bool(flagResolveImage, false, "Resolve Docker image tag to reference")
	viper.BindPFlag(vParamSetResolveImage, paramSetCmd.Flags().Lookup(flagResolveImage))

	return paramSetCmd
}
