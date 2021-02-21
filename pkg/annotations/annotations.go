package annotations

import (
	"fmt"
	"regexp"

	"github.com/spf13/viper"
	"k8s.io/klog"
)

const FlagKey = "annotations"

func getAnnotations() map[string]string {
	return viper.GetStringMapString(FlagKey)
}

func getExpr(name string) *regexp.Regexp {
	value, ok := getAnnotations()[name]
	if !ok {
		return nil
	}

	// Currently we crash if the regex is invalid.
	return regexp.MustCompile(value)
}

// InitValidations performs a regex compil check on all annotations
func InitValidations() {
	annotations := getAnnotations()
	klog.Infof("Validating %d annotations...\n", len(annotations))
	for k := range annotations {
		klog.Infof("Annotation: %s\n", k)
		// Will force a crash on a failed compile of a regex
		getExpr(k)
	}

}

// Validate performs a validation of the annotation value against the regex
func Validate(name, value string) error {
	if expr := getExpr(name); expr != nil {
		if valid := expr.MatchString(value); !valid {
			return fmt.Errorf("%s does not match regex expression %s", name, expr.String())
		}
	}

	return nil
}

// GetAnnotationKeys returns all the names of the annotations that we expect to validate
func GetAnnotationKeys() []string {
	annotations := getAnnotations()
	keys := make([]string, 0, len(annotations))
	for k := range annotations {
		keys = append(keys, k)
	}

	return keys
}
