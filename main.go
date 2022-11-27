// main.go

package main

import (
	"os"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	// Global constants
	prgname = "zls"
	prgver  = "171"
	mg_url  = "https://graph.microsoft.com"
	az_url  = "https://management.azure.com"
)

var (
	// Global variables
	confdir       = ""
	tenant_id     = ""
	client_id     = ""
	client_secret = ""
	interactive   = ""
	authority_url = ""
	mg_token      = ""
	mg_headers    map[string]string
	az_token      = ""
	az_headers    map[string]string
	// String map for each ARM and MG object type to help generesize many of the functions
	oMap = map[string]string{
		"d":  "roleDefinitions",
		"a":  "roleAssignments",
		"s":  "subscriptions",
		"m":  "managementGroups",
		"u":  "users",
		"g":  "groups",
		"sp": "servicePrincipals",
		"ap": "applications",
		"ra": "directoryRoles",
		"rd": "adRoleDef",          // Treated differently in each function. Yeah, this is ugly :-(
	}
)

func PrintUsage() {
	print(prgname + " Azure RBAC and MS Graph listing utility v" + prgver + "\n" +
		"    -Xj                List all X objects in JSON format\n" +
		"    -X                 List all X objects tersely (UUID and most essential attributes)\n" +
		"    -X \"string\"        List all X objects whose name has \"string\" in it\n" +
		"    -Xj UUID|\"string\"  List specific X or matching objects in JSON format\n" +
		"    -X UUID            List specific X object in YAML-like human-readable format\n" +
		"    -X <specfile>      Compare X object specfile to what's in Azure\n" +
		"    -Xx                Delete X object cache local file\n" +
		"\n" +
		"    Where 'X' can be any of these object types:\n" +
		"      d  = RBAC Role Definitions   a  = RBAC Role Assignments   s  = Azure Subscriptions  \n" +
		"      m  = Management Groups       u  = Azure AD Users          g  = Azure AD Groups      \n" +
		"      sp = Service Principals      ap = Applications            ra = Azure AD Roles Active\n" +
		"      rd = Azure AD Roles Defs\n" +
		"\n" +
		"    -ar                              List all role assignments with resolved names\n" +
		"    -mt                              List Management Group and subscriptions tree\n" +
		"    -pags                            List all Azure AD Privileged Access Groups\n" +
		"    -cr                              Dump values in credentials file\n" +
		"    -cr  TENANT_ID CLIENT_ID SECRET  Set up secret login\n" +
		"    -cri TENANT_ID CLIENT_ID         Set up interactive login (NOT WORKING)\n" +
		"    -st                              List local cache count and Azure count of all objects\n" +
		"    -tx                              Delete accessTokens cache file\n" +
		"    -xx                              Delete ALL cache local file\n" +
		"    -v                               Print this usage page\n")
	exit(0)
}

func exit(code int) {
	os.Exit(code) // Syntactic sugar. 'exit' is much more beautiful to type and read than 'os.Exit' :-)
}

func print(format string, args ...interface{}) (n int, err error) {
	return fmt.Printf(format, args...) // More syntactic sugar
}

func sprint(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)	// More syntactic sugar
}

