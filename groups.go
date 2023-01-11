// groups.go

package main

import (
	"fmt"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintGroup(x JsonObject, z aza.AzaBundle, oMap MapString) {
	// Print group object in YAML-like format
	if x == nil {
		return
	}
	id := StrVal(x["id"])

	// First, print the most important attributes of this group
	list := []string{"displayName", "description", "id", "isAssignableToRole", "mailEnabled", "mailNickname"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" {
			fmt.Printf("%s: %s\n", i, v) // Only print non-null attributes
		}
	}

	// Print owners of this group
	url := aza.ConstMgUrl + "/beta/groups/" + id + "/owners"
	r := ApiGet(url, z.MgHeaders, nil)
	ApiErrorCheck(r, utl.Trace())
	if r["value"] != nil {
		owners := r["value"].([]interface{}) // Assert as JSON array type
		if len(owners) > 0 {
			fmt.Printf("owners:\n")
			for _, i := range owners {
				o := i.(map[string]interface{}) // Assert as JSON object type
				fmt.Printf("  %-50s %s\n", StrVal(o["userPrincipalName"]), StrVal(o["id"]))
			}
		} else {
			fmt.Printf("%s: %s\n", "owners", "None")
		}
	}

	// Print groups & roles this group is a member of
	memberOf := GetObjectMemberOfs("g", id, z, oMap) // For this Group object
	PrintMemberOfs("g", memberOf)

	// Print members of this group
	url = aza.ConstMgUrl + "/beta/groups/" + id + "/members"
	r = ApiGet(url, z.MgHeaders, nil)
	ApiErrorCheck(r, utl.Trace())
	if r["value"] != nil {
		members := r["value"].([]interface{}) // Assert as JSON array type
		if len(members) > 0 {
			fmt.Printf("members:\n")
			for _, i := range members {
				m := i.(map[string]interface{}) // Assert as JSON object type
				Type, Name := "-", "-"
				Type = utl.LastElem(StrVal(m["@odata.type"]), ".")
				switch Type {
				case "group", "servicePrincipal":
					Name = StrVal(m["displayName"])
				default:
					Name = StrVal(m["userPrincipalName"])
				}
				fmt.Printf("  %-50s %s (%s)\n", Name, StrVal(m["id"]), Type)
			}
		} else {
			fmt.Printf("%s: %s\n", "members", "None")
		}
	}
}

func PrintPags(z aza.AzaBundle, oMap MapString) {
	// List all Privileged Access Groups
	pagGroups := GetObjects("g", "", false, z, oMap)
	for _, i := range pagGroups {
		x := i.(map[string]interface{})     // Assert JSON object type
		if x["isAssignableToRole"] != nil {
			isAssignableToRole := x["isAssignableToRole"].(bool)
			if isAssignableToRole {
				PrintTersely("g", x) // Pring group tersely
			}
		}
	}
}
