// helper.go

package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

func DumpVariables() {
	// Dump essential global variables
	print("%-16s %s\n", "tenant_id:", tenant_id)
    if interactive == "true" {
		print("%-16s %s\n", "username:", username)	
		print("%-16s %s\n", "interactive:", "true")	
	} else {
		print("%-16s %s\n", "client_id:", client_id)
		print("%-16s %s\n", "client_secret:", client_secret)	
	}
	print("%-16s %s\n%-16s %s\n%-16s %s\n", "authority_url:", authority_url, "mg_url:", mg_url, "az_url:", az_url)
	print("mg_headers:\n")
	for k, v := range mg_headers {
		print("  %-14s %s\n", StrVal(k) + ":", StrVal(v))
	}
	print("az_headers:\n")
	for k, v := range az_headers {
		print("  %-14s %s\n", StrVal(k) + ":", StrVal(v))
	}
	exit(1)
}

func DumpCredentials() {
	// Dump credentials file
	creds_file := filepath.Join(confdir, "credentials.yaml")
	creds := LoadFileYAML(creds_file)
	print("%-14s %s\n", "tenant_id:", StrVal(creds["tenant_id"]))
	if strings.ToLower(StrVal(creds["interactive"])) == "true" {
		print("%-14s %s\n", "username:", StrVal(creds["username"]))
		print("%-14s %s\n", "interactive:", "true")
	} else {
		print("%-14s %s\n", "client_id:", StrVal(creds["client_id"]))
		print("%-14s %s\n", "client_secret:", StrVal(creds["client_secret"]))
	}
	exit(1)
}

func SetupInterativeLogin(tenant_id, username string) {
	// Set up credentials file for interactive login
	f := filepath.Join(confdir, "credentials.yaml")
	if !ValidUUID(tenant_id) {
		die("Error. TENANT_ID is an invalid UUIs.\n")
	}
	content := sprint("%-14s %s\n%-14s %s\n%-14s %s\n", "tenant_id:", tenant_id, "username:", username, "interactive:", "true")
	if err := ioutil.WriteFile(f, []byte(content), 0600); err != nil { // Write string to file
		panic(err.Error())
	}
	print("[%s] Updated credentials\n", f)
}

func SetupAutomatedLogin(tenant_id, client_id, secret string) {
	// Set up credentials file for client_id + secret login
	f := filepath.Join(confdir, "credentials.yaml")
	if !ValidUUID(tenant_id) {
		die("Error. TENANT_ID is an invalid UUIs.\n")
	}
	if !ValidUUID(client_id) {
		die("Error. CLIENT_ID is an invalid UUIs.\n")
	}
	content := sprint("%-14s %s\n%-14s %s\n%-14s %s\n", "tenant_id:", tenant_id, "client_id:", client_id, "client_secret:", secret)
	if err := ioutil.WriteFile(f, []byte(content), 0600); err != nil { // Write string to file
		panic(err.Error())
	}
	print("[%s] Updated credentials\n", f)
}

func SetupCredentials() {
	// Read credentials file and set up authentication parameters as global variables
	creds_file := filepath.Join(confdir, "credentials.yaml")
	if FileNotExist(creds_file) && FileSize(creds_file) < 1 {
		die("Missing credentials file: '%s'\n", creds_file +
			"Please rerun program using '-cr' or '-cri' option to specify credentials.\n")
	}
	creds := LoadFileYAML(creds_file)

	// Note we're updating global variables
	tenant_id = StrVal(creds["tenant_id"])
	if !ValidUUID(tenant_id) {
		die("[%s] tenant_id '%s' is not a valid UUID\n", creds_file, tenant_id)
	}
	interactive = strings.ToLower(StrVal(creds["interactive"]))
	if interactive == "true" {
		username = strings.ToLower(StrVal(creds["username"]))
	} else {
		client_id = StrVal(creds["client_id"])
		if !ValidUUID(client_id) {
			die("[%s] client_id '%s' is not a valid UUID\n", creds_file, client_id)
		}	
		client_secret = StrVal(creds["client_secret"])
		if client_secret == "" {
			die("[%s] client_secret is blank\n", creds_file)
		}	
	}
}

func SetupApiTokens() {
	// Initialize necessary global variables, acquire all API tokens, and set them up for use
	SetupCredentials()  // Sets up tenant ID, client ID, authentication method, etc
	authority_url = "https://login.microsoftonline.com/" + tenant_id

	// This utility calls 2 different resources, the Azure Resource Management (ARM) and MS Graph APIs, and each needs
	// its own separate token. The Microsoft identity platform does not allow you to get a token for several resources at once.
	// See https://learn.microsoft.com/en-us/azure/active-directory/develop/msal-net-user-gets-consent-for-multiple-resources

	az_scope := []string{az_url + "/.default"}  // The scope is a slice of strings
	// Appending '/.default' will allow using all static and consented permissions of the identity in use
	// See https://learn.microsoft.com/en-us/azure/active-directory/develop/msal-v1-app-scopes
	az_token, _ = GetToken(az_scope)  // Note, these are 2 global variable we are updating!
	az_headers = map[string]string{ "Authorization": "Bearer " + az_token, "Content-Type":  "application/json", }
	
	// Rinse, repeat for MS Graph access	
	mg_scope := []string{mg_url + "/.default"}
	mg_token, _ = GetToken(mg_scope)
	mg_headers = map[string]string{"Authorization": "Bearer " + mg_token, "Content-Type":  "application/json", 	}

	// You would get other API tokens here ...
}
