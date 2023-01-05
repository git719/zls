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
	// Print role definition object in YAML
	if x == nil { return }
	if x["name"] != nil { print("id: %s\n", StrVal(x["name"])) }

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
	if pName == "" { pName = "???" }
	print("  %-17s %s  # principaltype = %s, displayName = \"%s\"\n", "principalId:", pId, pType, pName)

	subMap := GetIdNameMap("s")  // Get subscriptions id:name map
	scope := StrVal(xProp["scope"])
	if scope == "" { scope = StrVal(xProp["Scope"]) }  // Account for possibly capitalized key
	if strings.HasPrefix(scope, "/subscriptions") {
		split := strings.Split(scope, "/")
		subName := subMap[split[2]]
		print("  %-17s %s  # Sub = %s\n", "scope:", scope, subName)
	} else if scope == "/" {
		print("  %-17s %s  # Entire tenant\n", "scope:", scope)
	} else {
		print("  %-17s %s\n", "scope:", scope)
	}
}

func GetAzRoleAssignmentAll() (oList []interface{}) {
	// Get all role assigments from Azure

	// See https://learn.microsoft.com/en-us/azure/role-based-access-control/role-assignments-list-rest
	oList = nil
	scopes := GetAzRbacScopes()  // Look for objects under all the RBAC hierarchy scopes
	var uuids []string           // Keep track of each unique objects to whittle out inherited repeats
	calls := 1                   // Track number of API calls below
	params := map[string]string{
		"api-version": "2022-04-01",  // roleAssignments
	}
	for _, scope := range scopes {
		url := az_url + scope + "/providers/Microsoft.Authorization/roleAssignments"
		r := ApiGet(url, az_headers, params)
		if r["value"] != nil {
			assignments := r["value"].([]interface{}) // Assert as JSON array type
			// Using global var rUp to overwrite last line. Defer newline until done
			print("%s(API calls = %d) %d assignments in set %d", rUp, calls, len(assignments), calls)
			for _, i := range assignments {
				x := i.(map[string]interface{}) // Assert as JSON object type
				uuid := StrVal(x["name"])  // NOTE that 'name' key is the role assignment UUID
				if !ItemInList(uuid, uuids) {
					// Role assignments DO repeat! Add to growing list ONLY if it's NOT in it already
					oList = append(oList, x)
					uuids = append(uuids, uuid)
				}
			}
		}
		ApiErrorCheck(r, trace())
		calls++
	}
	print("\n")  // Use newline now
	return oList
}

func GetAzRoleAssignment(x map[string]interface{}) (y map[string]interface{}) {
	// Retrieve role assignment y from Azure if it exists and matches given x object's roleId, principalId, and scope
	// Above 3 parameters make role assignments unique

	// First, make sure x is a searchable role assignment object
	if x == nil { return nil }  // Don't look for empty objects

	xProps := x["properties"].(map[string]interface{})
	if xProps == nil { return nil }  // Return nil if properties missing

	xRoleDefinitionId := LastElem(StrVal(xProps["roleDefinitionId"]), "/")
	xPrincipalId      := StrVal(xProps["principalId"])
	xScope            := StrVal(xProps["scope"])
	if xScope == "" { xScope = StrVal(xProps["Scope"]) }  // Account for possibly capitalized key
	if xScope == "" || xPrincipalId == "" || xRoleDefinitionId == "" {
		return nil
	}

	// Get all role assignments for xPrincipalId under xScope
    params := map[string]string{
		"api-version": "2022-04-01",  // roleAssignments
		"$filter":     "principalId eq '" + xPrincipalId + "'",
	}
	url := az_url + xScope + "/providers/Microsoft.Authorization/roleAssignments"
	r := ApiGet(url, az_headers, params)
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
