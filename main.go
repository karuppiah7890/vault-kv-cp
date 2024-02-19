package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/hashicorp/vault/api"
)

var usage = `usage: vault-kv-cp <source-kv-mount-path> <destination-kv-mount-path>
`

func main() {

	// Get Config for Source Vault
	sourceConfig := getSourceVaultConfig()

	// Create a new client to the source vault
	sourceDefaultConfig := api.DefaultConfig()
	sourceDefaultConfig.Address = sourceConfig.Address

	if sourceConfig.CACertPath != "" {
		sourceDefaultConfig.ConfigureTLS(&api.TLSConfig{CACert: sourceConfig.CACertPath})
	}
	sourceClient, err := api.NewClient(sourceDefaultConfig)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating source vault client: %s\n", err)
		os.Exit(1)
	}

	// Set the token for the source vault client
	sourceClient.SetToken(sourceConfig.Token)

	// Get Config for Destination Vault
	destinationConfig := getDestinationVaultConfig()

	// Create a new client to the destination vault
	destinationDefaultConfig := api.DefaultConfig()
	destinationDefaultConfig.Address = destinationConfig.Address
	if destinationConfig.CACertPath != "" {
		destinationDefaultConfig.ConfigureTLS(&api.TLSConfig{CACert: destinationConfig.CACertPath})
	}
	destinationClient, err := api.NewClient(destinationDefaultConfig)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating destination vault client: %s\n", err)
		os.Exit(1)
	}

	// Set the token for the destination vault client
	destinationClient.SetToken(destinationConfig.Token)

	// Get the path to the secrets in the source vault.
	// Get the path to the secrets in the destination vault
	var sourceMountPath, destinationMountPath string

	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "%s", usage)
		os.Exit(0)
	}
	flag.Parse()

	if flag.NArg() < 2 {
		flag.Usage()
	}

	sourceMountPath = flag.Args()[0]
	destinationMountPath = flag.Args()[1]

	// Get the list of secrets from the source vault and keep it in memory.
	// Create the list of secrets in the destination vault.
	// If the destination vault already has the same path, then it should be overwritten.
	// OR
	// Directly copy the list of secrets from the source Vault to the destination Vault
	// without keeping it in memory.
	// If the destination vault already has the same path, then it should be overwritten.

	walkVaultPath(sourceMountPath, "", destinationMountPath, "", sourceClient, destinationClient)
}

type VaultConfig struct {
	Address    string
	Token      string
	CACertPath string
}

func getSourceVaultConfig() VaultConfig {
	// SOURCE_VAULT_ADDR
	// SOURCE_VAULT_TOKEN
	// SOURCE_VAULT_CACERT
	return getVaultConfig("SOURCE_")
}

func getDestinationVaultConfig() VaultConfig {
	// DESTINATION_VAULT_ADDR
	// DESTINATION_VAULT_TOKEN
	// DESTINATION_VAULT_CACERT
	return getVaultConfig("DESTINATION_")
}

func getVaultConfig(envPrefix string) VaultConfig {
	config := VaultConfig{}

	// <envPrefix>VAULT_ADDR
	if address := os.Getenv(envPrefix + api.EnvVaultAddress); address != "" {
		config.Address = address
	}

	// <envPrefix>VAULT_TOKEN
	if token := os.Getenv(envPrefix + api.EnvVaultToken); token != "" {
		config.Token = token
	}

	// <envPrefix>VAULT_CACERT
	if caCertPath := os.Getenv(envPrefix + api.EnvVaultCACert); caCertPath != "" {
		config.CACertPath = caCertPath
	}

	return config
}

func walkVaultPath(sourceMounthPath, sourcePath, destinationMountPath, destinationPath string, sourceClient, destinationClient *api.Client) {
	sourceLogicalClient := sourceClient.Logical()

	sourceListPath := path.Join(sourceMounthPath, "metadata", sourcePath)

	sourceKVv2Secrets, err := sourceLogicalClient.List(sourceListPath)
	if err != nil {
		log.Fatalf("error occurred while listing metadata at path %s in source: %v", sourceListPath, err)
	}

	if sourceKVv2Secrets == nil {
		fmt.Fprintf(os.Stdout, "copying secret at `%s` in source to `%s` in destination\n\n", sourcePath, destinationMountPath)
		copySecrets(sourceMounthPath, sourcePath, destinationMountPath, destinationPath, sourceClient, destinationClient)
		return
	}

	fmt.Fprintf(os.Stdout, "%+v\n\n", sourceKVv2Secrets)

	data := sourceKVv2Secrets.Data
	if data == nil {
		log.Fatalf("no data found at path `%s` in source", sourceListPath)
	}

	keys, ok := data["keys"]
	if !ok {
		log.Fatalf("no data found at path `%s` in source", sourceListPath)
	}

	// TODO: `keys` can be of non-array type too. So, type assertion is required.
	// No problems if `keys` is of array type.
	// If `keys` is not an array, then it will panic. So, handle this issue.
	for _, key := range keys.([]interface{}) {
		newSourcePath := path.Join(sourcePath, key.(string))
		newDestinationPath := path.Join(destinationPath, key.(string))
		walkVaultPath(sourceMounthPath, newSourcePath, destinationMountPath, newDestinationPath, sourceClient, destinationClient)
	}
}

func copySecrets(sourceMounthPath, sourcePath, destinationMountPath, destinationPath string, sourceClient, destinationClient *api.Client) {
	sourceKVv2Client := sourceClient.KVv2(sourceMounthPath)
	kvSecret, err := sourceKVv2Client.Get(context.TODO(), sourcePath)

	if err != nil {
		log.Fatalf("error occurred while getting latest version of the secret at path `%s` in source: %v", sourcePath, err)
	}

	if kvSecret == nil {
		log.Fatalf("no secret found at path `%s` in source", sourcePath)
	}

	destinationKVv2Client := destinationClient.KVv2(destinationMountPath)

	finalKvSecret, err := destinationKVv2Client.Put(context.TODO(), destinationPath, kvSecret.Data)

	if err != nil {
		log.Fatalf("error occurred while putting/writing the secret at path `%s` in destination: %v", destinationPath, err)
	}

	if finalKvSecret == nil {
		log.Fatalf("no secret at path `%s` in destination", destinationPath)
	}
}
