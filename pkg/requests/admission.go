package requests

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

type admitFunc func(v1.AdmissionRequest) *v1.AdmissionResponse

// ErrorToAdmissionResponse converts an error into the appropriate response type
func ErrorToAdmissionResponse(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Result: &metav1.Status{
			Status:  "Failure",
			Message: err.Error(),
		},
	}
}

// RegisterAdmission registers a new http handler for admission requests
func RegisterAdmission(path string, fn admitFunc) {
	http.HandleFunc(path, handler(fn))
}

func handler(fn admitFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			klog.Errorf("contentType=%s, expect application/json", contentType)
			return
		}

		var body []byte
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}

		req := v1.AdmissionReview{}
		resp := v1.AdmissionReview{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AdmissionReview",
				APIVersion: "admission.k8s.io/v1",
			},
		}
		if _, _, err := codecs.UniversalDeserializer().Decode(body, nil, &req); err != nil {
			klog.Error(err)
			resp.Response = ErrorToAdmissionResponse(err)
		} else {
			resp.Response = fn(*req.Request)
		}

		// The UID have to match
		resp.Response.UID = req.Request.UID

		if resp.Response.Result.Details != nil {
			resp.Response.Result.Details.Name = req.Request.Name
			resp.Response.Result.Details.Kind = req.Request.Kind.Kind
			resp.Response.Result.Details.Group = req.Request.Kind.Group
		}

		if !resp.Response.Allowed {
			klog.Warning("Validation response: ", resp.Response)
		} else {
			klog.Info("Validation response: ", resp.Response)
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			klog.Error(err)
		}

		if _, err := w.Write(respBytes); err != nil {
			klog.Error(err)
		}
	}
}
