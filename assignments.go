// assignments.go

package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintRoleAssignment(x JsonObject, z aza.AzaBundle, oMap map[string]string) {
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

	roleMap := GetIdNameMap("d", "", false, z, oMap) // Get all role id/names pairs
	roleId := utl.LastElem(StrVal(xProp["roleDefinitionId"]), "/")
	fmt.Printf("  %-17s %s  # roleName = \"%s\"\n", "roleDefinitionId:", roleId, roleMap[roleId])

	var nameMap map[string]string
	pType := StrVal(xProp["principalType"])
	switch pType {
	case "User":
		nameMap = GetIdNameMap("u", "", false, z, oMap) // Get all user id/names pairs
	case "Group":
		nameMap = GetIdNameMap("g", "", false, z, oMap) // Get all group id/names pairs
	case "ServicePrincipal":
		nameMap = GetIdNameMap("sp", "", false, z, oMap) // Get all SP id/names pairs
	default:
		pType = "not provided"
	}
	pId := StrVal(xProp["principalId"])
	pName := nameMap[StrVal(xProp["principalId"])]
	if pName == "" {
		pName = "???"
	}
	fmt.Printf("  %-17s %s  # principaltype = %s, displayName = \"%s\"\n", "principalId:", pId, pType, pName)

	subMap := GetIdNameMap("s", "", false, z, oMap) // Get all subscriptions id/names pairs
	scope := StrVal(xProp["scope"])
	if scope == "" { scope = StrVal(xProp["Scope"]) }  // Account for possibly capitalized key
	if strings.HasPrefix(scope, "/subscriptions") {
		split := strings.Split(scope, "/")
		subName := subMap[split[2]]
		fmt.Printf("  %-17s %s  # Sub = %s\n", "scope:", scope, subName)
	} else if scope == "/" {
		fmt.Printf("  %-17s %s  # Entire tenant\n", "scope:", scope)
	} else {
		fmt.Printf("  %-17s %s\n", "scope:", scope)
	}
}

func PrintRoleAssignmentReport(z aza.AzaBundle, oMap map[string]string)  {
	// Print a human-readable report of all role assignments
	roleMap := GetIdNameMap("d", "", false, z, oMap) // Get all role id/names pairs
	subMap := GetIdNameMap("s", "", false, z, oMap) // Get all subscriptions id/names pairs
	userMap := GetIdNameMap("u", "", false, z, oMap) // Get all user id/names pairs
	groupMap := GetIdNameMap("g", "", false, z, oMap) // Get all group id/names pairs
	spMap := GetIdNameMap("sp", "", false, z, oMap) // Get all SP id/names pairs
	
	assignments := GetAzRoleAssignments(false, z)
	for _, i := range assignments {
		x := i.(map[string]interface{})
		xProp := x["properties"].(map[string]interface{})
		Rid := utl.LastElem(StrVal(xProp["roleDefinitionId"]), "/")
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

		fmt.Printf("\"%s\",\"%s\",\"%s\",\"%s\"\n", roleMap[Rid], pName, Type, Scope)
	}
}

func GetRoleAssignments(filter string, force, verbose bool, z aza.AzaBundle, oMap map[string]string) (list JsonArray) {
	// Get all roleAssignments that match on provided filter. An empty "" filter means return
	// all of them. It always uses local cache if it's within the cache retention period. The
	// force boolean option will force a call to Azure.
	// See https://learn.microsoft.com/en-us/azure/role-based-access-control/role-assignments-list-rest
	list = nil
	cachePeriod := int64(3660 * 24 * 7) // 1 week cache retention period 
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_roleAssignments.json")
	cacheNoGood, list := CheckLocalCache(cacheFile, cachePeriod)
	if cacheNoGood || force {
		list = GetAzRoleAssignments(verbose, z) // Get the entire set from Azure
	}

	// Do filter matching
	if filter == "" {
		return list
	}
	var matchingList []interface{} = nil
	roleMap := GetIdNameMap("d", "", false, z, oMap) // Get all role definition id/names pairs (used later below)
	for _, i := range list { // Parse every object
		x := i.(map[string]interface{})
		// Match against relevant roleDefinitions attributes
		xProp := x["properties"].(map[string]interface{})
		rdId := StrVal(xProp["roleDefinitionId"])
		roleName := roleMap[utl.LastElem(rdId, "/")]
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

func GetAzRoleAssignments(verbose bool, z aza.AzaBundle) (list JsonArray) {
	// Get ALL roleAssignments in current Azure tenant AND save them to local cache file
	// Option to be verbose (true) or quiet (false), since it can take a while. 
	list = nil // We have to zero it out
	scopes := GetAzRbacScopes(z) // Get all RBAC hierarchy scopes to search for all role assignments 
	var uuids []string // Keep track of each unique objects to eliminate inherited repeats
	calls := 1 // Track number of API calls below
	params := aza.MapString{"api-version": "2022-04-01"}  // roleAssignments
	for _, scope := range scopes {
		url := aza.ConstAzUrl + scope + "/providers/Microsoft.Authorization/roleAssignments"
		r := ApiGet(url, z.AzHeaders, params)
		ApiErrorCheck(r, utl.Trace())
		if r["value"] != nil {
			assignments := r["value"].([]interface{})
			if verbose {
				// Using global var rUp to overwrite last line. Defer newline until done
				fmt.Printf("%s(API calls = %d) %d assignments in set %d", rUp, calls, len(assignments), calls)
			}
			for _, i := range assignments {
				x := i.(map[string]interface{})
				uuid := StrVal(x["name"])  // NOTE that 'name' key is the role assignment Object Id
				if !utl.ItemInList(uuid, uuids) {
					// Role assignments DO repeat! Add to growing list ONLY if it's NOT in it already
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
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_roleAssignments.json")
	utl.SaveFileJson(list, cacheFile) // Update the local cache
	return list
}

func GetAzRoleAssignment(x JsonObject, z aza.AzaBundle) (y JsonObject) {
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
    params := aza.MapString{
		"api-version": "2022-04-01",  // roleAssignments
		"$filter":     "principalId eq '" + xPrincipalId + "'",
	}
	url := aza.ConstAzUrl + xScope + "/providers/Microsoft.Authorization/roleAssignments"
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
