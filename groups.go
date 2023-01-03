// groups.go

package main

func PrintGroup(x map[string]interface{}) {
	// Print group object in YAML
	if x == nil { return }
	id := StrVal(x["id"])

	// Print the most important attributes
	list := []string{"displayName", "description", "id", "isAssignableToRole", "mailEnabled", "mailNickname"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" { print("%-21s %s\n", i+":", v) } // Only print non-null attributes
	}

	// OPTIONAL: Print other attributes here

	url := mg_url + "/beta/groups/" + id + "/owners"
	r := ApiGet(url, mg_headers, nil)
	if r["value"] != nil {
		owners := r["value"].([]interface{}) // Assert as JSON array type
		if len(owners) > 0 {
			print("owners:\n")
			for _, i := range owners {
				o := i.(map[string]interface{}) // Assert as JSON object type
				print("  %-50s %s\n", StrVal(o["userPrincipalName"]), StrVal(o["id"]))
			}
		} else {
			print("%-28s %s\n", "owners:", "None")
		}
	}
	ApiErrorCheck(r, trace())

	url = mg_url + "/beta/groups/" + id + "/members"
	r = ApiGet(url, mg_headers, nil)
	if r["value"] != nil {
		members := r["value"].([]interface{}) // Assert as JSON array type
		if len(members) > 0 {
			print("members:\n")
			// PrintJson(members) // DEBUG
			for _, i := range members {
				m := i.(map[string]interface{}) // Assert as JSON object type
				Type, Name := "-", "-"
				Type = LastElem(StrVal(m["@odata.type"]), ".")
				switch Type {
				case "group", "servicePrincipal":
					Name = StrVal(m["displayName"])
				default:
					Name = StrVal(m["userPrincipalName"])
				}
				print("  %-50s %s (%s)\n", Name, StrVal(m["id"]), Type)
			}
		} else {
			print("%-28s %s\n", "members:", "None")
		}
	}
	ApiErrorCheck(r, trace())

	// Print all groups/roles it is a member of
	memberOf := GetObjectMemberOfs("g", id) // For this Group object
	PrintMemberOfs("g", memberOf)
}

func PrintPAGs() {
	// List all Privileged Access Groups
	for _, i := range GetAllObjects("g") {  // Iterate through all objects
		x := i.(map[string]interface{})     // Assert JSON object type
		if x["isAssignableToRole"] != nil {
			isAssignableToRole := x["isAssignableToRole"].(bool)
			if isAssignableToRole {
				PrintTersely("g", x) // Pring group tersely
			}
		}
	}
}
