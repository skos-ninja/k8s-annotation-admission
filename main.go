package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"

	"github.com/skos-ninja/k8s-annotation-admission/handler"
	"github.com/skos-ninja/k8s-annotation-admission/pkg/annotations"
	"github.com/skos-ninja/k8s-annotation-admission/pkg/requests"
)

var cmd = &cobra.Command{
	Use:   "k8s-annotation-admission",
	Short: "",
	Long:  "",
	RunE:  runE,
	PostRun: func(cmd *cobra.Command, args []string) {
		klog.Flush()
	},
}

func init() {
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	cmd.Flags().IntP("port", "p", 8080, "Specify port to run server on")
	viper.BindPFlag("port", cmd.Flags().Lookup("port"))

	cmd.Flags().StringP("tls-cert", "c", "", "Specify TLS certificate path")
	viper.BindPFlag("tls-cert", cmd.Flags().Lookup("tls-cert"))

	cmd.Flags().StringP("tls-key", "k", "", "Specify TLS key path")
	viper.BindPFlag("tls-key", cmd.Flags().Lookup("tls-key"))

	cmd.Flags().BoolP(annotations.FlagWarning, "w", false, "Only warn on a failed validation")
	viper.BindPFlag(annotations.FlagWarning, cmd.Flags().Lookup(annotations.FlagWarning))

	cmd.Flags().StringP(annotations.FlagKey, "a", "{}", "Specify annotations")
	viper.BindPFlag(annotations.FlagKey, cmd.Flags().Lookup(annotations.FlagKey))

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func runE(cmd *cobra.Command, args []string) error {
	requests.RegisterAdmission("/validate", handler.Handler)

	annotations.InitValidations()

	port := viper.GetInt("port")
	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	tlsCert := viper.GetString("tls-cert")
	tlsKey := viper.GetString("tls-key")
	if tlsCert != "" && tlsKey != "" {
		sCert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
		if err != nil {
			return err
		}
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{sCert},
		}

		klog.Info("Listening on tls :", port)
		return server.ListenAndServeTLS("", "")
	}

	klog.Info("Listening on :", port)
	return server.ListenAndServe()
}
