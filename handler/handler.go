package handler

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/skos-ninja/k8s-annotation-admission/pkg/annotations"
	"github.com/skos-ninja/k8s-annotation-admission/pkg/requests"
)

// Handler will validate the annotations on the resource
func Handler(req v1.AdmissionRequest) (resp *v1.AdmissionResponse) {
	resp = &v1.AdmissionResponse{
		Allowed: true,
		Result:  &metav1.Status{},
	}

	// Ignore any requests that aren't a create or update
	if req.Operation != v1.Create && req.Operation != v1.Update {
		return
	}

	// We convert to an unstructured as we don't care about the type here
	meta, err := convertToUnstructured(&req)
	if err != nil {
		return requests.ErrorToAdmissionResponse(err)
	}

	metaAnnotations := meta.GetAnnotations()
	annotationKeys := annotations.GetAnnotationKeys()

	// Validate each annotation we expect
	for _, k := range annotationKeys {
		v := metaAnnotations[k]
		if strings.TrimSpace(v) == "" {
			addFailure(resp, k, metav1.CauseTypeFieldValueRequired, "missing annotation")
		} else {
			err := annotations.Validate(k, v)
			if err != nil {
				// If the annotation is invalid then add an annotation for that annotation stating why
				addFailure(resp, k, metav1.CauseTypeFieldValueInvalid, err.Error())
			}
		}

	}

	// If we fail validation then ensure we return a human readable response
	if !resp.Allowed {
		resp.Result.Message = "Failed to validate annotations"
		resp.Result.Details.UID = meta.GetUID()
	}

	return
}

func convertToUnstructured(req *v1.AdmissionRequest) (*unstructured.Unstructured, error) {
	var obj runtime.Object
	var scope conversion.Scope
	err := runtime.Convert_runtime_RawExtension_To_runtime_Object(&req.Object, &obj, scope)
	if err != nil {
		return nil, err
	}

	innerObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: innerObj}, nil
}

func addFailure(resp *v1.AdmissionResponse, key string, causeType metav1.CauseType, message string) {
	resp.Allowed = false
	if resp.Result.Details == nil {
		resp.Result.Details = &metav1.StatusDetails{}
	}

	cause := metav1.StatusCause{
		Type:    causeType,
		Message: message,
		Field:   fmt.Sprintf("items[0].annotations.%s", key),
	}

	resp.Result.Details.Causes = append(resp.Result.Details.Causes, cause)
}
