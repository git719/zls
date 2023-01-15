// printing.go

package main

import (
	"fmt"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintCountStatus(z aza.AzaBundle) {
	fmt.Printf("Note: Counting some Azure resources can take a long time.\n")
	fmt.Printf("%-36s %10s %10s\n", "OBJECTS", "LOCAL","AZURE")
	fmt.Printf("%-36s %10d %10d\n", "Azure AD Users", UsersCountLocal(z), UsersCountAzure(z))
	fmt.Printf("%-36s %10d %10d\n", "Azure AD Groups", GroupsCountLocal(z), GroupsCountAzure(z))
	fmt.Printf("%-36s %10d %10d\n", "Azure App Registrations", AppsCountLocal(z), AppsCountAzure(z))
	nativeSpsLocal, msSpsLocal := SpsCountLocal(z)
	nativeSpsAzure, msSpsAzure := SpsCountAzure(z)
	fmt.Printf("%-36s %10d %10d\n", "Azure SPs (multi-tenant)", msSpsLocal, msSpsAzure)
	fmt.Printf("%-36s %10d %10d\n", "Azure SPs (native to tenant)", nativeSpsLocal, nativeSpsAzure)
	fmt.Printf("%-36s %10d %10d\n", "Azure AD Roles", AdRolesCountLocal(z), AdRolesCountAzure(z))
	fmt.Printf("%-36s %10d %10d\n", "Azure Management Groups", MgGroupCountLocal(z), MgGroupCountAzure(z))
	fmt.Printf("%-36s %10d %10d\n", "Azure Subscriptions", SubsCountLocal(z), SubsCountAzure(z))
	builtinLocal, customLocal := RoleDefinitionCountLocal(z)
	builtinAzure, customAzure := RoleDefinitionCountAzure(z)
	fmt.Printf("%-36s %10d %10d\n", "Resource Role Definitions BuiltIn", builtinLocal, builtinAzure)
	fmt.Printf("%-36s %10d %10d\n", "Resource Role Definitions Custom", customLocal, customAzure)
	fmt.Printf("%-36s %10d %10d\n", "Resource Role Assignments", RoleAssignmentsCountLocal(z), RoleAssignmentsCountAzure(z))
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

func PrintObject(t string, x map[string]interface{}, z aza.AzaBundle) {
	switch t {
	case "d":
		PrintRoleDefinition(x, z)
	case "a":
		PrintRoleAssignment(x, z)
	case "s":
		PrintSubscription(x)
	case "m":
		PrintMgGroup(x)
	case "u":
		PrintUser(x, z)
	case "g":
		PrintGroup(x, z)
	case "sp":
		PrintSp(x, z)
	case "ap":
		PrintApp(x, z)
	case "ad":
		PrintAdRole(x, z)
	}
}

func PrintMemberOfs(t string, memberOf []interface{}) {
	// Print all memberOf entries
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
