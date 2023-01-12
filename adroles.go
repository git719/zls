// adroles.go

package main

import (
	"fmt"
	"path/filepath"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintAdRole(x JsonObject, z aza.AzaBundle) {
	// Print Azure AD role definition object in YAML-like format
	if x == nil {
		return
	}
	// Print the most important attributes first
	list := []string{"id", "displayName", "description", "isBuiltIn", "isEnabled", "templateId"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" { // Only print non-null attributes
			fmt.Printf("%s: %s\n", i, v)
		}
	}
	// List permissions
	if x["rolePermissions"] != nil {
		rolePerms := x["rolePermissions"].([]interface{})
		if len(rolePerms) > 0 {
			// Unclear why rolePermissions is a list instead of the single entry that it usually is
			perms := rolePerms[0].(map[string]interface{})
			if perms["allowedResourceActions"] != nil && len(perms["allowedResourceActions"].([]interface{})) > 0 {
				fmt.Printf("permissions:\n")
				for _, i := range perms["allowedResourceActions"].([]interface{}) {
					fmt.Printf("  %s\n", StrVal(i))
				}
			}
		} 
	}
	// Print members of this role
	// See https://github.com/microsoftgraph/microsoft-graph-docs/blob/main/api-reference/v1.0/api/directoryrole-list-members.md
	url := aza.ConstMgUrl + "/v1.0/directoryRoles(roleTemplateId='" + StrVal(x["templateId"]) + "')/members"
	r := ApiGet(url, z.MgHeaders, nil)
	if r["value"] != nil {
		members := r["value"].([]interface{})
		if len(members) > 0 {
			fmt.Printf("members:\n")
			for _, i := range members {
				m := i.(map[string]interface{})
				fmt.Printf("  %s  %-40s   %s\n", StrVal(m["id"]), StrVal(m["userPrincipalName"]), StrVal(m["displayName"]))
			}
		} else {
			fmt.Printf("%s: %s\n", "members", "None")
		}
	} else {
		fmt.Printf("members: (Can't find members for this templateId)\n")
	}
}

func AdRolesCountLocal(z aza.AzaBundle) (int64) {
	// Return number of entries in local cache file
	var cachedList JsonArray = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_directoryRoles.json")
	if utl.FileUsable(cacheFile) {
		rawList, _ := utl.LoadFileJson(cacheFile)
		if rawList != nil {
			cachedList = rawList.([]interface{})
			return int64(len(cachedList))
		}
	}
	return 0
}	

func AdRolesCountAzure(z aza.AzaBundle) (int64) {
	// Return number of entries in Azure tenant.
	// Unfortunately, there is no API $count option, to so something like:
	//   z.MgHeaders["ConsistencyLevel"] = "eventual"
	//   url := aza.ConstMgUrl + "/v1.0/roleManagement/directory/roleDefinitions/$count"
	//   r := ApiGet(url, z.MgHeaders, nil)
	//   ApiErrorCheck(r, utl.Trace())
	//   if r["value"] != nil {
	// 	     return r["value"].(int64)
	//   }
	// So we'll just use the local count:
	return AdRolesCountLocal(z)
}

func GetAdRoles(filter string, force bool, z aza.AzaBundle) (list JsonArray) {
	// Get all Azure AD role definitions whose searchAttributes match on 'filter'. An empty "" filter returns all.
	// Uses local cache if it's less than 1hr old. The 'force' option forces calling Azure query.
	list = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_directoryRoles.json")
	cacheNoGood, list := CheckLocalCache(cacheFile, 3660) // cachePeriod = 1hr = 3600sec
	if cacheNoGood || force {
		list = GetAzAdRoles(cacheFile, z.MgHeaders, true) // Get all from Azure and show progress (verbose = true)
	}
	
	// Do filter matching
	if filter == "" {
		return list
	}
	var matchingList JsonArray = nil
	searchAttributes := []string{"id", "displayName", "description", "templateId"}
	var ids []string // Keep track of each unique objects to eliminate repeats
	for _, i := range list {
		x := i.(map[string]interface{})
		id := StrVal(x["id"])
		for _, i := range searchAttributes {
			if utl.SubString(StrVal(x[i]), filter) && !utl.ItemInList(id, ids) {
				matchingList = append(matchingList, x)
				ids = append(ids, id)
			}
		}
	}
	return matchingList	
}

func GetAzAdRoles(cacheFile string, headers aza.MapString, verbose bool) (list JsonArray) {
	// Get all Azure AD role definitions in current tenant AND save them to local cache file. Show progress if verbose = true.
	// See https://learn.microsoft.com/en-us/graph/api/rbacapplication-list-roledefinitions

	// There's not delta options for this, as it appears to be a relatively short list
	url := aza.ConstMgUrl + "/v1.0/roleManagement/directory/roleDefinitions"
	r := ApiGet(url, headers, nil)
	ApiErrorCheck(r, utl.Trace())
	if r["value"] != nil {
		list := r["value"].([]interface{})
		utl.SaveFileJson(list, cacheFile) // Update the local cache
	}
	return list
}

func GetAzAdRoleById(id string, headers aza.MapString) (list JsonObject) {
	// Get Azure AD role definition by UUID, with extended attributes
	// Note that role definitions are under a different area, until they are activated
	baseUrl := aza.ConstMgUrl + "/v1.0/roleManagement/directory/roleDefinitions"
	selection := "?$select=id,displayName,description,isBuiltIn,isEnabled,resourceScopes,"
	selection += "templateId,version,rolePermissions,inheritsPermissionsFrom"
	url := baseUrl + "/" + id + selection
	r := ApiGet(url, headers, nil)
	ApiErrorCheck(r, utl.Trace())
	return r
}
