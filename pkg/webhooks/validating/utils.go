/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2019 Red Hat, Inc.
 */

package validating

import (
	"fmt"

	templatev1 "github.com/openshift/api/template/v1"

	k6tv1 "kubevirt.io/kubevirt/pkg/api/v1"

	"github.com/fromanirh/kubevirt-template-validator/pkg/virtinformers"

	"github.com/fromanirh/kubevirt-template-validator/internal/pkg/log"
)

const (
	annotationTemplateNameKey      string = "vm.cnv.io/template"
	annotationTemplateNamespaceKey string = "vm.cnv.io/template-namespace"
)

func getTemplateKey(vm *k6tv1.VirtualMachine) (string, bool) {
	if vm.Annotations == nil {
		log.Log.Warningf("VM %s missing annotations entirely", vm.Name)
		return "", false
	}

	templateNamespace := vm.Annotations[annotationTemplateNamespaceKey]
	if templateNamespace == "" {
		log.Log.Warningf("VM %s missing template namespace annotation", vm.Name)
		return "", false
	}

	templateName := vm.Annotations[annotationTemplateNameKey]
	if templateNamespace == "" {
		log.Log.Warningf("VM %s missing template annotation", vm.Name)
		return "", false
	}

	return fmt.Sprintf("%s/%s", templateNamespace, templateName), true
}

func getParentTemplateForVM(vm *k6tv1.VirtualMachine) (*templatev1.Template, error) {
	informers := virtinformers.GetInformers()

	if informers == nil || informers.TemplateInformer == nil {
		// no error, it can happen: we're been deployed ok K8S, not OKD/OCD.
		return nil, nil
	}

	cacheKey, ok := getTemplateKey(vm)
	if !ok {
		// baked VM (aka no parent template). May happen, it's OK.
		return nil, nil
	}

	obj, exists, err := informers.TemplateInformer.GetStore().GetByKey(cacheKey)
	if err != nil {
		return nil, err
	}

	if !exists {
		// ok, this is weird
		return nil, fmt.Errorf("unable to find template object %s for VM %s", cacheKey, vm.Name)
	}

	tmpl := obj.(*templatev1.Template)
	// TODO explain deepcopy
	return tmpl.DeepCopy(), nil
}