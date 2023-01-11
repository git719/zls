// adroles.go

package main

import (
	"fmt"
	"path/filepath"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintAdRole(x JsonObject, z aza.AzaBundle) {
	// Print active Azure AD role object in YAML-like format
	if x == nil { return }
	id := StrVal(x["id"])

	// Print the most important attributes first
	list := []string{"id", "displayName", "description", "roleTemplateId"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" { // Only print non-null attributes
			fmt.Printf("%-15s %s\n", i+":", v)
		 }
	}

	// Print members of this role
	url := aza.ConstMgUrl + "/v1.0/directoryRoles/" + id + "/members"
	r := ApiGet(url, z.MgHeaders, nil)
	ApiErrorCheck(r, utl.Trace())
	if r["value"] != nil {
		members := r["value"].([]interface{})
		if len(members) > 0 {
			fmt.Printf("members:\n")
			// PrintJson(members) // DEBUG
			for _, i := range members {
				m := i.(map[string]interface{})
				Upn := StrVal(m["userPrincipalName"])
				Name := StrVal(m["displayName"])
				fmt.Printf("  %-40s %s\n", Upn, Name)
			}
		} else {
			fmt.Printf("%-15s %s\n", "members:", "None")
		}
	}
}

func PrintAdRoleDef(x JsonObject) {
	// Print Azure AD role definition object in YAML-like format
	if x == nil { return }

	// Print the most important attributes first
	list := []string{"id", "displayName", "description", "isBuiltIn", "isEnabled"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" { // Only print non-null attributes
			fmt.Printf("%-12s %s\n", i+":", v)
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
}

func GetAdRoleDefs(filter string, force bool, z aza.AzaBundle) (list JsonArray) {
	// Get all adRoleDefs that match on provided filter. An empty "" filter means return
	// all of them. It always uses local cache if it's within the cache retention period.
	list = nil
	cachePeriod := int64(3660 * 24 * 7) // 1 week cache retention period 
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_adRoleDef.json")
	cacheNoGood, list := CheckLocalCache(cacheFile, cachePeriod)
	if cacheNoGood || force {
		list = nil // We have to zero it out
		// Get all AD Role Definitions in current Azure tenant again
		// Azure adRoleDefs are not so uniform and reside under 'roleManagement/directory', so
		// we're forced to process them differently than other MS Graph objects.
		url := aza.ConstMgUrl + "/v1.0/roleManagement/directory/roleDefinitions"
		r := ApiGet(url, z.MgHeaders, nil)
		ApiErrorCheck(r, utl.Trace())
		if r["value"] != nil {
			list = r["value"].([]interface{})
			utl.SaveFileJson(list, cacheFile) // Update the local cache
		}
	}
	
	// Do filter matching
	if filter == "" {
		return list
	}
	var matchingList JsonArray = nil
	for _, i := range list { // Parse every object
		x := i.(map[string]interface{})
		// Match against relevant adRoleDefs attributes
		if utl.SubString(StrVal(x["id"]), filter) || utl.SubString(StrVal(x["displayName"]), filter) ||
			utl.SubString(StrVal(x["description"]), filter) {
					matchingList = append(matchingList, x)
		}
	}
	return matchingList
}
