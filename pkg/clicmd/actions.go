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
	"github.com/spf13/viper"
)

type initName int

const (
	actionApply initName = iota
	actionComponentList
	actionComponentRm
	actionDelete
	actionDiff
	actionEnvAdd
	actionEnvCurrent
	actionEnvDescribe
	actionEnvList
	actionEnvRm
	actionEnvSet
	actionEnvTargets
	actionEnvUpdate
	actionImport
	actionInit
	actionModuleCreate
	actionModuleList
	actionParamDelete
	actionParamDiff
	actionParamList
	actionParamSet
	actionParamUnset
	actionPkgDescribe
	actionPkgInstall
	actionPkgList
	actionPkgRemove
	actionPrototypeDescribe
	actionPrototypeList
	actionPrototypePreview
	actionPrototypeSearch
	actionPrototypeUse
	actionRegistryAdd
	actionRegistryDescribe
	actionRegistryList
	actionRegistrySet
	actionShow
	actionUpgrade
	actionValidate
)

type actionFn func(map[string]interface{}) error

var (
	actionFns = map[initName]actionFn{
		actionApply:             actions.RunApply,
		actionComponentList:     actions.RunComponentList,
		actionComponentRm:       actions.RunComponentRm,
		actionDelete:            actions.RunDelete,
		actionDiff:              actions.RunDiff,
		actionEnvAdd:            actions.RunEnvAdd,
		actionEnvCurrent:        actions.RunEnvCurrent,
		actionEnvDescribe:       actions.RunEnvDescribe,
		actionEnvList:           actions.RunEnvList,
		actionEnvRm:             actions.RunEnvRm,
		actionEnvSet:            actions.RunEnvSet,
		actionEnvTargets:        actions.RunEnvTargets,
		actionEnvUpdate:         actions.RunEnvUpdate,
		actionImport:            actions.RunImport,
		actionInit:              actions.RunInit,
		actionModuleCreate:      actions.RunModuleCreate,
		actionModuleList:        actions.RunModuleList,
		actionParamDiff:         actions.RunParamDiff,
		actionParamDelete:       actions.RunParamDelete,
		actionParamUnset:        actions.RunParamDelete,
		actionParamList:         actions.RunParamList,
		actionParamSet:          actions.RunParamSet,
		actionPkgDescribe:       actions.RunPkgDescribe,
		actionPkgInstall:        actions.RunPkgInstall,
		actionPkgList:           actions.RunPkgList,
		actionPkgRemove:         actions.RunPkgRemove,
		actionPrototypeDescribe: actions.RunPrototypeDescribe,
		actionPrototypeList:     actions.RunPrototypeList,
		actionPrototypePreview:  actions.RunPrototypePreview,
		actionPrototypeSearch:   actions.RunPrototypeSearch,
		actionPrototypeUse:      actions.RunPrototypeUse,
		actionRegistryAdd:       actions.RunRegistryAdd,
		actionRegistryDescribe:  actions.RunRegistryDescribe,
		actionRegistryList:      actions.RunRegistryList,
		actionRegistrySet:       actions.RunRegistrySet,
		actionShow:              actions.RunShow,
		actionUpgrade:           actions.RunUpgrade,
		actionValidate:          actions.RunValidate,
	}
)

func runAction(name initName, args map[string]interface{}) error {
	fn, ok := actionFns[name]
	if !ok {
		return errors.Errorf("invalid action %q", name)
	}

	return fn(args)
}

func addGlobalFlags(m map[string]interface{}) {
	m[actions.OptionTLSSkipVerify] = viper.GetBool(flagTLSSkipVerify)
}
