// apps.go

package main

func PrintApp(obj map[string]interface{}) {
	// Print application object in YAML-like style format
	if obj["id"] == nil {
		return
	}
	id := StrVal(obj["id"])

	// Print the most important attributes first
	list := []string{"displayName", "appId", "id"}
	for _, i := range list {
		print("%-21s %s\n", i+":", StrVal(obj[i]))
	}

	// Print owners
	r := APIGet(mg_url+"/beta/applications/"+id+"/owners", mg_headers, nil, false)
	if r["value"] != nil {
		owners := r["value"].([]interface{}) // Assert as JSON array type
		if len(owners) > 0 {
			print("owners:\n")
			// PrintJSON(groups) // DEBUG
			for _, i := range owners {
				o := i.(map[string]interface{}) // Assert as JSON object type
				Type, Name := "???", "???"
				Type = LastElem(StrVal(o["@odata.type"]), ".")
				switch Type {
				case "user":
					Name = StrVal(o["userPrincipalName"])
				case "group":
					Name = StrVal(o["displayName"])
				case "servicePrincipal":
					Name = StrVal(o["servicePrincipalType"])
				default:
					Name = "???"
				}
				print("  %-50s %s (%s)\n", Name, StrVal(o["id"]), Type)
			}
		} else {
			print("%-21s %s\n", "owners:", "None")
		}
	}

	// Print all groups/roles it is a member of
	memberOf := GetObjectMemberOfs("ap", id) // For this App object
	PrintMemberOfs("ap", memberOf)

	// Print API permissions
	// Just look under this object's 'requiredResourceAccess' attribute
	if obj["requiredResourceAccess"] != nil && len(obj["requiredResourceAccess"].([]interface{})) > 0 {
		print("api_permissions:\n")
		APIs := obj["requiredResourceAccess"].([]interface{}) // Assert to JSON array
		for _, a := range APIs {
			api := a.(map[string]interface{}) // Assert as JSON object type
			// Getting this API's name and permission value such as Directory.Read.All is a 2-step process:
			// 1) Get all the roles for given API and put their id/value pairs in a map, then
			// 2) Use that map to enumerate and print them

			// Let's drill down into the permissions for this API
			if api["resourceAppId"] == nil {
				print("  %-50s %s\n", "Unknown API", "Missing resourceAppId")
				continue // Skip this API, move on to next one
			}

			// Let's drill down into the permissions for this API
			resAppId := StrVal(api["resourceAppId"])

			// Get this API's SP object with all relevant attributes
			r := APIGet(mg_url+"/beta/servicePrincipals?filter=appId+eq+'"+resAppId+"'", mg_headers, nil, false)
			// Unclear why result is a list instead of a single entry

			if r["value"] == nil {
				print("  %-50s %s\n", resAppId, "Unable to get Resource App object. Skipping this API.")
				continue
			}

			SPs := r["value"].([]interface{})

			if len(SPs) > 1 {
				print("  %-50s %s\n", resAppId, "Error. Multiple SPs for this AppId. Aborting.")
				exit(1)
			}

			sp := SPs[0].(map[string]interface{}) // The only expected entry, asserted as JSON object type

			// 1. Put all API role id:name pairs into roleMap list
			roleMap := make(map[string]string)
			if sp["appRoles"] != nil {
				for _, i := range sp["appRoles"].([]interface{}) { // Iterate through all roles
					// These are for Application types
					role := i.(map[string]interface{}) // Assert JSON object type
					if role["id"] != nil && role["value"] != nil {
						roleMap[StrVal(role["id"])] = StrVal(role["value"]) // Add entry to map
					}
				}
			}
			if sp["publishedPermissionScopes"] != nil {
				for _, i := range sp["publishedPermissionScopes"].([]interface{}) {
					// These are for Delegated types
					role := i.(map[string]interface{})
					if role["id"] != nil && role["value"] != nil {
						roleMap[StrVal(role["id"])] = StrVal(role["value"])
					}
				}
			}
			if roleMap == nil {
				print("  %-50s %s\n", resAppId, "Error getting list of appRoles.")
				continue
			}

			// 2. Parse this app permissions, and use roleMap to display permission value
			if api["resourceAccess"] != nil && len(api["resourceAccess"].([]interface{})) > 0 {
				Perms := api["resourceAccess"].([]interface{})
				apiName := StrVal(sp["displayName"]) // This API's name
				for _, i := range Perms {            // Iterate through perms
					perm := i.(map[string]interface{})
					pid := StrVal(perm["id"]) // JSON string
					print("  %-50s %s\n", apiName, roleMap[pid])
				}
			} else {
				print("  %-50s %s\n", resAppId, "Error getting list of appRoles.")
			}
		}
	}
}
