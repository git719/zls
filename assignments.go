// assignments.go

package main

import (
	"strings"
)

func PrintRoleAssignmentReport() {
	// Print a human-readable report of all role assignments
	roleMap := GetIdNameMap("d")
	subMap := GetIdNameMap("s")
	userMap := GetIdNameMap("u")
	groupMap := GetIdNameMap("g")
	spMap := GetIdNameMap("sp")

	for _, i := range GetAllObjects("a") { // Iterate through all objects
		x := i.(map[string]interface{}) // Assert JSON object type
		xProp := x["properties"].(map[string]interface{})
		Rid := LastElem(StrVal(xProp["roleDefinitionId"]), "/")
		Pid := StrVal(xProp["principalId"])
		Type := StrVal(xProp["principalType"])
		pName := "ID-Not-Found"
		switch Type {
		case "User":
			pName = userMap[Pid]
		case "ServicePrincipal":
			pName = spMap[Pid]
		case "Group":
			pName = groupMap[Pid]
		}

		Scope := StrVal(xProp["scope"])
		if strings.HasPrefix(Scope, "/subscriptions") {
			// Replace sub ID to name
			split := strings.Split(Scope, "/")
			// Map subscription Id to its name + the rest of the resource path
			Scope = subMap[split[2]] + " " + strings.Join(split[3:], "/")
		}
		Scope = strings.TrimSpace(Scope)

		print("\"%s\",\"%s\",\"%s\",\"%s\"\n", roleMap[Rid], pName, Type, Scope)
	}
}

func PrintRoleAssignment(x map[string]interface{}) {
	// Print role definition object in YAML-like style format
	if x == nil {
		return
	}

	if x["name"] != nil {
		print("id: %s\n", StrVal(x["name"]))
	}
	// ######### DEBUG ##########
	//PrintJson(x); print("\n")

	print("properties:\n")
	if x["properties"] == nil {
		print("  <Missing??>\n")
		return
	}
	xProp := x["properties"].(map[string]interface{})

	roleMap := GetIdNameMap("d")  // Get role definitions id:name map
	roleId := LastElem(StrVal(xProp["roleDefinitionId"]), "/")
	print("  %-17s %s  # roleName = \"%s\"\n", "roleDefinitionId:", roleId, roleMap[roleId])

	var nameMap map[string]string
	pType := StrVal(xProp["principalType"])
	switch pType {
	case "User":
		nameMap = GetIdNameMap("u")
	case "ServicePrincipal":
		nameMap = GetIdNameMap("sp")
	case "Group":
		nameMap = GetIdNameMap("g")
	default:
		pType = "not provided"
	}
	pId := StrVal(xProp["principalId"])
	pName := nameMap[StrVal(xProp["principalId"])]
	if pName == "" {
		pName = "???" 
	}
	print("  %-17s %s  # principaltype = %s, displayName = \"%s\"\n", "principalId:", pId, pType, pName)

	subMap := GetIdNameMap("s")  // Get subscriptions id:name map
	scope := StrVal(xProp["scope"])
	if strings.HasPrefix(scope, "/subscriptions") {
		split := strings.Split(scope, "/")
		subName := subMap[split[2]]
		print("  %-17s %s  # Sub = %s\n", "scope:", scope, subName)
	} else {
		print("  %-17s %s\n", "scope:", scope)
	}
}

func GetRoleAssignments() (oList []interface{}) {
	// Get all role assigments from Azure
	// See https://docs.microsoft.com/en-us/rest/api/authorization/role-assignments/list

	oList = nil

	// First, build list of scopes, which is all subscriptions to look under, and include the tenant root level
	scopes := GetSubScopes()
	scopes = append(scopes, "/providers/Microsoft.Management/managementGroups/" + tenant_id)
	// Should we include all other Management Groups scopes?

	var uuids []string  // Keep track of each unique role definition to whittle out repeats that come up in lower scopes
	apiCalls := 1       // Track number of API calls below
	subMap := GetIdNameMap("s")  // To ease printing sub names during calling

	// Look for objects under all these scopes
	params := map[string]string{
		"api-version": "2022-04-01",  // roleAssignments
	}
	for _, scope := range scopes {
		subId := LastElem(scope, "/")
		subName := subMap[subId]
		url := az_url + scope + "/providers/Microsoft.Authorization/roleAssignments"
		r := ApiGet(url, az_headers, params, false)
		if r["value"] != nil {
			assignments := r["value"].([]interface{}) // Assert as JSON array type
			print("\r(API calls = %d) %d assignments at '%s'", apiCalls, len(assignments), subName)
			PadSpaces(20)
			for _, i := range assignments {
				x := i.(map[string]interface{}) // Assert as JSON object type
				uuid := StrVal(x["name"])  // NOTE that 'name' key is the role assignment UUID
				if !ItemInList(uuid, uuids) {
					// Add this role assignment to growing list ONLY if it's NOT in it already - they DO repeat!
					oList = append(oList, x)
					uuids = append(uuids, uuid)
				}
			}
		}
		ApiErrorCheck(r, trace())
		apiCalls++
	}
	print("\n")
	return oList
}

func GetAzRoleAssignment(x map[string]interface{}) (y map[string]interface{}) {
	// Retrieve role assignment y from Azure if it exists and matches given x object's roleId, principalId, and scope
	// Above 3 parameters make role assignments unique

	// First, make sure x is a searchable role assignment object
	if x == nil {
		return nil      // Don't look for empty objects
	}
	xProps := x["properties"].(map[string]interface{})
	if xProps == nil {  // Return nil if properties missing
		return nil
	}
	xRoleDefinitionId := LastElem(StrVal(xProps["roleDefinitionId"]), "/")
	xPrincipalId      := StrVal(xProps["principalId"])
	xScope            := StrVal(xProps["scope"])
	if xScope == "" {
		xScope = StrVal(xProps["Scope"])  // Account for capitalized Scope key
	}
	if xScope == "" || xPrincipalId == "" || xRoleDefinitionId == "" {
		return nil
	}

	// Get all role assignments for xPrincipalId under xScope
    params := map[string]string{
		"api-version": "2022-04-01",  // roleAssignments
		"$filter":     "principalId eq '" + xPrincipalId + "'",
	}
	r := ApiGet(az_url + xScope + "/providers/Microsoft.Authorization/roleAssignments", az_headers, params, false)
	if r != nil && r["value"] != nil {
		results := r["value"].([]interface{})
		for _, i := range results {
			y = i.(map[string]interface{})
			yProps := y["properties"].(map[string]interface{})
			yScope := StrVal(yProps["scope"])
			yRoleDefinitionId := LastElem(StrVal(yProps["roleDefinitionId"]), "/")
			if yScope == xScope && yRoleDefinitionId == xRoleDefinitionId {
				return y  // We found it
			}
		}
	}
	ApiErrorCheck(r, trace())
	return nil
}
