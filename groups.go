// groups.go

package main

import (
	"fmt"
)

func PrintGroup(x map[string]interface{}) {
	// Print group object in YAML-like style format
	if x["id"] == nil {
		return
	}
	id := StrVal(x["id"])

	// Print the most important attributes
	list := []string{"displayName", "description", "id", "isAssignableToRole", "mailEnabled", "mailNickname"}
	for _, i := range list {
		fmt.Printf("%-28s %s\n", i+":", StrVal(x[i]))
	}

	// OPTIONAL: Print other attributes here

	r := APIGet(mg_url+"/beta/groups/"+id+"/owners", mg_headers, nil, false)
	if r["value"] != nil {
		owners := r["value"].([]interface{}) // Assert as JSON array type
		if len(owners) > 0 {
			fmt.Printf("owners:\n")
			for _, i := range owners {
				o := i.(map[string]interface{}) // Assert as JSON object type
				fmt.Printf("  %-50s %s\n", StrVal(o["userPrincipalName"]), StrVal(o["id"]))
			}
		} else {
			fmt.Printf("%-28s %s\n", "owners:", "None")
		}
	}

	r = APIGet(mg_url+"/beta/groups/"+id+"/members", mg_headers, nil, false)
	if r["value"] != nil {
		members := r["value"].([]interface{}) // Assert as JSON array type
		if len(members) > 0 {
			fmt.Printf("members:\n")
			// PrintJSON(members) // DEBUG
			for _, i := range members {
				m := i.(map[string]interface{}) // Assert as JSON object type
				Type, Name := "-", "-"
				Type = LastElem(StrVal(m["@odata.type"]), ".")
				switch Type {
				case "group":
					Name = StrVal(m["displayName"])
				case "servicePrincipal":
					spType := StrVal(m["servicePrincipalType"])
					switch spType {
					case "ManagedIdentity":
						Type = "MI"
						Name = StrVal(m["displayName"])
					case "Application":
						Type = "APP"
						Name = StrVal(m["displayName"])
					default:
						Name = "???"
					}
				default:
					Name = StrVal(m["userPrincipalName"])
				}
				fmt.Printf("  %-50s %s (%s)\n", Name, StrVal(m["id"]), Type)
			}
		} else {
			fmt.Printf("%-28s %s\n", "members:", "None")
		}
	}

	// Print all groups/roles it is a member of
	memberOf := GetObjectMemberOfs("g", id) // For this Group object
	PrintMemberOfs("g", memberOf)
}

func PrintPAGs() {
	// List all Privileged Access Groups
	for _, i := range GetAllObjects("g") { // Iterate through all objects
		x := i.(map[string]interface{}) // Assert JSON object type
		if x["isAssignableToRole"] != nil {
			PrintTersely("g", x) // Pring group tersely
		}
	}
}
