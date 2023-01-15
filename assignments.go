// assignments.go

package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"github.com/git719/maz"
	"github.com/git719/utl"
)

func PrintRoleAssignment(x map[string]interface{}, z maz.Bundle) {
	// Print role definition object in YAML-like
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

	roleNameMap := GetIdMapRoleDefs(z) // Get all role definition id:name pairs
	roleId := utl.LastElem(StrVal(xProp["roleDefinitionId"]), "/")
	fmt.Printf("  %-17s %s  # roleName = \"%s\"\n", "roleDefinitionId:", roleId, roleNameMap[roleId])

	var principalNameMap map[string]string = nil
	pType := StrVal(xProp["principalType"])
	switch pType {
	case "Group":
		principalNameMap = GetIdMapGroups(z) // Get all users id:name pairs
	case "User":
		principalNameMap = GetIdMapUsers(z) // Get all users id:name pairs
	case "ServicePrincipal":
		principalNameMap = GetIdMapSps(z) // Get all SPs id:name pairs
	default:
		pType = "not provided"
	}
	principalId := StrVal(xProp["principalId"])
	pName := principalNameMap[principalId]
	if pName == "" {
		pName = "???"
	}
	fmt.Printf("  %-17s %s  # principaltype = %s, displayName = \"%s\"\n", "principalId:", principalId, pType, pName)

	subNameMap := GetIdMapSubs(z) // Get all subscription id:name pairs
	scope := StrVal(xProp["scope"])
	if scope == "" { scope = StrVal(xProp["Scope"]) }  // Account for possibly capitalized key
	if strings.HasPrefix(scope, "/subscriptions") {
		split := strings.Split(scope, "/")
		subName := subNameMap[split[2]]
		fmt.Printf("  %-17s %s  # Sub = %s\n", "scope:", scope, subName)
	} else if scope == "/" {
		fmt.Printf("  %-17s %s  # Entire tenant\n", "scope:", scope)
	} else {
		fmt.Printf("  %-17s %s\n", "scope:", scope)
	}
}

func PrintRoleAssignmentReport(z maz.Bundle)  {
	// Print a human-readable report of all role assignments
	roleNameMap := GetIdMapRoleDefs(z) // Get all role definition id:name pairs
	subNameMap := GetIdMapSubs(z) // Get all subscription id:name pairs
	groupNameMap := GetIdMapGroups(z) // Get all users id:name pairs
	userNameMap := GetIdMapUsers(z) // Get all users id:name pairs
	spNameMap := GetIdMapSps(z) // Get all SPs id:name pairs
	
	assignments := GetAzRoleAssignments(false, z)
	for _, i := range assignments {
		x := i.(map[string]interface{})
		xProp := x["properties"].(map[string]interface{})
		Rid := utl.LastElem(StrVal(xProp["roleDefinitionId"]), "/")
		principalId := StrVal(xProp["principalId"])
		Type := StrVal(xProp["principalType"])
		pName := "ID-Not-Found"
		switch Type {
		case "Group":
			pName = groupNameMap[principalId]
		case "User":
			pName = userNameMap[principalId]
		case "ServicePrincipal":
			pName = spNameMap[principalId]
		}

		Scope := StrVal(xProp["scope"])
		if strings.HasPrefix(Scope, "/subscriptions") {
			// Replace sub ID to name
			split := strings.Split(Scope, "/")
			// Map subscription Id to its name + the rest of the resource path
			Scope = subNameMap[split[2]] + " " + strings.Join(split[3:], "/")
		}
		Scope = strings.TrimSpace(Scope)

		fmt.Printf("\"%s\",\"%s\",\"%s\",\"%s\"\n", roleNameMap[Rid], pName, Type, Scope)
	}
}

func RoleAssignmentsCountLocal(z maz.Bundle) (int64) {
	var cachedList []interface{} = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_roleAssignments.json")
    if utl.FileUsable(cacheFile) {
		rawList, _ := utl.LoadFileJson(cacheFile)
		if rawList != nil {
			cachedList = rawList.([]interface{})
			return int64(len(cachedList))
		}
	}
	return 0
}

func RoleAssignmentsCountAzure(z maz.Bundle) (int64) {
	list := GetAzRoleAssignments(false, z) // false = quiet
	return int64(len(list))
}

