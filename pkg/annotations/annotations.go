package annotations

import (
	"fmt"
	"regexp"

	"github.com/spf13/viper"
	"k8s.io/klog/v2"
)

const (
	// FlagKey is the key of the flag for the annotations map
	FlagKey = "annotations"
	// FlagWarning is the key of the flag for the validator to run in warning only mode
	FlagWarning = "warning"
)

var (
	annotationsCache = make(map[string]*regexp.Regexp)
	warningMode      = false
)

func getExpr(name string) *regexp.Regexp {
	value, ok := annotationsCache[name]
	if !ok {
		return nil
	}

	return value
}

// InitValidations performs a regex compil check on all annotations
func InitValidations() {
	warningMode = viper.GetBool(FlagWarning)
	annotations := viper.GetStringMapString(FlagKey)
	klog.Infof("Validating %d annotations...\n", len(annotations))
	for k, v := range annotations {
		klog.Infof("Annotation: %s\n", k)
		// Will force a crash on a failed compile of a regex
		annotationsCache[k] = regexp.MustCompile(v)
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
	keys := make([]string, 0, len(annotationsCache))
	for k := range annotationsCache {
		keys = append(keys, k)
	}

	return keys
}

// IsWarningMode returns if the validators are set to only run in warning mode
func IsWarningMode() bool {
	return warningMode
}
