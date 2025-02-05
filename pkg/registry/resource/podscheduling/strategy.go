/*
Copyright 2022 The Kubernetes Authors.

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

package podscheduling

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/apis/resource"
	"k8s.io/kubernetes/pkg/apis/resource/validation"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

// podSchedulingStrategy implements behavior for PodScheduling objects
type podSchedulingStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating
// ResourceClaim objects via the REST API.
var Strategy = podSchedulingStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

func (podSchedulingStrategy) NamespaceScoped() bool {
	return true
}

// GetResetFields returns the set of fields that get reset by the strategy and
// should not be modified by the user. For a new PodScheduling that is the
// status.
func (podSchedulingStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	fields := map[fieldpath.APIVersion]*fieldpath.Set{
		"resource.k8s.io/v1alpha2": fieldpath.NewSet(
			fieldpath.MakePathOrDie("status"),
		),
	}

	return fields
}

func (podSchedulingStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	scheduling := obj.(*resource.PodScheduling)
	// Status must not be set by user on create.
	scheduling.Status = resource.PodSchedulingStatus{}
}

func (podSchedulingStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	scheduling := obj.(*resource.PodScheduling)
	return validation.ValidatePodScheduling(scheduling)
}

func (podSchedulingStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (podSchedulingStrategy) Canonicalize(obj runtime.Object) {
}

func (podSchedulingStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (podSchedulingStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newScheduling := obj.(*resource.PodScheduling)
	oldScheduling := old.(*resource.PodScheduling)
	newScheduling.Status = oldScheduling.Status
}

func (podSchedulingStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newScheduling := obj.(*resource.PodScheduling)
	oldScheduling := old.(*resource.PodScheduling)
	errorList := validation.ValidatePodScheduling(newScheduling)
	return append(errorList, validation.ValidatePodSchedulingUpdate(newScheduling, oldScheduling)...)
}

func (podSchedulingStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

func (podSchedulingStrategy) AllowUnconditionalUpdate() bool {
	return true
}

type podSchedulingStatusStrategy struct {
	podSchedulingStrategy
}

var StatusStrategy = podSchedulingStatusStrategy{Strategy}

// GetResetFields returns the set of fields that get reset by the strategy and
// should not be modified by the user. For a status update that is the spec.
func (podSchedulingStatusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	fields := map[fieldpath.APIVersion]*fieldpath.Set{
		"resource.k8s.io/v1alpha2": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}

	return fields
}

func (podSchedulingStatusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newScheduling := obj.(*resource.PodScheduling)
	oldScheduling := old.(*resource.PodScheduling)
	newScheduling.Spec = oldScheduling.Spec
}

func (podSchedulingStatusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newScheduling := obj.(*resource.PodScheduling)
	oldScheduling := old.(*resource.PodScheduling)
	return validation.ValidatePodSchedulingStatusUpdate(newScheduling, oldScheduling)
}

// WarningsOnUpdate returns warnings for the given update.
func (podSchedulingStatusStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// Match returns a generic matcher for a given label and field selector.
func Match(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	scheduling, ok := obj.(*resource.PodScheduling)
	if !ok {
		return nil, nil, errors.New("not a PodScheduling")
	}
	return labels.Set(scheduling.Labels), toSelectableFields(scheduling), nil
}

// toSelectableFields returns a field set that represents the object
func toSelectableFields(scheduling *resource.PodScheduling) fields.Set {
	fields := generic.ObjectMetaFieldsSet(&scheduling.ObjectMeta, true)
	return fields
}
