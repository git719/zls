// main.go

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

const (
	prgname = "zls"
	prgver  = "1.9.0"
	rUp     = "\x1B[2K\r" // See https://stackoverflow.com/questions/1508490/erase-the-current-printed-console-line
)

func PrintUsage() {
	fmt.Printf(prgname + " Azure RBAC and MS Graph listing utility v" + prgver + "\n" +
		"    -vs Specfile                   Compare YAML or JSON specfile to what's in Azure (only for d and a objects)\n" +
		"    -X[j]                          List all X objects tersely, with JSON output option\n" +
		"    -X[j] UUID|\"string\"            Show/list X object(s) matching on UUID or \"string\" attribute, JSON option\n" +
		"    -Xx                            Delete X object local file cache\n" +
		"\n" +
		"    Where 'X' can be any of these object types:\n" +
		"      d  = RBAC Role Definitions   a  = RBAC Role Assignments   s  = Azure Subscriptions  \n" +
		"      m  = Management Groups       u  = Azure AD Users          g  = Azure AD Groups      \n" +
		"      sp = Service Principals      ap = Applications            ad = Azure AD Roles\n" +
		"\n" +
		"    -xx                            Delete ALL cache local files\n" +
		"    -ar                            List all RBAC role assignments with resolved names\n" +
		"    -mt                            List Management Group and subscriptions tree\n" +
		"    -pags                          List all Azure AD Privileged Access Groups\n" +
		"    -st                            List local cache count and Azure count of all objects\n" +
		"\n" +
		"    -z                             Dump variables in running program\n" +
		"    -cr                            Dump values in credentials file\n" +
		"    -cr  TenantId ClientId Secret  Set up MSAL automated ClientId + Secret login\n" +
		"    -cri TenantId Username         Set up MSAL interactive browser popup login\n" +
		"    -tx                            Delete MSAL accessTokens cache file\n" +
		"    -v                             Print this usage page\n")
	os.Exit(0)
}

func SetupVariables(z *aza.AzaBundle) (aza.AzaBundle) {
	// Set up variable object struct
	*z = aza.AzaBundle{
		ConfDir:      "",
		CredsFile:    "credentials.yaml",
		TokenFile:    "accessTokens.json",
		TenantId:     "",
		ClientId:     "",
		ClientSecret: "",
		Interactive:  false,
		Username:     "",
		AuthorityUrl: "",
		MgToken:      "",
		MgHeaders:    aza.MapString{},
		AzToken:      "",
		AzHeaders:    aza.MapString{},  
	}
	// Set up configuration directory
	z.ConfDir = filepath.Join(os.Getenv("HOME"), "." + prgname)
	if utl.FileNotExist(z.ConfDir) {
		if err := os.Mkdir(z.ConfDir, 0700); err != nil {
			panic(err.Error())
		}
	}
	return *z
}

func main() {
	//TestFunction()
	numberOfArguments := len(os.Args[1:]) // Not including the program itself
	if numberOfArguments < 1 || numberOfArguments > 4 {
		PrintUsage() // Don't accept less than 1 or more than 4 arguments
	}

	var z aza.AzaBundle
	z = SetupVariables(&z)

	switch numberOfArguments {
	case 1: // Process 1-argument requests
		arg1 := os.Args[1]
		// This first set of 1-arg requests do not require API tokens to be set up
		switch arg1 {
	    case "-v":
			PrintUsage()
		case "-cr":
			aza.DumpCredentials(z)
		}
		z = aza.SetupApiTokens(&z) // The remaining 1-arg requests DO required API tokens to be set up
		switch arg1 {
		case "-xx":
			RemoveCacheFile("all", z)
		case "-tx", "-dx", "-ax", "-sx", "-mx", "-ux", "-gx", "-spx", "-apx", "-adx":
			t := arg1[1 : len(arg1)-1] // Single out the object type
			RemoveCacheFile(t, z)
		case "-dj", "-aj", "-sj", "-mj", "-uj", "-gj", "-spj", "-apj", "-adj":
			t := arg1[1 : len(arg1)-1]
			all := GetObjects(t, "", false, z) // false means do not force Azure call, ok to use cache
			utl.PrintJson(all) // Print entire set in JSON
		case "-d", "-a", "-s", "-m", "-u", "-g", "-sp", "-ap", "-ad":
			t := arg1[1:]
			all := GetObjects(t, "", false, z)
			for _, i := range all { // Print entire set tersely
				PrintTersely(t, i)
			}
		case "-ar":
			PrintRoleAssignmentReport(z)
		case "-mt":
			PrintMgTree(z)
		case "-pags":
			PrintPags(z)
		case "-st":
			PrintCountStatus(z)
		case "-z":
			aza.DumpVariables(z)
		default:
			PrintUsage()
		}
	case 2: // Process 2-argument requests
		arg1 := os.Args[1] ; arg2 := os.Args[2]
		z = aza.SetupApiTokens(&z)
		switch arg1 {
		case "-vs":
			CompareSpecfileToAzure(arg2, z)
		case "-dj", "-aj", "-sj", "-mj", "-uj", "-gj", "-spj", "-apj", "-adj":
			t := arg1[1 : len(arg1)-1] // Single out the object type
			if utl.ValidUuid(arg2) { // Search/print single object, if it's valid UUID
				x := GetObjectById(t, arg2, z)
				utl.PrintJson(x)
			} else {
				matchingObjects := GetObjects(t, arg2, false, z)
				if len(matchingObjects) == 1 {
					utl.PrintJson(matchingObjects[0]) // Print single matching object in JSON
				} else if len(matchingObjects) > 1 {
					utl.PrintJson(matchingObjects) // Print all matching objects in JSON
				}
			}
		case "-d", "-a", "-s", "-m", "-u", "-g", "-sp", "-ap", "-ad":
			t := arg1[1:] // Single out the object type
			if utl.ValidUuid(arg2) { // Search/print single object, if it's valid UUID
				x := GetObjectById(t, arg2, z)
				PrintObject(t, x, z)
			} else {
				matchingObjects := GetObjects(t, arg2, false, z)
				if len(matchingObjects) == 1 {
					x := matchingObjects[0].(map[string]interface{})
					PrintObject(t, x, z)
				} else if len(matchingObjects) > 1 {
					for _, i := range matchingObjects { // Print all matching object teresely
						x := i.(map[string]interface{})
						PrintTersely(t, x)
					}
				}
			}
		default:
			PrintUsage()
		}
	case 3: // Process 3-argument requests
		arg1 := os.Args[1] ; arg2 := os.Args[2] ; arg3 := os.Args[3]
		switch arg1 {
		case "-cri":
			z.TenantId = arg2 ; z.Username = arg3
			aza.SetupInterativeLogin(z)
		default:
			PrintUsage()
		}
	case 4: // Process 4-argument requests
		arg1 := os.Args[1] ; arg2 := os.Args[2]; arg3 := os.Args[3] ; arg4 := os.Args[4]
		switch arg1 {
		case "-cr":
			z.TenantId = arg2 ; z.ClientId = arg3 ; z.ClientSecret = arg4
			aza.SetupAutomatedLogin(z)
		default:
			PrintUsage()
		}
	}
	os.Exit(0)
}