func GetRoleAssignments(filter string, force bool, z maz.Bundle) (list []interface{}) {
	// Get all roleAssignments that match on provided filter. An empty "" filter means return
	// all of them. It always uses local cache if it's within the cache retention period. The
	// force boolean option will force a call to Azure.
	// See https://learn.microsoft.com/en-us/azure/role-based-access-control/role-assignments-list-rest
	list = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_roleAssignments.json")
	cacheNoGood, list := CheckLocalCache(cacheFile, 604800) // cachePeriod = 1 week in seconds
	if cacheNoGood || force {
		list = GetAzRoleAssignments(true, z) // Get the entire set from Azure, true = show progress
	}

	// Do filter matching
	if filter == "" {
		return list
	}
	var matchingList []interface{} = nil
	roleNameMap := GetIdMapRoleDefs(z) // Get all role definition id:name pairs
	for _, i := range list { // Parse every object
		x := i.(map[string]interface{})
		// Match against relevant roleDefinitions attributes
		xProp := x["properties"].(map[string]interface{})
		rdId := StrVal(xProp["roleDefinitionId"])
		roleName := roleNameMap[utl.LastElem(rdId, "/")]
		principalId := StrVal(xProp["principalId"])
		description := StrVal(xProp["description"])
		principalType := StrVal(xProp["principalType"])
		scope := StrVal(xProp["scope"])
		if utl.SubString(StrVal(x["name"]), filter) || utl.SubString(rdId, filter) ||
			utl.SubString(roleName, filter) || utl.SubString(principalId, filter) ||
			utl.SubString(description, filter) || utl.SubString(principalType, filter) ||
				utl.SubString(scope, filter) {
					matchingList = append(matchingList, x)
		}
	}
	return matchingList
}

func GetAzRoleAssignments(verbose bool, z maz.Bundle) (list []interface{}) {
	// Get ALL roleAssignments in current Azure tenant AND save them to local cache file
	// Option to be verbose (true) or quiet (false), since it can take a while. 
	// See https://learn.microsoft.com/en-us/rest/api/authorization/role-assignments/list-for-subscription
    // 1st, we look for all tenant-level role assignments
	list = nil // We have to zero it out
	var assignmentIds []string // Keep track of each unique object to eliminate inherited repeats
	k := 1 // Track number of API calls to provide progress
	params := map[string]string{"api-version": "2022-04-01"}  // roleAssignments
	params["$filter"] = "atScope()" // Needed for this scope only
	url := maz.ConstAzUrl + "/providers/Microsoft.Authorization/roleAssignments"
	r := ApiGet(url, z.AzHeaders, params)
	ApiErrorCheck(r, utl.Trace())
	if r != nil && r["value"] != nil {
		tenantLevelAssignments := r["value"].([]interface{})
		for _, i := range tenantLevelAssignments {
			x := i.(map[string]interface{})
			// Append to growing list, keeping track of all assignmentIds
			list = append(list, x)
			assignmentIds = append(assignmentIds, StrVal(x["name"])) // Note, name is the object UUID
		}
		if verbose { // Using global var rUp to overwrite last line. Defer newline until done
			fmt.Printf("%s(API calls = %d) %d role assignments under root scope", rUp, k, len(tenantLevelAssignments))
		}
		k++
	}
	// 2nd, we look under subscription-level role assignments 
	subscriptionScopes := GetAzSubscriptionsIds(z)
	subNameMap := GetIdMapSubs(z) // Get all subscription id:name pairs
	delete(params, "$filter") // Removing, so we can pull all assignments under each subscription
	for _, scope := range subscriptionScopes {
		subName :=  subNameMap[utl.LastElem(scope, "/")] // Get subscription name
		url = maz.ConstAzUrl + scope + "/providers/Microsoft.Authorization/roleAssignments"
		r = ApiGet(url, z.AzHeaders, params)
		ApiErrorCheck(r, utl.Trace())
		if r != nil && r["value"] != nil {
			assignmentsInThisSubscription := r["value"].([]interface{})
			u := 0 // Count assignments in this subscription
			for _, i := range assignmentsInThisSubscription {
				x := i.(map[string]interface{})
				id := StrVal(x["name"])
				if utl.ItemInList(id, assignmentIds) {
					continue // Skip repeats
				}
				list = append(list, x) // This one is unique, append to growing list 
				assignmentIds = append(assignmentIds, id) // Keep track of the Id
				u++
			}
			if verbose {
				fmt.Printf("%s(API calls = %d) %d role assignments under subscription %s", rUp, k, u, subName)
			}
			k++
		}
	}
	if verbose {
		fmt.Printf("\n") // Use newline now
	}
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_roleAssignments.json")
	utl.SaveFileJson(list, cacheFile) // Update the local cache
	return list
}

// func GetAzRoleAssignments(verbose bool, z maz.Bundle) (list []interface{}) {
// 	// Get ALL roleAssignments in current Azure tenant AND save them to local cache file
// 	// Option to be verbose (true) or quiet (false), since it can take a while. 
// 	list = nil // We have to zero it out
// 	scopes := GetAzRbacScopes(z) // Get all RBAC hierarchy scopes to search for all role assignments 
// 	var uuids []string // Keep track of each unique objects to eliminate inherited repeats
// 	k := 1 // Track number of API calls to provide progress
// 	params := map[string]string{"api-version": "2022-04-01"}  // roleAssignments
// 	for _, scope := range scopes {
// 		url := maz.ConstAzUrl + scope + "/providers/Microsoft.Authorization/roleAssignments"
// 		r := ApiGet(url, z.AzHeaders, params)
// 		ApiErrorCheck(r, utl.Trace())
// 		if r["value"] != nil {
// 			assignments := r["value"].([]interface{})
// 			u := 0 // Keep track of assignments in this scope
// 			for _, i := range assignments {
// 				x := i.(map[string]interface{})
// 				uuid := StrVal(x["name"])  // NOTE that 'name' key is the role assignment Object Id
// 				if utl.ItemInList(uuid, uuids) {
// 					continue // Role assignments DO repeat! Skip if it's already been added.
// 				}
// 				list = append(list, x)
// 				uuids = append(uuids, uuid)
// 				u++
// 			}
// 			// if verbose { // Using global var rUp to overwrite last line. Defer newline until done
// 			// 	fmt.Printf("%s(API calls = %d) %d assignments in this set", rUp, k, u)
// 			// }
// 			fmt.Printf("(API calls = %d) %d assignments in this set (%s)\n", k, u, scope)
// 		}
// 		k++
// 	}
// 	if verbose {
// 		fmt.Printf("\n") // Use newline now
// 	}
// 	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_roleAssignments.json")
// 	utl.SaveFileJson(list, cacheFile) // Update the local cache
// 	return list
// }

