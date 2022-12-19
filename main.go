// main.go

package main

import (
	"os"
	"path/filepath"
)

const (
	// Global constants
	prgname = "zls"
	prgver  = "177"
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
	username      = ""
	authority_url = ""
	mg_token      = ""
	mg_headers    map[string]string
	az_token      = ""
	az_headers    map[string]string
	oMap = map[string]string{	    // String map to help generesize many of the functions
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
		"    -X <specfile>      Compare X object specification file to what's in Azure\n" +
		"    -Xx                Delete X object cache local file\n" +
		"\n" +
		"    Where 'X' can be any of these object types:\n" +
		"      d  = RBAC Role Definitions   a  = RBAC Role Assignments   s  = Azure Subscriptions  \n" +
		"      m  = Management Groups       u  = Azure AD Users          g  = Azure AD Groups      \n" +
		"      sp = Service Principals      ap = Applications            ra = Azure AD Roles Active\n" +
		"      rd = Azure AD Roles Defs\n" +
		"\n" +
		"    -ar                               List all RBAC role assignments with resolved names\n" +
		"    -mt                               List Management Group and subscriptions tree\n" +
		"    -pags                             List all Azure AD Privileged Access Groups\n" +
		"    -st                               List local cache count and Azure count of all objects\n" +
		"\n" +
		"    -z                                Dump variables in running program\n" +
		"    -cr                               Dump values in credentials file\n" +
		"    -cr  TENANT_ID CLIENT_ID SECRET   Set up MSAL automated client_id + secret login\n" +
		"    -cri TENANT_ID USERNAME           Set up MSAL interactive browser popup login\n" +
		"    -tx                               Delete MSAL accessTokens cache file\n" +
		"    -xx                               Delete ALL cache local file\n" +
		"    -v                                Print this usage page\n")
	exit(0)
}

func main() {
	// TestFunction() // DEBUG
	numberOfArguments := len(os.Args[1:]) // Not including the program itself
	if numberOfArguments < 1 || numberOfArguments > 4 {
		PrintUsage()  // Don't accept less than 1 or more than 4 arguments
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
		arg1 := os.Args[1]
		switch arg1 {     // First, process 1-arg requests that don't need credentials and API tokens set up
	    case "-v":
			PrintUsage()
		}
		SetupApiTokens()  // The rest do need global credentials API tokens to be available
		switch arg1 {
		case "-tx", "-dx", "-ax", "-sx", "-mx", "-ux", "-gx", "-spx", "-apx", "-rax", "-rdx":
			t := arg1[1 : len(arg1)-1]  // Single out the object type
			RemoveCacheFile(t)          // Chop off the 1st 2 characters, to leverage oMap
		case "-xx":
			RemoveCacheFile("all")
		case "-cr":
			DumpCredentials()
		case "-st":
			PrintCountStatus()
		case "-dj", "-aj", "-sj", "-mj", "-uj", "-gj", "-spj", "-apj", "-raj", "-rdj": // Handle JSON-printing of all objects
			t := arg1[1 : len(arg1)-1] // Single out the object type
			PrintAllJson(t)
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
			PrintUsage()
		}
	case 2:
		arg1 := os.Args[1]
		arg2 := os.Args[2]
		SetupApiTokens()
		switch arg1 {
		case "-dj", "-aj", "-sj", "-mj", "-uj", "-gj", "-spj", "-apj", "-raj", "-rdj":
			t := arg1[1 : len(arg1)-1] // Single out our object type letter (see oMap)
			if ValidUuid(arg2) {
				x := GetObjectById(t, arg2) // Get single object by ID and print in JSON format
				PrintJson(x)
			} else {
				oList := GetMatching(t, arg2) // Get all matching objects
				if len(oList) > 1 {           // Print all matching objects as JSON
					PrintJson(oList)
				} else if len(oList) > 0 { // Print single matching object as JSON
					x := oList[0].(map[string]interface{})
					PrintJson(x)
				}
			}
		case "-d", "-a", "-s", "-m", "-u", "-g", "-sp", "-ap", "-ra", "-rd":
			t := arg1[1:] // Single out the object type
			if ValidUuid(arg2) {
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
			PrintUsage()
		}
	case 3:
		arg1 := os.Args[1]
		arg2 := os.Args[2]
		arg3 := os.Args[3]
		switch arg1 {
		case "-cri":
			SetupInterativeLogin(arg2, arg3)
		default:
			PrintUsage()
		}
	case 4:
		arg1 := os.Args[1]
		arg2 := os.Args[2]
		arg3 := os.Args[3]
		arg4 := os.Args[4]
		switch arg1 {
		case "-cr":
			SetupAutomatedLogin(arg2, arg3, arg4)
		default:
			PrintUsage()
		}
	}
	exit(0)
}
