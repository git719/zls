// main.go

package main

import (
	"os"
	"path/filepath"
)

const (
	// Global constants
	prgname = "zls"
	prgver  = "184"
	mg_url  = "https://graph.microsoft.com"
	az_url  = "https://management.azure.com"
	rUp     = "\x1B[2K\r"
		// See https://stackoverflow.com/questions/1508490/erase-the-current-printed-console-line
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
		"    -vs SPECFILE         Compare YAML or JSON specfile to what's in Azure (only for d and a objects)\n" +
		"    -X[j]                List all X objects tersely, with JSON output option\n" +
		"    -X[j] UUID|\"string\"  Show/list X object(s) matching on UUID or \"string\" attribute, JSON option\n" +
		"    -Xx                  Delete X object local file cache\n" +
		"\n" +
		"    Where 'X' can be any of these object types:\n" +
		"      d  = RBAC Role Definitions   a  = RBAC Role Assignments   s  = Azure Subscriptions  \n" +
		"      m  = Management Groups       u  = Azure AD Users          g  = Azure AD Groups      \n" +
		"      sp = Service Principals      ap = Applications            ra = Azure AD Roles Active\n" +
		"      rd = Azure AD Roles Defs\n" +
		"\n" +
		"    -xx                               Delete ALL cache local files\n" +
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
	case 1:    // Process 1-argument requests
		arg1 := os.Args[1]
		switch arg1 {         // First 1-arg check doesn't require API tokens or global vars
	    case "-v":
			PrintUsage()
		case "-cr":
			DumpCredentials()
		}
		SetupApiTokens()
		switch arg1 {        // The rest do required API tokens and global vars
		case "-xx":
			RemoveCacheFile("all")
		case "-tx", "-dx", "-ax", "-sx", "-mx", "-ux", "-gx", "-spx", "-apx", "-rax", "-rdx":
			t := arg1[1 : len(arg1)-1]  // Single out the object type
			RemoveCacheFile(t)          // Chop off the 1st 2 characters, to leverage oMap
		case "-st":
			PrintCountStatus()
		case "-dj", "-aj", "-sj", "-mj", "-uj", "-gj", "-spj", "-apj", "-raj", "-rdj":
			// Handle JSON-printing of all objects
			t := arg1[1 : len(arg1)-1]
			PrintAllJson(t)
		case "-d", "-a", "-s", "-m", "-u", "-g", "-sp", "-ap", "-ra", "-rd":
			// Handle tersely printing for all objects
			t := arg1[1:]               // Single out the object type
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
	case 2:    // Process 2-argument requests
		arg1 := os.Args[1] ; arg2 := os.Args[2]
		SetupApiTokens()
		switch arg1 {
		case "-vs":
			CompareSpecfile(arg2)
		case "-dj", "-aj", "-sj", "-mj", "-uj", "-gj", "-spj", "-apj", "-raj", "-rdj":
			t := arg1[1 : len(arg1)-1] // Single out our object type letter (see oMap)
			if ValidUuid(arg2) {
				x := GetAzObjectById(t, arg2) // Get single object by ID and print in JSON
				PrintJson(x)
			} else {
				oList := GetMatching(t, arg2) // Get all matching objects
				if len(oList) > 1 {           // Print all matching objects in JSON
					PrintJson(oList)
				} else if len(oList) > 0 {    // Print single matching object in JSON
					x := oList[0].(map[string]interface{})
					PrintJson(x)
				}
			}
		case "-d", "-a", "-s", "-m", "-u", "-g", "-sp", "-ap", "-ra", "-rd":
			t := arg1[1:] // Single out the object type
			if ValidUuid(arg2) {
				x := GetAzObjectById(t, arg2) // Get single object by ID
				PrintObject(t, x)             // Print in YAML
			} else {
				oList := GetMatching(t, arg2) // Get all matching objects
				if len(oList) > 1 {           // Print all matching objects tersely
					for _, i := range oList {
						x := i.(map[string]interface{})
						PrintTersely(t, x)
					}
				} else if len(oList) > 0 {    // Print single matching object in YAML
					x := oList[0].(map[string]interface{})
					PrintObject(t, x)
				}
			}
		default:
			PrintUsage()
		}
	case 3:    // Process 3-argument requests
		arg1 := os.Args[1] ; arg2 := os.Args[2] ; arg3 := os.Args[3]
		switch arg1 {
		case "-cri":
			SetupInterativeLogin(arg2, arg3)
		default:
			PrintUsage()
		}
	case 4:    // Process 4-argument requests
		arg1 := os.Args[1] ; arg2 := os.Args[2] ; arg3 := os.Args[3] ; arg4 := os.Args[4]
		switch arg1 {
		case "-cr":
			SetupAutomatedLogin(arg2, arg3, arg4)
		default:
			PrintUsage()
		}
	}
	exit(0)
}