func main() {
	// TestFunction() // DEBUG

	numberOfArguments := len(os.Args[1:]) // Not including the program itself
	if numberOfArguments < 1 || numberOfArguments > 4 {
		// Don't accept less than 1 or more than 4 arguments
		PrintUsage()
	}

	// Set up program configuration directory
	confdir = filepath.Join(os.Getenv("HOME"), "."+prgname)
	if FileNotExist(confdir) {
		if err := os.Mkdir(confdir, 0700); err != nil {
			panic(err.Error())
		}
	}

	// Process given arguments
	switch numberOfArguments {
	case 1:
		arg1 := strings.ToLower(os.Args[1]) // Always treat 1st argument as Lowercase, to ease comparisons

		ReadCredentials() // Set up tenant ID and credentials

		// First process these simple requests that don't need API tokens
		switch arg1 {
		case "-v":
			PrintUsage()
		case "-tx", "-dx", "-ax", "-sx", "-mx", "-ux", "-gx", "-spx", "-apx", "-rax", "-rdx":
			t := arg1[1 : len(arg1)-1] // Single out the object type
			RemoveCacheFile(t)         // Chop off the 1st 2 characters, to leverage oMap
		case "-xx":
			RemoveCacheFile("all")
		case "-cr":
			DumpCredentials()
		}

		SetupTokens() // Remaining requests need API tokens

		// TestFunction() // DEBUG

		// Handle the three(3) primary single-argument list functions for all object types
		switch arg1 {
		case "-st":
			PrintCountStatus()
		case "-dj", "-aj", "-sj", "-mj", "-uj", "-gj", "-spj", "-apj", "-raj", "-rdj": // Handle JSON-printing of all objects
			t := arg1[1 : len(arg1)-1] // Single out the object type
			PrintAllJSON(t)
		case "-d", "-a", "-s", "-m", "-u", "-g", "-sp", "-ap", "-ra", "-rd": // Handle tersely printing for all objects
			t := arg1[1:] // Single out the object type
			PrintAllTersely(t)
		case "-ar":
			PrintRoleAssignmentReport()
		case "-mt":
			PrintManagementGroupTree()
		case "-pags":
			PrintPAGs()
		case "-z":
			DumpVariables()
		default:
			print("No such option.\n")
		}

	case 2:
		arg1 := strings.ToLower(os.Args[1])
		arg2 := os.Args[2]

		ReadCredentials()
		SetupTokens() // Remaining requests need API tokens

		switch arg1 {
		case "-dj", "-aj", "-sj", "-mj", "-uj", "-gj", "-spj", "-apj", "-raj", "-rdj":
			t := arg1[1 : len(arg1)-1] // Single out our object type letter (see oMap)
			if ValidUUID(arg2) {
				x := GetObjectById(t, arg2) // Get single object by ID and print in JSON format
				PrintJSON(x)
			} else {
				oList := GetMatching(t, arg2) // Get all matching objects
				if len(oList) > 1 {           // Print all matching objects as JSON
					PrintJSON(oList)
				} else if len(oList) > 0 { // Print single matching object as JSON
					x := oList[0].(map[string]interface{})
					PrintJSON(x)
				}
			}
		case "-d", "-a", "-s", "-m", "-u", "-g", "-sp", "-ap", "-ra", "-rd":
			t := arg1[1:] // Single out the object type
			if ValidUUID(arg2) {
				x := GetObjectById(t, arg2) // Get single object by ID
				PrintObject(t, x)           // Print in YAML-like format
			} else if FileExist(arg2) && FileSize(arg2) > 0 {
				CompareSpecfile(t, arg2) // Compare specfile to what's in Azure
			} else {
				oList := GetMatching(t, arg2) // Get all matching objects
				if len(oList) > 1 {           // Print all matching objects tersely
					for _, i := range oList {
						x := i.(map[string]interface{})
						PrintTersely(t, x)
					}
				} else if len(oList) > 0 { // Print single matching object in YAML-like format
					x := oList[0].(map[string]interface{})
					PrintObject(t, x)
				}
			}
		default:
			print("No such option.\n")
		}

	case 3:
		arg1 := strings.ToLower(os.Args[1])
		arg2 := os.Args[2]
		arg3 := os.Args[3]

		switch arg1 {
		case "-cri":
			SetupCredentialsInterativeLogin(arg2, arg3)
		default:
			print("No such option.\n")
		}

	case 4:
		arg1 := strings.ToLower(os.Args[1])
		arg2 := os.Args[2]
		arg3 := os.Args[3]
		arg4 := os.Args[4]

		switch arg1 {
		case "-cr":
			SetupCredentialsSecretLogin(arg2, arg3, arg4)
		default:
			print("No such option.\n")
		}
	}
	os.Exit(0)
}
