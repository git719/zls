// printing.go

package main

import (
	"fmt"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintTersely(t string, x map[string]interface{}) {
	// List this single object of type 't' tersely (minimal attributes)
	switch t {
	case "d":
		xProp := x["properties"].(map[string]interface{})
		fmt.Printf("%s  %-60s  %s\n", StrVal(x["name"]), StrVal(xProp["roleName"]), StrVal(xProp["type"]))
	// case "a":
	// 	Id := StrVal(x["name"]) // This is the role assignment's UUID
	// 	Props := x["properties"].(map[string]interface{})
	// 	RoleId := utl.LastElem(StrVal(Props["roleDefinitionId"]), "/")
	// 	PrincipalId := StrVal(Props["principalId"])
	// 	PrinType := StrVal(Props["principalType"])
	// 	Scope := StrVal(Props["scope"])
	// 	print("%s  %s  %s %-20s %s\n", Id, RoleId, PrincipalId, "("+PrinType+")", Scope)
	case "s":
		fmt.Printf("%s  %-10s  %s\n", StrVal(x["subscriptionId"]), StrVal(x["state"]), StrVal(x["displayName"]))
	case "m":
		xProp := x["properties"].(map[string]interface{})
		fmt.Printf("%-38s  %-20s  %s\n", StrVal(x["name"]), StrVal(xProp["displayName"]), MgType(StrVal(x["type"])))
	// case "u", "g", "sp", "ap", "ra", "rd":
	// 	Id := StrVal(x["id"])
	// 	Name := StrVal(x["displayName"])
	// 	Type := StrVal(x["servicePrincipalType"])
	// 	AppId := StrVal(x["appId"])
	// 	Desc := StrVal(x["description"])
	// 	switch t {
	// 	case "u":
	// 		Upn := StrVal(x["userPrincipalName"])
	// 		onPremisesSamAccountName := StrVal(x["onPremisesSamAccountName"])
	// 		print("%s  %-50s %-18s %s\n", Id, Upn, onPremisesSamAccountName, Name)
	// 	case "g":
	// 		print("%s  %s\n", Id, Name)
	// 	case "sp":
	// 		print("%s  %-60s %-22s %s\n", Id, Name, Type, AppId)
	// 	case "ap":
	// 		print("%s  %-60s %s\n", Id, Name, AppId)
	// 	case "ra":
	// 		print("%s  %-60s %s\n", Id, Name, Desc)
	// 	case "rd":
	// 		BuiltIn := "Custom"
	// 		if StrVal(x["isBuiltIn"]) == "true" { BuiltIn = "BuiltIn" }
	// 		Enabled := "NotEnabled"
	// 		if StrVal(x["isEnabled"]) == "true" { Enabled = "Enabled" }
	// 		print("%s  %-60s  %s  %s\n", Id, Name, BuiltIn, Enabled)
	// 	}
	}
}

func PrintObject(t string, x JsonObject, z aza.AzaBundle, oMap MapString) {
	switch t {
	case "d":
		PrintRoleDefinition(x, z, oMap)
	// case "a":
	// 	PrintRoleAssignment(x)
	case "s":
		PrintSubscription(x)
	case "m":
		PrintMgGroup(x)
	// case "u":
	// 	PrintUser(x)
	// case "g":
	// 	PrintGroup(x)
	// case "sp":
	// 	PrintSP(x)
	// case "ap":
	// 	PrintApp(x, z)
	// case "ra":
	// 	PrintAdRole(x)     // Active AD role
	// case "rd":
	// 	PrintAdRoleDef(x)  // Definition of AD role
	}
}

func PrintMemberOfs(t string, memberOf []interface{}) {
	// Print all memberof entries
	// Object type t is for future use
	if len(memberOf) > 0 {
		print("memberof:\n")
		for _, i := range memberOf {
			o := i.(map[string]interface{}) // Assert as JSON object type
			Name := StrVal(o["displayName"])
			Type := utl.LastElem(StrVal(o["@odata.type"]), ".")
			Id := StrVal(o["id"])
			print("  %-50s %s (%s)\n", Name, Id, Type)
		}
	} else {
		print("%-21s %s\n", "memberof:", "None")
	}
}
