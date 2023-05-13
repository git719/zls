// main.go

package main

import (
	"fmt"
	"github.com/git719/maz"
	"github.com/git719/utl"
	"os"
	"path/filepath"
)

const (
	prgname = "zls"
	prgver  = "1.11.12"
)

func PrintUsage() {
	X := utl.Red("X")
	fmt.Printf(prgname + " Azure Resource RBAC and MS Graph READER v" + prgver + "\n" +
		"    READER FUNCTIONS\n" +
		"    UUID                              Show object for given UUID\n" +
		"    -vs Specfile                      Compare YAML or JSON specfile to what's in Azure (only for d and a objects)\n" +
		"    -" + X + "[j] [Specifier]                 List all " + X + " objects tersely, with option for JSON output and/or match on Specifier\n" +
		"    -" + X + "x                               Delete " + X + " object local file cache\n\n" +
		"      Where '" + X + "' can be any of these object types:\n" +
		"      d  = RBAC Role Definitions   a  = RBAC Role Assignments   s  = Azure Subscriptions  \n" +
		"      m  = Management Groups       u  = Azure AD Users          g  = Azure AD Groups      \n" +
		"      sp = Service Principals      ap = Applications            ad = Azure AD Roles\n" +
		"\n" +
		"    -xx                               Delete ALL cache local files\n" +
		"    -ar                               List all RBAC role assignments with resolved names\n" +
		"    -mt                               List Management Group and subscriptions tree\n" +
		"    -pags                             List all Azure AD Privileged Access Groups\n" +
		"    -st                               List local cache count and Azure count of all objects\n" +
		"\n" +
		"    -z                                Dump important program variables\n" +
		"    -cr                               Dump values in credentials file\n" +
		"    -cr  TenantId ClientId Secret     Set up MSAL automated ClientId + Secret login\n" +
		"    -cri TenantId Username            Set up MSAL interactive browser popup login\n" +
		"    -tx                               Delete MSAL accessTokens cache file\n" +
		"    -tc \"TokenString\"                 Dump token claims\n" +
		"    -v                                Print this usage page\n")
	os.Exit(0)
}

func SetupVariables(z *maz.Bundle) maz.Bundle {
	// Set up variable object struct
	*z = maz.Bundle{
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
		MgHeaders:    map[string]string{},
		AzToken:      "",
		AzHeaders:    map[string]string{},
	}
	// Set up configuration directory
	z.ConfDir = filepath.Join(os.Getenv("HOME"), ".maz") // IMPORTANT: Setting config dir = "~/.maz"

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

	var z maz.Bundle
	z = SetupVariables(&z)

	switch numberOfArguments {
	case 1: // Process 1-argument requests
		arg1 := os.Args[1]
		// This first set of 1-arg requests do not require API tokens to be set up
		switch arg1 {
		case "-v":
			PrintUsage()
		case "-cr":
			maz.DumpCredentials(z)
		}
		z = maz.SetupApiTokens(&z) // The remaining 1-arg requests DO required API tokens to be set up
		switch arg1 {
		case "-xx":
			maz.RemoveCacheFile("all", z)
		case "-tx", "-dx", "-ax", "-sx", "-mx", "-ux", "-gx", "-spx", "-apx", "-adx":
			t := arg1[1 : len(arg1)-1] // Single out the object type
			maz.RemoveCacheFile(t, z)
		case "-dj", "-aj", "-sj", "-mj", "-uj", "-gj", "-spj", "-apj", "-adj":
			t := arg1[1 : len(arg1)-1]
			all := maz.GetObjects(t, "", false, z) // false means do not force Azure call, ok to use cache
			utl.PrintJsonColor(all)                // Print entire set in JSON
		case "-d", "-a", "-s", "-m", "-u", "-g", "-sp", "-ap", "-ad":
			t := arg1[1:]
			all := maz.GetObjects(t, "", false, z)
			for _, i := range all { // Print entire set tersely
				maz.PrintTersely(t, i)
			}
		case "-ar":
			maz.PrintRoleAssignmentReport(z)
		case "-mt":
			maz.PrintMgTree(z)
		case "-pags":
			maz.PrintPags(z)
		case "-st":
			maz.PrintCountStatus(z)
		case "-z":
			maz.DumpVariables(z)
		default:
			c := rune(arg1[0])                            // Grab 1st charater of string, to check if it's hex
			if utl.IsHexDigit(c) && utl.ValidUuid(arg1) { // If valid UUID, search/print matching object(s?)
				maz.PrintObjectByUuid(arg1, z)
			} else {
				PrintUsage()
			}
		}
	case 2: // Process 2-argument requests
		arg1 := os.Args[1]
		arg2 := os.Args[2]
		z = maz.SetupApiTokens(&z)
		switch arg1 {
		case "-tc":
			maz.DecodeJwtToken(arg2)
		case "-vs":
			maz.CompareSpecfileToAzure(arg2, z)
		case "-dj", "-aj", "-sj", "-mj", "-uj", "-gj", "-spj", "-apj", "-adj":
			t := arg1[1 : len(arg1)-1] // Single out the object type
			maz.PrintMatching("json", t, arg2, z)
		case "-d", "-a", "-s", "-m", "-u", "-g", "-sp", "-ap", "-ad":
			t := arg1[1:] // Single out the object type
			maz.PrintMatching("reg", t, arg2, z)
		default:
			PrintUsage()
		}
	case 3: // Process 3-argument requests
		arg1 := os.Args[1]
		arg2 := os.Args[2]
		arg3 := os.Args[3]
		switch arg1 {
		case "-cri":
			z.TenantId = arg2
			z.Username = arg3
			maz.SetupInterativeLogin(z)
		default:
			PrintUsage()
		}
	case 4: // Process 4-argument requests
		arg1 := os.Args[1]
		arg2 := os.Args[2]
		arg3 := os.Args[3]
		arg4 := os.Args[4]
		switch arg1 {
		case "-cr":
			z.TenantId = arg2
			z.ClientId = arg3
			z.ClientSecret = arg4
			maz.SetupAutomatedLogin(z)
		default:
			PrintUsage()
		}
	}
	os.Exit(0)
}
