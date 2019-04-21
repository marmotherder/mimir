package main

import (
	"log"

	"mimir/clients"
)

func main() {
	kc, err := clients.NewK8SClient(false)
	if err != nil {
		log.Fatalln(err.Error())
	}
	vc, err := clients.NewHashicorpVaultClient("", "", "", clients.HashicorpVaultTokenAuth{Token: ""})
	if err != nil {
		log.Fatalln(err.Error())
	}
	namespaces, err := clients.GetNamespaces(kc)
	if err != nil {
		log.Fatalln(err.Error())
	}
	secrets, err := vc.GetSecrets(namespaces...)
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = clients.ManageSecrets(kc, clients.HashicorpVault, secrets...)
	if err != nil {
		log.Fatalln(err.Error())
	}
}
