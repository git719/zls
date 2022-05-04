// helper.go

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func SetupCredentials() {
	// Set up tenant credentials
	f := filepath.Join(confdir, "credentials.json")
	if _, err := os.Stat(f); os.IsNotExist(err) {
		fmt.Printf("Missing credentials file: \"" + f + "\"\n")
		content := fmt.Sprintf("{\n  \"tenant_id\" : \"UUID\",\n  \"client_id\" : \"UUID\",\n  \"client_secret\" : \"SECRET\"\n}\n")
		if err = ioutil.WriteFile(f, []byte(content), 0600); err != nil { // Write string to file
			panic(err.Error())
		}
		fmt.Println("Created new skeleton one: Please edit it, fill-in required values, then re-run program.")
		os.Exit(1)
	} else {
		creds := LoadFileJSON(f).(map[string]interface{}) // Assert as JSON object
		if creds == nil {
			log.Printf("Unable to load/parse file \"%s\"\n", f)
			os.Exit(1)
		}
		tenant_id = StrVal(creds["tenant_id"])
		if !ValidUUID(tenant_id) {
			log.Printf("tenant_id \"%s\" in \"%s\" is not a valid UUID\n", tenant_id, f)
			os.Exit(1)
		}
		client_id = StrVal(creds["client_id"])
		if !ValidUUID(client_id) {
			log.Printf("client_id \"%s\" in \"%s\" is not a valid UUID\n", client_id, f)
			os.Exit(1)
		}
		client_secret = StrVal(creds["client_secret"])
		if client_secret == "" {
			log.Printf("client_secret in \"%s\" is blank\n", f)
			os.Exit(1)
		}
	}
}

func SetupTokens() {
	// Initialize global variables and grab tokens for each API
	authority_url = "https://login.microsoftonline.com/" + tenant_id

	// For Azure Resource Management (ARM) API calls
	// Scope '/.default' uses whatever static permissions are defined for the SP being used
	az_scope := []string{az_url + "/.default"}
	az_token, _ = GetToken(az_scope)
	az_headers = map[string]string{ // Default headers for ARM calls
		"Authorization": "Bearer " + az_token,
		"Content-Type":  "application/json",
	}

	// For MS Graph (MG) API calls
	mg_scope := []string{mg_url + "/.default"}
	mg_token, _ = GetToken(mg_scope)
	mg_headers = map[string]string{ // Default headers for MG calls
		"Authorization": "Bearer " + mg_token,
		"Content-Type":  "application/json",
	}
}
