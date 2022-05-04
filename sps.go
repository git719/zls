// sps.go

package main

import (
	"fmt"
)

func PrintSP(obj map[string]interface{}) {
	// Print service principal object in YAML-like style format
	if obj["id"] == nil {
		return
	}
	id := StrVal(obj["id"])

	// Print the most important attributes
	list := []string{"displayName", "id", "appId", "accountEnabled", "servicePrincipalType"}
	for _, i := range list {
		fmt.Printf("%-21s %s\n", i+":", StrVal(obj[i]))
	}

	// Print owners
	r := APIGet(mg_url+"/beta/servicePrincipals/"+id+"/owners", mg_headers, nil, false)
	if r["value"] != nil {
		owners := r["value"].([]interface{}) // JSON array
		if len(owners) > 0 {
			fmt.Printf("owners:\n")
			for _, i := range owners {
				o := i.(map[string]interface{}) // JSON object
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
				fmt.Printf("  %-50s %s (%s)\n", Name, StrVal(o["id"]), Type)
			}
		} else {
			fmt.Printf("%-28s %s\n", "owners:", "None")
		}
	}

	// Print members and their roles
	r = APIGet(mg_url+"/beta/servicePrincipals/"+id+"/appRoleAssignedTo", mg_headers, nil, false)
	if r["value"] != nil {
		members := r["value"].([]interface{}) // JSON array
		if len(members) > 0 {
			fmt.Printf("members:\n")

			// Build roleMap
			roleMap := make(map[string]string)
			if obj["appRoles"] != nil {
				objAppRoles := obj["appRoles"].([]interface{})
				if len(objAppRoles) > 0 {
					for _, i := range objAppRoles {
						ar := i.(map[string]interface{})
						roleMap[StrVal(ar["id"])] = StrVal(ar["displayName"])
					}
				}
			}
			// Add Default Access role
			roleMap["00000000-0000-0000-0000-000000000000"] = "Default Access"

			for _, i := range members {
				rm := i.(map[string]interface{}) // JSON object
				principalName := StrVal(rm["principalDisplayName"])
				roleName := roleMap[StrVal(rm["appRoleId"])] // Reference role name
				principalId := StrVal(rm["principalId"])
				principalType := StrVal(rm["principalType"])
				fmt.Printf("  %-50s %-20s %s (%s)\n", principalName, roleName, principalId, principalType)
			}
		} else {
			fmt.Printf("%-28s %s\n", "members:", "None")
		}
	}

	// Print all groups/roles it is a member of
	memberOf := GetObjectMemberOfs("sp", id) // For this SP object
	PrintMemberOfs("sp", memberOf)

	// Print API permissions
	r = APIGet(mg_url+"/beta/servicePrincipals/"+id+"/appRoleAssignments", mg_headers, nil, false)
	if r["value"] != nil && len(r["value"].([]interface{})) > 0 {
		fmt.Println("api_permissions:")
		apiPerms := r["value"].([]interface{}) // Assert as JSON array

		// Getting API app role permission name, such as Directory.Read.All, is a 2-step process:
		// 1) Put all all the API "app role id":names pairs in a map.
		//    We do one loop to preprocess id:names, so we can cache and speeds things up.
		// 2) Do another loop to enumerate and print them

		// Create unique list of all API IDs
		var apiIds []string
		for _, i := range apiPerms {
			api := i.(map[string]interface{}) // Assert as JSON object type
			id := StrVal(api["resourceId"])
			if ItemInList(id, apiIds) {
				continue // This API ID is already in our growing list. Skip it and check the next one
			}
			apiIds = append(apiIds, id)
		}

		// Create unique map of all API app role ID + name pairs
		apiRoles := make(map[string]string)
		for _, resId := range apiIds {
			r := APIGet(mg_url+"/beta/servicePrincipals/"+resId, mg_headers, nil, false)
			if r["appRoles"] != nil {
				for _, i := range r["appRoles"].([]interface{}) { // Iterate through all roles
					role := i.(map[string]interface{}) // Assert JSON object type
					if role["id"] != nil && role["value"] != nil {
						apiRoles[StrVal(role["id"])] = StrVal(role["value"]) // Add entry to map
					}
				}
			}
		}

		// Print them
		for _, a := range apiPerms {
			api := a.(map[string]interface{}) // Assert as JSON object type

			apiName := StrVal(api["resourceDisplayName"]) // This API's name
			resId := StrVal(api["resourceId"])            // This API's object id

			if resId != "" {
				pid := StrVal(api["appRoleId"]) // App role ID
				fmt.Printf("  %-50s %s\n", apiName, apiRoles[pid])
			} else {
				fmt.Printf("  %-50s %s\n", apiName, "(Missing resourceId)")
			}
		}
	}
}
