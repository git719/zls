// helper.go

package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

func DumpCredentials() {
	// Dump credentials file
	f := filepath.Join(confdir, "credentials.yaml")
	if FileExist(f) {
		creds := LoadFileYAML(f)
		tenant_id := StrVal(creds["tenant_id"])
		client_id := StrVal(creds["client_id"])
		client_secret := StrVal(creds["client_secret"])
		print("%-14s %s\n", "tenant_id:", tenant_id)
		print("%-14s %s\n", "client_id:", client_id)
		print("%-14s %s\n", "client_secret:", client_secret)
		exit(1)
	}
}

func SetupCredentialsInterativeLogin(tenant_id, client_id string) {
	// Set up credentials file for interactive login
	f := filepath.Join(confdir, "credentials.yaml")
	if ValidUUID(tenant_id) && ValidUUID(client_id) {
		content := sprint("%-14s %s\n%-14s %s\n%-14s %s\n", "tenant_id:", tenant_id, "client_id:", client_id, "interactive:", "true")
		if err := ioutil.WriteFile(f, []byte(content), 0600); err != nil { // Write string to file
			panic(err.Error())
		}
	} else {
		die("Error. TENANT_ID and/or CLIENT_ID are invalid UUIDs.\n")
	}
	print("Updated credentials file: %s\n", f)
}

func SetupCredentialsSecretLogin(tenant_id, client_id, secret string) {
	// Set up credentials file for client_id + secret login
	f := filepath.Join(confdir, "credentials.yaml")
	if ValidUUID(tenant_id) && ValidUUID(client_id) {
		content := sprint("%-14s %s\n%-14s %s\n%-14s %s\n", "tenant_id:", tenant_id, "client_id:", client_id, "client_secret:", secret)
		if err := ioutil.WriteFile(f, []byte(content), 0600); err != nil { // Write string to file
			panic(err.Error())
		}
	} else {
		die("Error. TENANT_ID and/or CLIENT_ID are invalid UUIDs.\n")
	}
	print("Updated credentials file: %s\n", f)
}

func ReadCredentials() {
	// Read credentials from file
	f := filepath.Join(confdir, "credentials.yaml")
	if FileExist(f) {
		// Read credentials file and update global variables accordingly
		//creds := LoadFileJSON(f).(map[string]interface{}) // Assert as JSON object
		creds := LoadFileYAML(f)
		if creds == nil {
			die("Unable to load/parse file '%s'\n", f)
		}
		tenant_id = StrVal(creds["tenant_id"])
		if !ValidUUID(tenant_id) {
			die("tenant_id '%s' in '%s' is not a valid UUID\n", tenant_id, f)
		}
		client_id = StrVal(creds["client_id"])
		if !ValidUUID(client_id) {
			die("client_id '%s' in '%s' is not a valid UUID\n", client_id, f)
		}
		interactive = strings.ToLower(StrVal(creds["interactive"]))
		if interactive != "true" {
			client_secret = StrVal(creds["client_secret"])
			if client_secret == "" {
				die("client_secret in '%s' is blank\n", f)
			}
		}
	} else {
		die("Missing credentials file: '%s'\n", f +
			"Please rerun program using '-cr' or '-cri' option to specify credentials.\n")
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
