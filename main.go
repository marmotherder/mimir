package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/marmotherder/mimir/clients"

	"github.com/gorilla/mux"
	"github.com/jessevdk/go-flags"
)

var opts Options
var sOpts ServerOptions
var release string
var re bool

func main() {
	parseArgs(&opts)

	if opts.IsPod {
		log.Println("Running mimir as a pod in k8s")
	}

	if opts.ServerMode {
		release, re = os.LookupEnv("RELEASE")
		if !re {
			release = "mimir"
		}

		r := mux.NewRouter()
		r.HandleFunc("/hook", hook).Methods(http.MethodPost)
		parseArgs(&sOpts)
		log.Printf("Running server on port: %d\n", sOpts.ServerPort)

		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", sOpts.ServerPort),
			Handler: r,
		}

		c := make(chan os.Signal, 1)

		// The init container for the application when running as webhook will create a csr and
		// a mutatingwebhookconfiguration. This loop keeps the server alive, and tries to run
		// shutdown login when we detect that the core container is being shutdown
		go func() {
			err := srv.ListenAndServeTLS(sOpts.TLSCertPath, sOpts.TLSKeyPath)
			log.Println(err.Error())
			c <- os.Interrupt
		}()

		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c

		shutdownServer(srv)
	} else {
		smc, mgr, err := loadClient()
		if err != nil {
			log.Fatalln(err.Error())
		}
		run(opts, smc, mgr)
	}
}

// parseArgs parses the cli flags, allowing a common point to parse later downstream options when
// building config from multiple structs
func parseArgs(opts interface{}) {
	parser := flags.NewParser(opts, flags.IgnoreUnknown)
	_, err := parser.ParseArgs(os.Args[1:])
	if err != nil {
		log.Fatalln(err.Error())
	}
}

// loadHashiCorpVaultClient loads a valid client for loading secrets from Hashicorp Vault
func loadHashiCorpVaultClient(opts Options, hvOpts HashiCorpVaultOptions, auth clients.HashicorpVaultAuth) (smc clients.SecretsManagerClient, mgr clients.SecretsManager) {
	client, err := clients.NewHashicorpVaultClient(hvOpts.Path, hvOpts.URL, hvOpts.Mount, auth)
	if err != nil {
		log.Fatalln(err.Error())
	}
	return client, clients.HashicorpVault
}

// loadAWSClient loads a valid client for loading secrets from AWS Secrets Manager
func loadAWSClient(opts Options, awsOpts AWSOptions, auth clients.AWSSecretsAuth) (smc clients.SecretsManagerClient, mgr clients.SecretsManager) {
	auth.SetRegion(awsOpts.Region)
	client, err := clients.NewAWSSecretsClient(auth)
	if err != nil {
		log.Fatalln(err.Error())
	}
	return client, clients.AWS
}

// run performs a run of mimir secret syncing for the given backend
func run(opts Options, smc clients.SecretsManagerClient, mgr clients.SecretsManager) {
	kc, err := clients.NewK8SClient(opts.IsPod, opts.KubeconfigPath)
	if err != nil {
		log.Fatalln(err.Error())
	}
	namespaces, err := clients.GetNamespaces(kc)
	if err != nil {
		log.Fatalln(err.Error())
	}
	secrets, err := smc.GetSecrets(namespaces...)
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = clients.ManageSecrets(kc, mgr, secrets...)
	if err != nil {
		log.Fatalln(err.Error())
	}
}
