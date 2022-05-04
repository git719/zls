// definitions.go

package main

import (
	"fmt"
	"strings"
)

func PrintRoleDefinition(x map[string]interface{}) {
	// Print role definition object in YAML-like style format
	if x["id"] == nil {
		return
	}

	fmt.Printf("id: %s\n", StrVal(x["name"]))
	xProp := x["properties"].(map[string]interface{})
	fmt.Printf("properties:\n")
	list := []string{"roleName", "description"}
	for _, i := range list {
		fmt.Printf("  %s %s\n", i+":", StrVal(xProp[i]))
	}

	scopes := xProp["assignableScopes"].([]interface{})
	if len(scopes) > 0 {
		fmt.Printf("  %-18s \n", "assignableScopes:")
		subNames := GetIdNameMap("s") // Get all subscription id/names pairs
		for _, i := range scopes {
			if strings.HasPrefix(i.(string), "/subscriptions") {
				// If scope is a subscription also print its name as a comment at end of line
				subId := LastElem(i.(string), "/")
				fmt.Printf("    - %s # %s\n", StrVal(i), subNames[subId])
			} else {
				fmt.Printf("    - %s\n", StrVal(i))
			}
		}
	} else {
		fmt.Printf("  %-18s %s\n", "assignableScopes:", "[]")
	}

	permsSet := xProp["permissions"].([]interface{})
	if len(permsSet) == 1 {
		fmt.Printf("  %-18s \n", "permissions:")
		perms := permsSet[0].(map[string]interface{})
		permsA := perms["actions"].([]interface{})
		if len(permsA) > 0 {
			fmt.Printf("    %-16s \n", "actions:")
			for _, i := range permsA {
				fmt.Printf("      - %s\n", StrVal(i))
			}
		}
		permsDA := perms["dataActions"].([]interface{})
		if len(permsDA) > 0 {
			fmt.Printf("    %-16s \n", "dataActions:")
			for _, i := range permsDA {
				fmt.Printf("      - %s\n", StrVal(i))
			}
		}
		permsNA := perms["notActions"].([]interface{})
		if len(permsNA) > 0 {
			fmt.Printf("    %-16s \n", "notActions:")
			for _, i := range permsNA {
				fmt.Printf("      - %s\n", StrVal(i))
			}
		}
		permsNDA := perms["notDataActions"].([]interface{})
		if len(permsNDA) > 0 {
			fmt.Printf("    %-16s \n", "notDataActions:")
			for _, i := range permsNDA {
				fmt.Printf("      - %s\n", StrVal(i))
			}
		}
	} else if len(permsSet) > 1 {
		fmt.Printf("%-20s %s\n", "permissions:", "ERROR. More than one set??")
	} else {
		fmt.Printf("%-20s %s\n", "permissions:", "[]")
	}
}

func GetRoleDefinitions() (oList []interface{}) {
	// Get all role definitions from Azure.
	// Ref: https://docs.microsoft.com/en-us/azure/role-based-access-control/role-definitions-list
	oList = nil
	var uuids []string // Keep track of each unique role definition to whittle out repeats that come up in lower scopes
	url := az_url + "/providers/Microsoft.Management/managementGroups/" + tenant_id
	url += "/providers/Microsoft.Authorization/roleDefinitions"
	params := map[string]string{"filter": "atScopeAndBelow()"} // Get all definitions in the entire tenant
	r := APIGet(url, az_headers, params, false)
	if r["value"] != nil {
		definitions := r["value"].([]interface{}) // Assert as JSON array type
		for _, i := range definitions {
			x := i.(map[string]interface{}) // Assert as JSON object type
			uuid := StrVal(x["name"])       // NOTE that 'name' key is the role definition UUID
			oList = append(oList, x)
			uuids = append(uuids, uuid)
		}
	}

	// STILL cannot seem to get all tenant role defs using atScopeAndBelow, so force to re-add below :-(

	// Step2: Now get all CUSTOM role definitions under each subscription, using the uuids list to ensure uniqueness
	subMap := GetIdNameMap("s") // To print sub names, not IDs
	apiCalls := 1               // Track number of API calls below
	for _, i := range GetSubIds() {
		r = APIGet(az_url+"/subscriptions/"+i+"/providers/Microsoft.Authorization/roleDefinitions?$filter=type+eq+'CustomRole'", az_headers, nil, false)
		if r["value"] != nil {
			definitions := r["value"].([]interface{})
			fmt.Printf("\r(API calls = %d) %d definitions in sub '%s'", apiCalls, len(definitions), subMap[i])
			PadSpaces(20)
			for _, j := range r["value"].([]interface{}) {
				x := j.(map[string]interface{})
				uuid := StrVal(x["name"])
				if !ItemInList(uuid, uuids) { // Add this role definition to growing list ONLY if it's NOT in it already
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
