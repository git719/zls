// helper.go

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func DumpCredentials() {
	// Dump credentials file
	f := filepath.Join(confdir, "credentials.json")
	if FileExist(f) {
		creds := LoadFileJSON(f).(map[string]interface{}) // Assert as JSON object
		//PrintYAML(creds)
		PrintJSON(creds)
		os.Exit(1)
	}
}

func SetupCredentialsInterativeLogin(tenant_id, client_id string) {
	// Set up credentials file for interactive login
	f := filepath.Join(confdir, "credentials.json")
	if ValidUUID(tenant_id) && ValidUUID(client_id) {
		content := fmt.Sprintf("{\n  \"tenant_id\" : \"" + tenant_id + "\",\n  \"client_id\" : \"" + client_id + "\",\n  \"interactive\" : \"true\"\n}\n")
		if err := ioutil.WriteFile(f, []byte(content), 0600); err != nil { // Write string to file
			panic(err.Error())
		}
	} else {
		fmt.Println("Error. TENANT_ID and/or CLIENT_ID are invalid UUIDs.")
		os.Exit(1)
	}
	fmt.Println("Updated credentials file:", f)
}

func SetupCredentialsSecretLogin(tenant_id, client_id, secret string) {
	// Set up credentials file for client_id + secret login
	f := filepath.Join(confdir, "credentials.json")
	if ValidUUID(tenant_id) && ValidUUID(client_id) {
		content := fmt.Sprintf("{\n  \"tenant_id\" : \"" + tenant_id + "\",\n  \"client_id\" : \"" + client_id + "\",\n  \"client_secret\" : \"" + secret + "\"\n}\n")
		if err := ioutil.WriteFile(f, []byte(content), 0600); err != nil { // Write string to file
			panic(err.Error())
		}
	} else {
		fmt.Println("Error. TENANT_ID and/or CLIENT_ID are invalid UUIDs.")
		os.Exit(1)
	}
	fmt.Println("Updated credentials file:", f)
}

func ReadCredentials() {
	// Read credentials from file
	f := filepath.Join(confdir, "credentials.json")
	if FileExist(f) {
		creds := LoadFileJSON(f).(map[string]interface{}) // Assert as JSON object
		// Read credentials file and update global variables accordingly
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
		interactive = strings.ToLower(StrVal(creds["interactive"]))
		if interactive != "true" {
			client_secret = StrVal(creds["client_secret"])
			if client_secret == "" {
				log.Printf("client_secret in \"%s\" is blank\n", f)
				os.Exit(1)
			}
		}
	} else {
		fmt.Printf("Missing credentials file: \"" + f + "\"\n")
		fmt.Println("Please rerun program using '-cr' or '-cri' option to specify credentials.")
		os.Exit(1)
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
