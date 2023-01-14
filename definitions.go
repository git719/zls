// definitions.go

package main

import (
	"fmt"
	"strings"
	"path/filepath"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintRoleDefinition(x map[string]interface{}, z aza.AzaBundle, oMap map[string]string) {
	// Print role definition object in YAML-like format
	if x == nil {
		return
	}
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

func GetRoleDefinitions(filter string, force, verbose bool, z aza.AzaBundle, oMap map[string]string) (list []interface{}) {
	// Get all roleDefinitions that match on provided filter. An empty "" filter means return all of them.
	// It always uses local cache if it's within the cache retention period. The force boolean option forces
	// a call to Azure. The verbose option details the progress. 
	list = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_roleDefinitions.json")
	cacheNoGood, list := CheckLocalCache(cacheFile, 604800) // cachePeriod = 1 week in seconds
	if cacheNoGood || force {
		list = GetAzRoleDefinitions(verbose, z) // Get the entire set from Azure
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

func GetAzRoleDefinitions(verbose bool, z aza.AzaBundle) (list []interface{}) {
	// Get ALL roleDefinitions in current Azure tenant AND save them to local cache file
	// Option to be verbose (true) or quiet (false), since it can take a while.
	// See https://learn.microsoft.com/en-us/rest/api/authorization/role-definitions/list
    // 1st, we look for all tenant-level role definitions (this includes all universal BuiltIn ones)
	list = nil // We have to zero it out
	var definitionIds []string // Keep track of each unique object to eliminate inherited repeats
	k := 1 // Track number of API calls to provide progress
	params := aza.MapString{"api-version": "2022-04-01"}  // roleDefinitions
	url := aza.ConstAzUrl + "/providers/Microsoft.Authorization/roleDefinitions"
	r := ApiGet(url, z.AzHeaders, params)
	ApiErrorCheck(r, utl.Trace())
	if r != nil && r["value"] != nil {
		tenantLevelDefinitions := r["value"].([]interface{})
		for _, i := range tenantLevelDefinitions {
			x := i.(map[string]interface{})
			// Append to growing list, keeping track of all definitionIds
			list = append(list, x)
			definitionIds = append(definitionIds, StrVal(x["name"])) // Note, name is the object UUID
		}
		if verbose { // Using global var rUp to overwrite last line. Defer newline until done
			fmt.Printf("%s(API calls = %d) %d unique definitions in this set", rUp, k, len(tenantLevelDefinitions))
		}
		k++
	}
	// 2nd, we look under subscription-level role definitions 
	subscriptionScopes := GetAzSubscriptionsIds(z)
	for _, scope := range subscriptionScopes {
		url = aza.ConstAzUrl + scope + "/providers/Microsoft.Authorization/roleDefinitions"
		r = ApiGet(url, z.AzHeaders, params)
		ApiErrorCheck(r, utl.Trace())
		if r != nil && r["value"] != nil {
			definitionsInThisSubscription := r["value"].([]interface{})
			u := 0 // Count unique definitions in this subscription
			for _, i := range definitionsInThisSubscription {
				x := i.(map[string]interface{})
				id := StrVal(x["name"])
				if utl.ItemInList(id, definitionIds) {
					continue // Skip repeats
				}
				list = append(list, x) // This one is unique, append to growing list 
				definitionIds = append(definitionIds, id) // Keep track of the Id
				u++
			}
			if verbose {
				fmt.Printf("%s(API calls = %d) %d unique definitions in this set", rUp, k, u)
			}
			k++
		}
	}
	if verbose {
		fmt.Printf("\n") // Use newline now
	}
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_roleDefinitions.json")
	utl.SaveFileJson(list, cacheFile) // Update the local cache
	return list

	// Important Azure resource RBAC role definitions API note: As of api-version 2022-04-01, the filter
	// AtScopeAndBelow() does not work as documented at:
	// https://learn.microsoft.com/en-us/azure/role-based-access-control/role-definitions-list.
	
	// This means that anyone searching for a comprehensive list of ALL role definitions within an Azure tenant
	// is forced to do so in 2 steps: 1) Getting all role definitions at the tenant level. This grabs
	// all universal Azure BuiltIn role definitions, as well as any custom ones also defined at that level.
	// 2) Then getting all the role definitions defined under each subscription scope. These are of course
	// all custom role definitions.

	// The best practice for teams to follow is to define ALL its custom roles as universally as possible, at
	// the highest scope, the Tenant Root Group scope. That way they are "visible" and therefore consumable
	// anywhere within the tenant. That is essentially how Azure BuiltIn roles are defined, but even more universally.

	// Note also that as of 2022-12-30, Microsoft Azure still has a limitation whereby any custom, role having
	// any DataAction or NotDataAction CANNOT be defined at any MG scope level, and that prevents this
	// good practice. Microsoft is actively working to lift this restriction, see: 
	// https://learn.microsoft.com/en-us/azure/role-based-access-control/custom-roles
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

func RoleDefinitionCountAzure(z aza.AzaBundle, oMap map[string]string) (builtin, custom int64) {
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

func GetAzRoleDefinition(x map[string]interface{}, z aza.AzaBundle) (y map[string]interface{}) {
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

func GetAzRoleDefinitionById(id string, z aza.AzaBundle) (map[string]interface{}) {
	// Get Azure resource roleDefinition by Object Id
	// See https://learn.microsoft.com/en-us/rest/api/authorization/role-definitions/get-by-id
    // 1st, we look for this "id" under tenant level role definitions
	params := aza.MapString{"api-version": "2022-04-01"}  // roleDefinitions
	url := aza.ConstAzUrl + "/providers/Microsoft.Authorization/roleDefinitions/" + id
	r := ApiGet(url, z.AzHeaders, params)
	ApiErrorCheck(r, utl.Trace())
	if r != nil && r["name"] != nil && StrVal(r["name"]) == id {
		return r // Return immediately as found
	}
    // 2nd, we look under subscription level role definitions 
	scopes := GetAzSubscriptionsIds(z)
	for _, scope := range scopes {
		url = aza.ConstAzUrl + scope + "/providers/Microsoft.Authorization/roleDefinitions/" + id
		r = ApiGet(url, z.AzHeaders, params)
		ApiErrorCheck(r, utl.Trace())
		if r != nil && r["name"] != nil && StrVal(r["name"]) == id {
			return r // Return immediately as found
		}
	}
	return nil
}
