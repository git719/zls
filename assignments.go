// assignments.go

package main

import (
	"fmt"
	"log"
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
		pName := ""
		switch Type {
		case "User":
			pName = userMap[Pid]
		case "SP":
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

		fmt.Printf("\"%s\",\"%s\",\"%s\",\"%s\"\n", roleMap[Rid], pName, Type, Scope)
	}
}

func PrintRoleAssignment(x map[string]interface{}) {
	// Print role definition object in YAML-like style format
	if x["id"] == nil {
		return
	}

	fmt.Printf("id: %s\n", StrVal(x["name"]))

	xProp := x["properties"].(map[string]interface{})
	fmt.Printf("properties:\n")

	roleMap := GetIdNameMap("d")
	RoleId := LastElem(StrVal(xProp["roleDefinitionId"]), "/")
	fmt.Printf("  %-17s %s  # roleName = \"%s\"\n", "roleDefinitionId:", RoleId, roleMap[RoleId])

	var nameMap map[string]string
	pType := StrVal(xProp["principalType"])
	switch pType {
	case "User":
		nameMap = GetIdNameMap("u")
	case "ServicePrincipal":
		nameMap = GetIdNameMap("sp")
	case "Group":
		nameMap = GetIdNameMap("g")
	}
	pId := StrVal(xProp["principalId"])
	pName := nameMap[StrVal(xProp["principalId"])]
	fmt.Printf("  %-17s %s  # %s displayName = \"%s\"\n", "principalId:", pId, pType, pName)

	subMap := GetIdNameMap("s")
	scope := StrVal(xProp["scope"])
	if strings.HasPrefix(scope, "/subscriptions") {
		split := strings.Split(scope, "/")
		subName := subMap[split[2]]
		fmt.Printf("  %-17s %s  # Sub = %s\n", "scope:", scope, subName)
	} else {
		fmt.Printf("  %-17s %s\n", "scope:", scope)
	}
}

func GetRoleAssignments() (oList []interface{}) {
	// Get all role assigments from Azure
	// See https://docs.microsoft.com/en-us/rest/api/authorization/role-assignments/list

	// First, get all role assignments at Tenant root level
	oList = nil
	var uuids []string // Keep track of each unique role assignment ID
	apiCalls := 1      // Track number of API calls below

	url := az_url + "/providers/Microsoft.Management/managementGroups/" + tenant_id
	url += "/providers/Microsoft.Authorization/roleAssignments"
	//params := map[string]string{"filter": "atScopeAndBelow()"} // This has never worked
	r := APIGet(url, az_headers, nil, false)
	if r["value"] != nil {
		assignments := r["value"].([]interface{}) // Assert as JSON array type
		log.Printf("%d assignments at Tenant level\n", len(assignments))
		for _, obj := range assignments {
			x := obj.(map[string]interface{})
			uuid := StrVal(x["name"]) // NOTE that 'name' key is the role assignment UUID
			oList = append(oList, x)
			uuids = append(uuids, uuid)
		}
	}

	// Finally, get all role assignments at each Subscription level
	subMap := GetIdNameMap("s") // To print sub names, not IDs
	for _, i := range GetSubIds() {
		url = az_url + "/subscriptions/" + i + "/providers/Microsoft.Authorization/roleAssignments"
		r = APIGet(url, az_headers, nil, false)
		if r["value"] != nil {
			assignments := r["value"].([]interface{})
			fmt.Printf("\r(API calls = %d) %d assignments in sub '%s'", apiCalls, len(assignments), subMap[i])
			PadSpaces(20)
			for _, j := range assignments {
				x := j.(map[string]interface{}) // Assert as JSON object type
				uuid := StrVal(x["name"])
				if !ItemInList(uuid, uuids) {
					// Add this role assignment to growing list ONLY if it's NOT in it already (they repeat)
					oList = append(oList, x)
					uuids = append(uuids, uuid)
				}
			}
		}
		apiCalls++
	}
	fmt.Printf("\n")
	return oList
}
