// definitions.go

package main

import "strings"

func PrintRoleDefinition(x map[string]interface{}) {
	// Print role definition object in YAML format
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
	
	list := []string{"roleName", "description"}
	for _, i := range list {
		print("  %s %s\n", i+":", StrVal(xProp[i]))
	}

	print("  %-18s", "assignableScopes: ")
	if xProp["assignableScopes"] == nil {
		print("[]\n")
	} else {
		print("\n")
		scopes := xProp["assignableScopes"].([]interface{})
		subNames := GetIdNameMap("s") // Get all subscription id/names pairs
		if len(scopes) > 0 {
			for _, i := range scopes {
				if strings.HasPrefix(i.(string), "/subscriptions") {
					// Print subscription name as a comment at end of line
					subId := LastElem(i.(string), "/")
					print("    - %s # %s\n", StrVal(i), subNames[subId])
				} else {
					print("    - %s\n", StrVal(i))
				}
			}
		} else {
			print("    <Not an arrays??>\n")
		}
	}

	print("  %-18s\n", "permissions:")
	if xProp["permissions"] == nil {
		print("    %s\n", "<No permissions??>")
	} else {
		permsSet := xProp["permissions"].([]interface{})
		if len(permsSet) == 1 {
			perms := permsSet[0].(map[string]interface{})    // Select the 1 expected single permission set

			print("    - actions:\n")      // Note that this one is different, as it starts the YAML array with the dash '-'
			if perms["actions"] != nil {
				permsA := perms["actions"].([]interface{})
				if VarType(permsA)[0] != '[' {           // Open bracket character means it's an array list
					print("        <Not an array??>\n")
				} else {
					for _, i := range permsA {
						print("        - %s\n", StrVal(i))
					}
				}
			}

			print("      notActions:\n")
			if perms["notActions"] != nil {
				permsNA := perms["notActions"].([]interface{})
				if VarType(permsNA)[0] != '[' {
					print("        <Not an array??>\n")
				} else {
					for _, i := range permsNA {
						print("        - %s\n", StrVal(i))
					}
				}
			}

			print("      dataActions:\n")
			if perms["dataActions"] != nil {
				permsDA := perms["dataActions"].([]interface{})
				if VarType(permsDA)[0] != '[' {
					print("        <Not an array??>\n")
				} else {
					for _, i := range permsDA {
						print("        - %s\n", StrVal(i))
					}
				}
			}

			print("      notDataActions:\n")
			if perms["notDataActions"] != nil {
				permsNDA := perms["notDataActions"].([]interface{})
				if VarType(permsNDA)[0] != '[' {
					print("        <Not an array??>\n")
				} else {
					for _, i := range permsNDA {
						print("        - %s\n", StrVal(i))
					}
				}
			}

		} else {
			print("    <More than one set??>\n")
		}
	}
}

func GetRoleDefinitions(mode string) (oList []interface{}) {
	// Get all role definitions that are available to use in current tenant 

	// As of api-version 2022-04-01, the RBAC API $filter=AtScopeAndBelow() does not appear to work as
    // documented at https://learn.microsoft.com/en-us/azure/role-based-access-control/role-definitions-list.
    // This means that anyone searching for a comprehensive list of ALL role definitions within an Azure tenant
    // is forced to do this in a piecemeal, cumulative fashion. One must build a list of scopes to search under,
	// then proceed through each of those subscope within the hierarchy. This gradually builds a list of all
    // BuiltIn and Custom definitions. The RBAC hierarchy is something like:
    
    //   PATH               EXAMPLE
    //   Tenant Root Group  /providers/Microsoft.Management/managementGroups/{tenantId}
    //   Management Group   /providers/Microsoft.Management/managementGroups/{groupId1}
    //   Subscription       /subscriptions/{subscriptionId1}
    // 	 Resource Group     /subscriptions/{subscriptionId1}/resourceGroups/{myResourceGroup1}
    // 	 Resource           /subscriptions/{subscriptionId1}/resourceGroups/{myResourceGroup1}/providers/Microsoft.Web/sites/mySite1
    
    // Microsoft Azure BuiltIn roles are defined universally, at scope "/", so they are retrieved when the call
    // is made to the Tenant Root Group scope. That means the bulk of calls is for Custom role types.
	
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
		"api-version": "2022-04-01",  // roleDefinitions
	}
	for _, scope := range scopes {
		subId := LastElem(scope, "/")
		subName := subMap[subId]
		url := az_url + scope + "/providers/Microsoft.Authorization/roleDefinitions"
		r := ApiGet(url, az_headers, params, false)
		if r["value"] != nil {
			definitions := r["value"].([]interface{}) // Assert as JSON array type
			if mode == "verbose" {
				print("\r(API calls = %d) %d definitions at '%s'", apiCalls, len(definitions), subName)
				PadSpaces(20)
			}
			for _, i := range definitions {
				x := i.(map[string]interface{})
				uuid := StrVal(x["name"])  // NOTE that 'name' key is the role definition UUID
				if !ItemInList(uuid, uuids) {
					// Add this role definition to growing list ONLY if it's NOT in it already
					oList = append(oList, x)
					uuids = append(uuids, uuid)
				}
			}
		}
		ApiErrorCheck(r, trace())
		apiCalls++
	}
	if mode == "verbose" {
		print("\n")
	}
	return oList
}

func GetAzRoleDefinition(x map[string]interface{}) (y map[string]interface{}) {
	// Retrieve role definition y from Azure if it exists and matches given x object's displayName and assignableScopes
    
	// First, make sure x is a searchable role definition object
	if x == nil {
		return nil  // Don't look for empty objects
	}
	xProps := x["properties"].(map[string]interface{})
	if xProps == nil {              // Return nil if properties missing
		return nil
	}
	xScopes := xProps["assignableScopes"].([]interface{})
	if VarType(xScopes)[0] != '[' || len(xScopes) < 1 {
		// Return nil if assignableScopes not an array, or it's empty
		return nil
	}
	xRoleName := StrVal(xProps["roleName"])
	if xRoleName == "" {
		return nil
	}

	// Include the root tenant scope anyway
	xScopes = append(xScopes, "/providers/Microsoft.Management/managementGroups/" + tenant_id)

	// Look for x under all its scopes
	for _, i := range xScopes {
		scope := StrVal(i)
		if scope == "/" {
			// Handle BuiltIn role types. This is unlikely to happen, but just in case
			scope = ""  
		}

		// Get all role assignments for xPrincipalId under xScope
		params := map[string]string{
			"api-version": "2022-04-01",  // roleDefinitions
			"$filter":     "roleName eq '" + xRoleName + "'",
		}
		url := az_url + scope + "/providers/Microsoft.Authorization/roleDefinitions"
		r := ApiGet(url, az_headers, params, false)
		if r != nil && r["value"] != nil {
			results := r["value"].([]interface{})
			if len(results) == 1 {
				y = results[0].(map[string]interface{})  // First entry
				return y    // We found it
			} else {
				// If there's more than one entry we have other problems, so just return nil
				return nil  
			}
		}
		ApiErrorCheck(r, trace())
	}
	return nil
}
