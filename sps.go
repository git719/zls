// sps.go

package main

import (
	"strings"
)

func PrintSP(obj map[string]interface{}) {
	// Print service principal object in YAML format
	if obj == nil { return 	}
	id := StrVal(obj["id"])

	// Print the most important attributes
	list := []string{"displayName", "id", "appId", "accountEnabled", "servicePrincipalType"}
	for _, i := range list {
		v := StrVal(obj[i])
		if v != "" { print("%-21s %s\n", i+":", v) } // Only print non-null attributes
	}

	// Print owners
	r := ApiGet(mg_url+"/beta/servicePrincipals/"+id+"/owners", mg_headers, nil, false)
	if r["value"] != nil {
		owners := r["value"].([]interface{}) // JSON array
		if len(owners) > 0 {
			print("owners:\n")
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
				print("  %-50s %s (%s)\n", Name, StrVal(o["id"]), Type)
			}
		} else {
			print("%-28s %s\n", "owners:", "None")
		}
	}
	ApiErrorCheck(r, trace())

	// Print members and their roles
	r = ApiGet(mg_url+"/beta/servicePrincipals/"+id+"/appRoleAssignedTo", mg_headers, nil, false)
	if r["value"] != nil {
		members := r["value"].([]interface{}) // JSON array
		if len(members) > 0 {
			print("members:\n")

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
				print("  %-50s %-20s %s (%s)\n", principalName, roleName, principalId, principalType)
			}
		} else {
			print("%-28s %s\n", "members:", "None")
		}
	}
	ApiErrorCheck(r, trace())

	// Print all groups/roles it is a member of
	memberOf := GetObjectMemberOfs("sp", id) // For this SP object
	PrintMemberOfs("sp", memberOf)

	// Print API permissions 
	r = ApiGet(mg_url+"/v1.0/servicePrincipals/"+id+"/oauth2PermissionGrants", mg_headers, nil, false)
	if r["value"] != nil && len(r["value"].([]interface{})) > 0 {
		print("api_permissions:\n")
		apiPerms := r["value"].([]interface{}) // Assert as JSON array

		// Print OAuth 2.0 scopes for each API
		for _, i := range apiPerms {
			api := i.(map[string]interface{}) // Assert as JSON object
			apiName := "Unknown"
			id := StrVal(api["resourceId"])   // Get API's SP to get its displayName
			r := ApiGet(mg_url+"/v1.0/servicePrincipals/"+id, mg_headers, nil, false)
			if r["appDisplayName"] != nil { apiName = StrVal(r["appDisplayName"]) }
			ApiErrorCheck(r, trace())

			// Print each delegated claim for this API
			scope := strings.TrimSpace(StrVal(api["scope"]))
            claims := strings.Split(scope, " ")
			for _, j := range claims {
				print("  %-50s %s\n", apiName, j)
			}
		}
	}
	ApiErrorCheck(r, trace())
}
