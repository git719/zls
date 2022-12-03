// printing.go

package main

import "log"

func PrintAllTersely(t string) {
	// List tersely all object of type t
	for _, i := range GetAllObjects(t) { // Iterate through all objects
		x := i.(map[string]interface{}) // Assert JSON object type
		PrintTersely(t, x)
	}
}

func PrintTersely(t string, x map[string]interface{}) {
	// List this single object of type 't' tersely (minimal attributes)
	switch t {
	case "d":
		Id := StrVal(x["name"]) // This is the role definition's UUID
		Props := x["properties"].(map[string]interface{})
		Type := StrVal(Props["type"])
		Name := StrVal(Props["roleName"])
		print("%s  %-60s  %s\n", Id, Name, Type)
	case "a":
		Id := StrVal(x["name"]) // This is the role assignment's UUID
		Props := x["properties"].(map[string]interface{})
		RoleId := LastElem(StrVal(Props["roleDefinitionId"]), "/")
		PrincipalId := StrVal(Props["principalId"])
		PrinType := StrVal(Props["principalType"])
		Scope := StrVal(Props["scope"])
		print("%s  %s  %s %-20s %s\n", Id, RoleId, PrincipalId, "("+PrinType+")", Scope)
	case "s":
		Id := StrVal(x["subscriptionId"])
		State := StrVal(x["state"])
		Name := StrVal(x["displayName"])
		print("%s  %-10s  %s\n", Id, State, Name)
	case "m":
		Id := StrVal(x["name"])
		Props := x["properties"].(map[string]interface{})
		Name := StrVal(Props["displayName"])
		Type := MGType(StrVal(x["type"]))
		print("%-38s  %-20s  %s\n", Id, Name, Type)
	case "u", "g", "sp", "ap", "ra", "rd":
		Id := StrVal(x["id"])
		Name := StrVal(x["displayName"])
		Type := StrVal(x["servicePrincipalType"])
		AppId := StrVal(x["appId"])
		Desc := StrVal(x["description"])
		switch t {
		case "u":
			Upn := StrVal(x["userPrincipalName"])
			onPremisesSamAccountName := StrVal(x["onPremisesSamAccountName"])
			print("%s  %-50s %-18s %s\n", Id, Upn, onPremisesSamAccountName, Name)
		case "g":
			print("%s  %s\n", Id, Name)
		case "sp":
			print("%s  %-60s %-22s %s\n", Id, Name, Type, AppId)
		case "ap":
			print("%s  %-60s %s\n", Id, Name, AppId)
		case "ra":
			print("%s  %-60s %s\n", Id, Name, Desc)
		case "rd":
			BuiltIn := "BuiltIn="+StrVal(x["isBuiltIn"])
			Enabled := "Enabled="+StrVal(x["isEnabled"])
			print("%s  %-60s  %s  %s\n", Id, Name, BuiltIn, Enabled)
		}
	}
}

func PrintObject(t string, x map[string]interface{}) {
	if x["id"] == nil {
		return
	}
	switch t {
	case "d":
		PrintRoleDefinition(x)
	case "a":
		PrintRoleAssignment(x)
	case "s":
		PrintSubscription(x)
	case "m":
		PrintManagementGroup(x)
	case "u":
		PrintUser(x)
	case "g":
		PrintGroup(x)
	case "sp":
		PrintSP(x)
	case "ap":
		PrintApp(x)
	case "ra":
		PrintAdRole(x)     // Active AD role
	case "rd":
		PrintAdRoleDef(x)  // Definition of AD role
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
			Type := LastElem(StrVal(o["@odata.type"]), ".")
			Id := StrVal(o["id"])
			print("  %-50s %s (%s)\n", Name, Id, Type)
		}
	} else {
		print("%-21s %s\n", "memberof:", "None")
	}
}

func CompareSpecfile(t, f string) {
	// Compare object of type t defined in specfile f with what's really in Azure
	switch t {
	case "d":
		// Load specfile
		jsonFile := LoadFileJSON(f)
		if jsonFile == nil {
			log.Printf("Invalid JSON specfile '%s'\n", f)
			return
		}
		x := jsonFile.(map[string]interface{}) // Assert as single JSON object
		print("==== This SPECFILE ===============================================\n")
		PrintJSON(x)

		xProps := x["properties"].(map[string]interface{})
		Name := StrVal(xProps["roleName"])
		// Search for role definition in all scopes defined in specfile
		notFound := true
		scopes := xProps["assignableScopes"].([]interface{})
		for _, scope := range scopes {
			url := az_url + "/" + scope.(string)
			url += "/providers/Microsoft.Authorization/roleDefinitions?$filter=roleName+eq+'" + Name + "'"
			r := APIGet(url, az_headers, nil, false)
			if r["value"] != nil && len(r["value"].([]interface{})) == 1 {
				y := r["value"].([]interface{})
				z := y[0].(map[string]interface{})
				if z["id"] != nil {
					notFound = false
					print("==== What's in AZURE (in YAML-like format for easier reading) ====\n")
					PrintObject("d", z)
					break // Break, since any other subsequent match will be exactly the same
				}
			}
		}
		if notFound {
			print("==== What's in AZURE =============================================\n")
			print("Role definition as defined in this specfile does NOT exist in Azure.\n")
		}
	case "a":
		// Load specfile
		x := LoadFileYAML(f)
		if x == nil {
			print("Invalid YAML specfile '%s'\n", f)
			return
		}
		print("==== SPECFILE ===========================\n")
		PrintYAML(x)

		xProps := x["properties"].(map[string]interface{})
		roleId := LastElem(StrVal(xProps["roleDefinitionId"]), "/")
		principalId := StrVal(xProps["principalId"])
		scope := StrVal(xProps["scope"])
		if scope == "" {
			scope = StrVal(xProps["Scope"]) // Uppercase version
		}
		// Search for role assignment in the scope defined in specfile
		url := az_url + scope
		url += "/providers/Microsoft.Authorization/roleAssignments?$filter=principalId+eq+'" + principalId + "'"
		r := APIGet(url, az_headers, nil, false)
		if r["value"] != nil && len(r["value"].([]interface{})) > 0 {
			for _, i := range r["value"].([]interface{}) {
				y := i.(map[string]interface{})
				yProps := y["properties"].(map[string]interface{})
				azRoleId := LastElem(StrVal(yProps["roleDefinitionId"]), "/")
				azPrincipalId := StrVal(yProps["principalId"])
				azScope := StrVal(yProps["scope"])
				if azRoleId == roleId && azPrincipalId == principalId && azScope == scope {
					print("==== AZURE ==============================\n")
					PrintObject("a", y)
					break // Break loop as soon as we find match
				}
			}
		} else {
			print("==== AZURE ==============================\n")
			print("Role assignment does not exist as defined in specfile\n")
		}
	default:
		print("This option is not yet available.\n")
	}
}
