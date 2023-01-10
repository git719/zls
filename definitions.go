// definitions.go

package main

import "strings"

import (
	"fmt"
	"path/filepath"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintRoleDefinition(x JsonObject, z aza.AzaBundle, oMap MapString) {
	// Print role definition object in YAML format
	if x == nil { return }
	if x["name"] != nil {
		fmt.Printf("id: %s\n", StrVal(x["name"]))
	}

	fmt.Printf("properties:\n")
	if x["properties"] == nil {
		fmt.Printf("  <Missing??>\n")
		return
	}
	xProp := x["properties"].(map[string]interface{})
	
	list := []string{"roleName", "description"}
	for _, i := range list {
		fmt.Printf("  %s %s\n", i+":", StrVal(xProp[i]))
	}

	fmt.Printf("  %-18s", "assignableScopes: ")
	if xProp["assignableScopes"] == nil {
		fmt.Printf("[]\n")
	} else {
		fmt.Printf("\n")
		scopes := xProp["assignableScopes"].([]interface{})
		subNames := GetIdNameMap("s", "", false, z, oMap) // Get all subscription id/names pairs
		if len(scopes) > 0 {
			for _, i := range scopes {
				if strings.HasPrefix(i.(string), "/subscriptions") {
					// Print subscription name as a comment at end of line
					subId := utl.LastElem(i.(string), "/")
					fmt.Printf("    - %s # %s\n", StrVal(i), subNames[subId])
				} else {
					fmt.Printf("    - %s\n", StrVal(i))
				}
			}
		} else {
			fmt.Printf("    <Not an arrays??>\n")
		}
	}

	fmt.Printf("  %-18s\n", "permissions:")
	if xProp["permissions"] == nil {
		fmt.Printf("    %s\n", "<No permissions??>")
	} else {
		permsSet := xProp["permissions"].([]interface{})
		if len(permsSet) == 1 {
			perms := permsSet[0].(map[string]interface{})    // Select the 1 expected single permission set

			fmt.Printf("    - actions:\n")      // Note that this one is different, as it starts the YAML array with the dash '-'
			if perms["actions"] != nil {
				permsA := perms["actions"].([]interface{})
				if utl.VarType(permsA)[0] != '[' {           // Open bracket character means it's an array list
					fmt.Printf("        <Not an array??>\n")
				} else {
					for _, i := range permsA {
						fmt.Printf("        - %s\n", StrVal(i))
					}
				}
			}

			fmt.Printf("      notActions:\n")
			if perms["notActions"] != nil {
				permsNA := perms["notActions"].([]interface{})
				if utl.VarType(permsNA)[0] != '[' {
					fmt.Printf("        <Not an array??>\n")
				} else {
					for _, i := range permsNA {
						fmt.Printf("        - %s\n", StrVal(i))
					}
				}
			}

			fmt.Printf("      dataActions:\n")
			if perms["dataActions"] != nil {
				permsDA := perms["dataActions"].([]interface{})
				if utl.VarType(permsDA)[0] != '[' {
					fmt.Printf("        <Not an array??>\n")
				} else {
					for _, i := range permsDA {
						fmt.Printf("        - %s\n", StrVal(i))
					}
				}
			}

			fmt.Printf("      notDataActions:\n")
			if perms["notDataActions"] != nil {
				permsNDA := perms["notDataActions"].([]interface{})
				if utl.VarType(permsNDA)[0] != '[' {
					fmt.Printf("        <Not an array??>\n")
				} else {
					for _, i := range permsNDA {
						fmt.Printf("        - %s\n", StrVal(i))
					}
				}
			}

		} else {
			fmt.Printf("    <More than one set??>\n")
		}
	}
}

// Notes on Azure RBAC role definitions and the API:
// As of api-version 2022-04-01, the RBAC API $filter=AtScopeAndBelow() does not appear to work as
// documented at https://learn.microsoft.com/en-us/azure/role-based-access-control/role-definitions-list.
// This means that anyone searching for a comprehensive list of ALL role definitions within an Azure tenant
// is forced to do this in a piecemeal, cumulative fashion. One must build a list of all scopes to search under,
// then proceed through each of those subscope within the hierarchy. This gradually builds a list of all
// BuiltIn and Custom definitions. The RBAC hierarchy is something like:
//   PATH               EXAMPLE
//   Tenant Root Group  /providers/Microsoft.Management/managementGroups/{tenantId}
//   Management Group   /providers/Microsoft.Management/managementGroups/{groupId1}
//   Subscription       /subscriptions/{subscriptionId1}
// 	 Resource Group     /subscriptions/{subscriptionId1}/resourceGroups/{myResourceGroup1}
// 	 Resource           /subscriptions/{subscriptionId1}/resourceGroups/{myResourceGroup1}/providers/Microsoft.Web/sites/mySite1
//
// Microsoft Azure BuiltIn roles are defined universally, at scope "/", so they are retrieved when the call
// is made to the Tenant Root Group scope. That means the bulk of calls is for Custom role types.
//
// The best practice to follow is to define ALL custom roles as universally as possible, at the highest
// scope, the Tenant Root Group scope. That way, they are always "visible" and therefore consumable
// anywhere within the tenant. That is essentially how Azure BuiltIn roles are defined, universally.
//
// Note that as of 2022-12-30, Microsoft Azure still has a limitation whereby any custom, role having
// any DataAction or NotDataAction CANNOT be defined at any MG scope level, and that prevents this
// good practice. Microsoft is actively working to lift this restriction, see: 
// https://learn.microsoft.com/en-us/azure/role-based-access-control/custom-roles
//
// There may be customers out there who at some point decided to define some of their custom roles within some
// of the hidden subscopes, and that's the reason why this utility follows this search algorithm to gather
// the full list of roles definitions. Note however, that this utility ONLY searches as deep as subscriptions,
// so if there are role definitions hidden within Resource Groups or individual resoures it may MISS them.

func GetRoleDefinitions(filter string, force, verbose bool, z aza.AzaBundle, oMap MapString) (list JsonArray) {
	// Get all roleDefinitions that match on provided filter. An empty "" filter means return
	// all of them. It always uses local cache if it's within the cache retention period. The
	// force boolean option will force a call to Azure.
	list = nil
	cachePeriod := int64(3660 * 24 * 7) // 1 week cache retention period 
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_roleDefinitions.json")
	cacheNoGood, list := CheckLocalCache(cacheFile, cachePeriod)
	if cacheNoGood || force {
		// Get all roleDefinitions in current Azure tenant again
		list = nil // We have to zero it out
		scopes := GetAzRbacScopes(z, oMap) // Get all RBAC hierarchy scopes to search for all role definitions 
		var uuids []string // Keep track of each unique objects to whittle out inherited repeats
		calls := 1 // Track number of API calls below
		params := aza.MapString{"api-version": "2022-04-01"}  // roleDefinitions
		for _, scope := range scopes {
			url := aza.ConstAzUrl + scope + "/providers/Microsoft.Authorization/roleDefinitions"
			r := ApiGet(url, z.AzHeaders, params)
			ApiErrorCheck(r, utl.Trace())
			if r["value"] != nil {
				definitions := r["value"].([]interface{})
				if verbose {
					// Using global var rUp to overwrite last line. Defer newline until done
					fmt.Printf("%s(API calls = %d) %d definitions in set %d", rUp, calls, len(definitions), calls)
				}
				for _, i := range definitions {
					x := i.(map[string]interface{})
					uuid := StrVal(x["name"])  // NOTE that 'name' key is the role definition Object Id
					if !utl.ItemInList(uuid, uuids) {
						// Add this role definition to growing list ONLY if it's NOT in it already
						list = append(list, x)
						uuids = append(uuids, uuid)
					}
				}
			}
			calls++
		}
		if verbose {
			fmt.Printf("\n") // Use newline now
		}
		utl.SaveFileJson(list, cacheFile) // Update the local cache
	}

	// Do filter matching
	if filter == "" {
		return list
	}
	var matchingList []interface{} = nil
	for _, i := range list { // Parse every object
		x := i.(map[string]interface{})
		// Match against relevant roleDefinitions attributes
		xProp := x["properties"].(map[string]interface{})
		if utl.SubString(StrVal(x["name"]), filter) || utl.SubString(StrVal(xProp["roleName"]), filter) ||
			utl.SubString(StrVal(x["description"]), filter) {
			matchingList = append(matchingList, x)
		}
	}
	return matchingList
}

func RoleDefinitionCountLocal(z aza.AzaBundle) (builtin, custom int64) {
	// Dedicated role definition local cache counter able to discern if role is custom to native tenant or it's an Azure BuilIn role
	var customList []interface{} = nil
	var builtinList []interface{} = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_roleDefinitions.json")
    if utl.FileUsable(cacheFile) {
		rawList, _:= utl.LoadFileJson(cacheFile)
		if rawList != nil {
			definitions := rawList.([]interface{})
			for _, i := range definitions {
				x := i.(map[string]interface{}) // Assert as JSON object type
				xProp := x["properties"].(map[string]interface{})
				if StrVal(xProp["type"]) == "CustomRole" {
					customList = append(customList, x)
				} else {
					builtinList = append(builtinList, x)
				}
			}			
			return int64(len(builtinList)), int64(len(customList))
		}
	}
	return 0, 0
}

func RoleDefinitionCountAzure(z aza.AzaBundle, oMap MapString) (builtin, custom int64) {
	// Dedicated role definition Azure counter able to discern if role is custom to native tenant or it's an Azure BuilIn role
	var customList []interface{} = nil
	var builtinList []interface{} = nil
	definitions := GetRoleDefinitions("", true, false, z, oMap) // true = force a call to Azure, false = be silent
	for _, i := range definitions {
		x := i.(map[string]interface{}) // Assert as JSON object type
		xProp := x["properties"].(map[string]interface{})
		if StrVal(xProp["type"]) == "CustomRole" {
			customList = append(customList, x)
		} else {
			builtinList = append(builtinList, x)
		}
	}			
	return int64(len(builtinList)), int64(len(customList))
}

func GetAzRoleDefinition(x JsonObject, z aza.AzaBundle) (y JsonObject) {
	// Special function to get RBAC role definition object from Azure if it exists
	// as defined by given x object, matching displayName and assignableScopes
    
	// First, make sure x is a searchable role definition object
	if x == nil { // Don't look for empty objects
		return nil
	}
	xProp := x["properties"].(map[string]interface{})
	if xProp == nil {
		return nil
	}
		
	xScopes := xProp["assignableScopes"].([]interface{})
	if utl.VarType(xScopes)[0] != '[' || len(xScopes) < 1 {
		return nil // Return nil if assignableScopes not an array, or it's empty
	}
	xRoleName := StrVal(xProp["roleName"])
	if xRoleName == "" {
		return nil
	}

	// Look for x under all its scopes
	for _, i := range xScopes {
		scope := StrVal(i)
		if scope == "/" { scope = "" } // Highly unlikely but just to avoid an err
		// Get all role assignments for xPrincipalId under xScope
		params := aza.MapString{
			"api-version": "2022-04-01",  // roleDefinitions
			"$filter":     "roleName eq '" + xRoleName + "'",
		}
		url := aza.ConstAzUrl + scope + "/providers/Microsoft.Authorization/roleDefinitions"
		r := ApiGet(url, z.AzHeaders, params)
		ApiErrorCheck(r, utl.Trace())
		if r != nil && r["value"] != nil {
			results := r["value"].([]interface{})
			if len(results) == 1 {
				y = results[0].(map[string]interface{}) // Select first index entry
				return y // We found it
			} else {
				return nil // If there's more than one entry we have other problems, so just return nil
			}
		}
	}
	return nil
}
