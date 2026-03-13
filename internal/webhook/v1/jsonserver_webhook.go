/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"strings"

	examplev1 "github.com/ntlaletsi70/json-server-controller/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var JsonServerlog = logf.Log.WithName("JsonServer-resource")

// SetupJsonServerWebhookWithManager registers the webhook for JsonServer in the manager.
func SetupJsonServerWebhookWithManager(mgr ctrl.Manager) error {
	var JsonServer examplev1.JsonServer
	return ctrl.NewWebhookManagedBy(mgr, &JsonServer).
		WithValidator(&JsonServerCustomValidator{}).
		Complete()
}

// JsonServerCustomValidator struct handles validation logic using Generics.

// +kubebuilder:webhook:path=/validate-example-com-v1-jsonserver,mutating=false,failurePolicy=fail,sideEffects=None,groups=example.com,resources=jsonservers,verbs=create;update,versions=v1,name=vjsonserver-v1.kb.io,admissionReviewVersions=v1
type JsonServerCustomValidator struct{}

// Interface guard ensuring we implement the Generic Validator interface.
var _ admission.Validator[*examplev1.JsonServer] = &JsonServerCustomValidator{}

// ValidateCreate implements admission.Validator.
func (v *JsonServerCustomValidator) ValidateCreate(
	ctx context.Context,
	obj *examplev1.JsonServer,
) (admission.Warnings, error) {

	JsonServerlog.Info("Validation for JsonServer upon creation", "name", obj.GetName())

	var allErrs field.ErrorList

	if !strings.HasPrefix(obj.Name, "app-") {

		allErrs = append(allErrs,
			field.Invalid(
				field.NewPath("metadata").Child("name"),
				obj.Name,
				"must start with 'app-'",
			),
		)
	}

	if len(allErrs) > 0 {

		return nil, errors.NewInvalid(
			schema.GroupKind{
				Group: "example.com",
				Kind:  "JsonServer",
			},
			obj.Name,
			allErrs,
		)
	}

	return nil, nil
}

// ValidateUpdate implements admission.Validator.
func (v *JsonServerCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *examplev1.JsonServer) (admission.Warnings, error) {
	JsonServerlog.Info("Validation for JsonServer upon update", "name", newObj.GetName())

	var allErrs field.ErrorList

	if !strings.HasPrefix(newObj.Name, "app-") {

		allErrs = append(allErrs,
			field.Invalid(
				field.NewPath("metadata").Child("name"),
				newObj.Name,
				"must start with 'app-'",
			),
		)
	}

	if len(allErrs) > 0 {

		return nil, errors.NewInvalid(
			schema.GroupKind{
				Group: "example.com",
				Kind:  "JsonServer",
			},
			newObj.Name,
			allErrs,
		)
	}

	return nil, nil
}

// ValidateDelete implements admission.Validator.
func (v *JsonServerCustomValidator) ValidateDelete(ctx context.Context, obj *examplev1.JsonServer) (admission.Warnings, error) {
	JsonServerlog.Info("Validation for JsonServer upon deletion", "name", obj.GetName())

	// TODO: Add deletion validation logic here
	return nil, nil
}
