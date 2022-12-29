// adroles.go

package main

func PrintAdRole(x map[string]interface{}) {
	// Print object in YAML-like style format
	if x == nil { return }
	id := StrVal(x["id"])

	// Print the most important attributes first
	list := []string{"id", "displayName", "description", "roleTemplateId"}
	for _, i := range list {
		print("%-21s %s\n", i+":", StrVal(x[i]))
	}

	// Print members of this role
	r := ApiGet(mg_url+"/v1.0/directoryRoles/"+id+"/members", mg_headers, nil, false)
	if r["value"] != nil {
		members := r["value"].([]interface{}) // Assert as JSON array type
		if len(members) > 0 {
			print("members:\n")
			// PrintJson(members) // DEBUG
			for _, i := range members {
				m := i.(map[string]interface{}) // Assert as JSON object type
				Upn := StrVal(m["userPrincipalName"])
				Name := StrVal(m["displayName"])
				print("  %-40s %s\n", Upn, Name)
			}
		} else {
			print("%-28s %s\n", "members:", "None")
		}
	}
	ApiErrorCheck(r, trace())
}

func PrintAdRoleDef(x map[string]interface{}) {
	// Print object in YAML-like style format
	if x == nil { return }

	// Print the most important attributes first
	list := []string{"id", "displayName", "description", "isBuiltIn", "isEnabled"}
	for _, i := range list {
		print("%-21s %s\n", i+":", StrVal(x[i]))
	}

	// List permissions
	if x["rolePermissions"] != nil {
		rolePerms := x["rolePermissions"].([]interface{}) // Assert to JSON array
		if len(rolePerms) > 0 {
			// Unclear why rolePermissions is a list instead of the single entry that it usually is
			perms := rolePerms[0].(map[string]interface{}) // Assert JSON object
			if perms["allowedResourceActions"] != nil && len(perms["allowedResourceActions"].([]interface{})) > 0 {
				print("permissions:\n")
				for _, i := range perms["allowedResourceActions"].([]interface{}) {
					print("  %s\n", StrVal(i))
				}
			}
		} 
	}
}
