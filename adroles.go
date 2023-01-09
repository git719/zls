// adroles.go

package main

import (
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintAdRole(x JsonObject, z aza.AzaBundle) {
	// Print active Azure AD role object in YAML format
	if x == nil { return }
	id := StrVal(x["id"])

	// Print the most important attributes first
	list := []string{"id", "displayName", "description", "roleTemplateId"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" { print("%-21s %s\n", i+":", v) } // Only print non-null attributes
	}

	// Print members of this role
	url := aza.ConstMgUrl + "/v1.0/directoryRoles/" + id + "/members"
	r := ApiGet(url, z.MgHeaders, nil)
	if r["value"] != nil {
		members := r["value"].(JsonArray)
		if len(members) > 0 {
			print("members:\n")
			// PrintJson(members) // DEBUG
			for _, i := range members {
				m := i.(JsonObject)
				Upn := StrVal(m["userPrincipalName"])
				Name := StrVal(m["displayName"])
				print("  %-40s %s\n", Upn, Name)
			}
		} else {
			print("%-28s %s\n", "members:", "None")
		}
	}
	ApiErrorCheck(r, utl.Trace())
}

func PrintAdRoleDef(x JsonObject, z aza.AzaBundle) {
	// Print Azure AD role definition object in YAML format
	if x == nil { return }

	// Print the most important attributes first
	list := []string{"id", "displayName", "description", "isBuiltIn", "isEnabled"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" { print("%-21s %s\n", i+":", v) } // Only print non-null attributes
	}

	// List permissions
	if x["rolePermissions"] != nil {
		rolePerms := x["rolePermissions"].(JsonArray)
		if len(rolePerms) > 0 {
			// Unclear why rolePermissions is a list instead of the single entry that it usually is
			perms := rolePerms[0].(JsonObject)
			if perms["allowedResourceActions"] != nil && len(perms["allowedResourceActions"].(JsonArray)) > 0 {
				print("permissions:\n")
				for _, i := range perms["allowedResourceActions"].(JsonArray) {
					print("  %s\n", StrVal(i))
				}
			}
		} 
	}
}