func GetAzRoleAssignment(x map[string]interface{}, z maz.Bundle) (y map[string]interface{}) {
	// Special function to get RBAC role assignment object from Azure if it exists
	// as defined by given x object, matching roleId, principalId, and scope (3 parameters
	// which make the role assignment unique)

	// First, make sure x is a searchable role assignment object
	if x == nil {
		return nil
	}
	xProp := x["properties"].(map[string]interface{})
	if xProp == nil {
		return nil
	}

	xRoleDefinitionId := utl.LastElem(StrVal(xProp["roleDefinitionId"]), "/")
	xPrincipalId      := StrVal(xProp["principalId"])
	xScope            := StrVal(xProp["scope"])
	if xScope == "" {
		xScope = StrVal(xProp["Scope"]) // Account for possibly capitalized key
	}
	if xScope == "" || xPrincipalId == "" || xRoleDefinitionId == "" {
		return nil
	}

	// Get all role assignments for xPrincipalId under xScope
    params := map[string]string{
		"api-version": "2022-04-01",  // roleAssignments
		"$filter":     "principalId eq '" + xPrincipalId + "'",
	}
	url := maz.ConstAzUrl + xScope + "/providers/Microsoft.Authorization/roleAssignments"
	r := ApiGet(url, z.AzHeaders, params)
	if r != nil && r["value"] != nil {
		results := r["value"].([]interface{})
		for _, i := range results {
			y = i.(map[string]interface{})
			yProp := y["properties"].(map[string]interface{})
			yScope := StrVal(yProp["scope"])
			yRoleDefinitionId := utl.LastElem(StrVal(yProp["roleDefinitionId"]), "/")
			if yScope == xScope && yRoleDefinitionId == xRoleDefinitionId {
				return y  // We found it
			}
		}
	}
	ApiErrorCheck(r, utl.Trace())
	return nil
}

func GetAzRoleAssignmentById(id string, z maz.Bundle) (map[string]interface{}) {
	// Get Azure resource roleAssignment by Object Id
	// Unfortunately we have to traverse and search the entire Azure resource scope hierarchy

	// 1st, we look for all tenant-level role assignments
	params := map[string]string{"api-version": "2022-04-01"}  // roleAssignments
	params["$filter"] = "AtScope()" // Needed for this scope only
	url := maz.ConstAzUrl + "/providers/Microsoft.Authorization/roleAssignments"
	r := ApiGet(url, z.AzHeaders, params)
	ApiErrorCheck(r, utl.Trace())
	if r != nil && r["value"] != nil {
		tenantLevelAssignments := r["value"].([]interface{})
		for _, i := range tenantLevelAssignments {
			x := i.(map[string]interface{})
			if StrVal(x["name"]) == id {
				return x // Return immediately if found
			}
		}
	}
	// 2nd, we look under subscription-level role assignments 
	subscriptionScopes := GetAzSubscriptionsIds(z)
	delete(params, "$filter") // Removing, so we can pull all assignments under each subscription
	for _, scope := range subscriptionScopes {
		url = maz.ConstAzUrl + scope + "/providers/Microsoft.Authorization/roleAssignments"
		r = ApiGet(url, z.AzHeaders, params)
		ApiErrorCheck(r, utl.Trace())
		if r != nil && r["value"] != nil {
			assignmentsInThisSubscription := r["value"].([]interface{})
			for _, i := range assignmentsInThisSubscription {
				x := i.(map[string]interface{})
				if StrVal(x["name"]) == id {
					return x // Return immediately if found
				}
			}
		}
	}
	// BELOW DOESN'T WORK. WOULD NEED TO have GetAzRbacScopes() bring back EVERY SINGLE scope across
	// the environment, which is not efficient.
	// scopes := GetAzRbacScopes(z) // Get all RBAC hierarchy scopes to search for all role assignments
	// scopes = append(scopes, "/") // This covers those at the root of the tenant
	// params := map[string]string{"api-version": "2022-04-01"}  // roleAssignments
	// for _, scope := range scopes {		
	// 	url := maz.ConstAzUrl + scope + "/providers/Microsoft.Authorization/roleAssignments/" + id
	// 	r := ApiGet(url, z.AzHeaders, params)
	// 	if r != nil && r["name"] != nil && StrVal(r["name"]) == id {
	// 		return r
	// 	}
	// }
	return nil
}
