package handler

import (
	"encoding/json"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/skos-ninja/k8s-annotation-admission/pkg/annotations"
)

type args = map[string]string

func Test_ValidAnnotation(t *testing.T) {
	assert := assert.New(t)
	viper.Set(annotations.FlagKey, map[string]string{
		"test": ".*",
	})

	req := buildRequest(t, args{
		"test": "valid",
	})
	resp := Handler(req)

	assert.Equal(true, resp.Allowed, "Expected accepted")
}

func Test_InvalidAnnotation(t *testing.T) {
	assert := assert.New(t)
	viper.Set(annotations.FlagKey, map[string]string{
		"test": "w.*",
	})

	req := buildRequest(t, args{
		"test": "valid",
	})
	resp := Handler(req)

	assert.Equal(false, resp.Allowed, "Expected rejection")
}

func Test_MissingAnnotation(t *testing.T) {
	assert := assert.New(t)
	viper.Set(annotations.FlagKey, map[string]string{
		"test": ".*",
	})

	req := buildRequest(t, args{})
	resp := Handler(req)

	assert.Equal(false, resp.Allowed, "Expected rejection")
}

func buildRequest(t *testing.T, arg args) v1.AdmissionRequest {
	c := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"annotations": arg,
			},
		},
	}

	d, err := json.Marshal(c)
	if err != nil {
		t.Error(err)
	}

	return v1.AdmissionRequest{
		UID:       "test",
		Operation: v1.Create,
		Object: runtime.RawExtension{
			Raw: d,
		},
	}
}
