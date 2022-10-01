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
	// Get all role definitions that are available to use in current tenant 

	// IMPORTANT: Presently (October 2022), the RBAC API $filter=AtScopeAndBelow() does NOT work as
    // documented at https://learn.microsoft.com/en-us/azure/role-based-access-control/role-definitions-list.
    // This means that anyone searching for a comprehensive list of ALL role definitions within an Azure tenant
    // is forced to do this in a piecemeal, cumulative fashion. The search must start at the Tenant Root Group
    // scope level, then proceed through EACH lower subscope within the hierarchy. This entails building a list
    // of all MGs at the root and searching for all roles DEFINED there; then building a list of all subscriptions
    // within each of those MGs and gettting all roles defined in them; then repeating for all resource groups
    // within each subscription, and so on. The hierarchy is something like:
    
    //   PATH               EXAMPLE
    //   Tenant Root Group  /providers/Microsoft.Management/managementGroups/{myTenantId}
    //   Management Group   /providers/Microsoft.Management/managementGroups/{groupId1}
    //   Subscription       /subscriptions/{subscriptionId1}
    // 	Resource Group     /subscriptions/{subscriptionId1}/resourceGroups/{myResourceGroup1}
    // 	Resource           /subscriptions/{subscriptionId1}/resourceGroups/{myResourceGroup1}/providers/Microsoft.Web/sites/mySite1
    
    // Note that because Microsoft Azure BUILT-IN roles are defined universally, at scope "/", they are all
    // gathered on the initial Tenant Root Group search. That means these comments are predominantly about CUSTOM
    // roles that a customer is likely to define within their respective tenant.
    
    // The best practice a customer can follow is to define ALL of their custom roles as universally as possible,
    // at the highest scope, the Tenant Root Group scope. That way, they are "visible" and therefore consumable
    // anywhere witin the tenant. As of this writing, Microsoft Azure still has a limitation whereby any custom
    // role having DataAction or NotDataAction CANNOT be defined at any MG scope level, and that prevents this
    // good practice. Microsoft is actively working to lift this restriction, see: 
    // https://learn.microsoft.com/en-us/azure/role-based-access-control/custom-roles

    // There may be customers out there who at some point decided to define some of their custom roles within some
    // of the hidden subscopes, and that's the reason why this utility follows this search algorithm to gather
    // the full list of roles definitions. Note however, that this utility ONLY searches as deep as subscriptions,
    // so if there are role definitions hidden within Resource Groups or individual resoures it will MISS them.

	// Step 1: Get all role definitions at the root tenant, including all Microsoft Azure Built-In-Roles
    oList = [] 
	oList = nil
	var uuids []string // Keep track of each unique role definition to whittle out repeats that come up in lower scopes
	url := az_url + "/providers/Microsoft.Management/managementGroups/" + tenant_id
	url += "/providers/Microsoft.Authorization/roleDefinitions"
	params := map[string]string{"filter": "atScopeAndBelow()"} // Again, this does not work as expected and documented
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

	// Step 2: Now get all CUSTOM roles under each subscription, ensuring uniqueness
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
