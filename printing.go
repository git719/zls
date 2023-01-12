// printing.go

package main

import (
	"fmt"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintCountStatus(z aza.AzaBundle, oMap MapString) {
	fmt.Printf("Note: Counting Azure objects can take some time.\n")
	fmt.Printf("%-38s %-20s %s\n", "OBJECTS", "LOCAL_CACHE_COUNT","AZURE_COUNT")
	fmt.Printf("%-38s %-20d %d\n", "Groups", GroupsCountLocal(z), GroupsCountAzure(z))
	fmt.Printf("%-38s %-20d %d\n", "Users", UsersCountLocal(z), UsersCountAzure(z))
	fmt.Printf("%-38s %-20d %d\n", "App Registrations", AppsCountLocal(z), AppsCountAzure(z))

	nativeSpsLocal, msSpsLocal := SpsCountLocal(z)
	nativeSpsAzure, msSpsAzure := SpsCountAzure(z)
	fmt.Printf("%-38s %-20d %d\n", "SPs Owned by Microsoft", msSpsLocal, msSpsAzure)
	fmt.Printf("%-38s %-20d %d\n", "SPs Owned by this Tenant", nativeSpsLocal, nativeSpsAzure)

	fmt.Printf("%-38s %-20d %d\n", "Azure AD Roles", AdRolesCountLocal(z), AdRolesCountAzure(z))
    fmt.Println("-----------------------------")
	fmt.Printf("%-38s %-20d %d\n", "Management Groups", ObjectCountLocal("m", z, oMap), ObjectCountAzure("m", z, oMap))
	fmt.Printf("%-38s %-20d %d\n", "Subscriptions", ObjectCountLocal("s", z, oMap), ObjectCountAzure("s", z, oMap))

	builtinLocal, customLocal := RoleDefinitionCountLocal(z)
	builtinAzure, customAzure := RoleDefinitionCountAzure(z, oMap)
	fmt.Printf("%-38s %-20d %d\n", "RBAC Role Definitions BuiltIn", builtinLocal, builtinAzure)
	fmt.Printf("%-38s %-20d %d\n", "RBAC Role Definitions Custom", customLocal, customAzure)

	assignmentsLocal := len(GetRoleAssignments("", false, false, z, oMap)) // false = prefer local, false = be silent
	assignmentsAzure := len(GetRoleAssignments("", true, false, z, oMap)) // true = force a call to Azure, false = be silent
	fmt.Printf("%-38s %-20d %d\n", "RBAC Role Assignments", assignmentsLocal, assignmentsAzure)
}

func PrintTersely(t string, object interface{}) {
	// Print this single object of type 't' tersely (minimal attributes)
	x := object.(map[string]interface{}) // Assert as JSON object
	switch t {
	case "d":
		xProp := x["properties"].(map[string]interface{})
		fmt.Printf("%s  %-60s  %s\n", StrVal(x["name"]), StrVal(xProp["roleName"]), StrVal(xProp["type"]))
	case "a":
		xProp := x["properties"].(map[string]interface{})
		rdId := utl.LastElem(StrVal(xProp["roleDefinitionId"]), "/")
		principalId := StrVal(xProp["principalId"])
		principalType := StrVal(xProp["principalType"])
		scope := StrVal(xProp["scope"])
		fmt.Printf("%s  %s  %s %-20s %s\n", StrVal(x["name"]), rdId, principalId, "(" + principalType + ")", scope)
	case "s":
		fmt.Printf("%s  %-10s  %s\n", StrVal(x["subscriptionId"]), StrVal(x["state"]), StrVal(x["displayName"]))
	case "m":
		xProp := x["properties"].(map[string]interface{})
		fmt.Printf("%-38s  %-20s  %s\n", StrVal(x["name"]), StrVal(xProp["displayName"]), MgType(StrVal(x["type"])))
	case "u":
		upn := StrVal(x["userPrincipalName"])
		onPremisesSamAccountName := StrVal(x["onPremisesSamAccountName"])
		fmt.Printf("%s  %-50s %-18s %s\n", StrVal(x["id"]), upn, onPremisesSamAccountName, StrVal(x["displayName"]))
	case "g":
		fmt.Printf("%s  %s\n", StrVal(x["id"]), StrVal(x["displayName"]))
	case "sp":
		fmt.Printf("%s  %-60s %-22s %s\n", StrVal(x["id"]), StrVal(x["displayName"]), StrVal(x["servicePrincipalType"]), StrVal(x["appId"]))
	case "ap":
		fmt.Printf("%s  %-60s %s\n", StrVal(x["id"]), StrVal(x["displayName"]), StrVal(x["appId"]))
	case "ad":
		builtIn := "Custom"
		if StrVal(x["isBuiltIn"]) == "true" {
			builtIn = "BuiltIn"
		}
		enabled := "NotEnabled"
		if StrVal(x["isEnabled"]) == "true" {
			enabled = "Enabled"
		}
		fmt.Printf("%s  %-60s  %s  %s\n", StrVal(x["id"]), StrVal(x["displayName"]), builtIn, enabled)
	}
}

func PrintObject(t string, x JsonObject, z aza.AzaBundle, oMap MapString) {
	switch t {
	case "d":
		PrintRoleDefinition(x, z, oMap)
	case "a":
		PrintRoleAssignment(x, z, oMap)
	case "s":
		PrintSubscription(x)
	case "m":
		PrintMgGroup(x)
	case "u":
		PrintUser(x, z, oMap)
	case "g":
		PrintGroup(x, z, oMap)
	case "sp":
		PrintSp(x, z, oMap)
	case "ap":
		PrintApp(x, z, oMap)
	case "ad":
		PrintAdRole(x, z)
	}
}

func PrintMemberOfs(t string, memberOf JsonArray) {
	// Print all memberof entries
	// Object type t is for future use
	if len(memberOf) > 0 {
		fmt.Printf("memberof:\n")
		for _, i := range memberOf {
			x := i.(map[string]interface{}) // Assert as JSON object type
			Type := utl.LastElem(StrVal(x["@odata.type"]), ".")
			fmt.Printf("  %-50s %s (%s)\n", StrVal(x["displayName"]), StrVal(x["id"]), Type)
		}
	} else {
		fmt.Printf("%s: %s\n", "memberof", "None")
	}
}
