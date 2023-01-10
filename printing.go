// printing.go

package main

import (
	"fmt"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintTersely(t string, x JsonObject) {
	// List this single object of type 't' tersely (minimal attributes)
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
	case "u", "g", "sp", "ap", "ra", "rd":
		switch t {
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
		case "ra":
			fmt.Printf("%s  %-60s %s\n", StrVal(x["id"]), StrVal(x["displayName"]), StrVal(x["description"]))
		case "rd":
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
	// case "g":
	// 	PrintGroup(x)
	// case "sp":
	// 	PrintSP(x)
	// case "ap":
	// 	PrintApp(x, z)
	case "ra":
		PrintAdRole(x, z) // Active AD role
	case "rd":
		PrintAdRoleDef(x) // Definition of AD role
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
		fmt.Printf("%-21s %s\n", "memberof:", "None")
	}
}
