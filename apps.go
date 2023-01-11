// apps.go

package main

import (
	"fmt"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintApp(x JsonObject, z aza.AzaBundle, oMap MapString) {
	// Print application object in YAML-like format
	if x == nil {
		return
	}
	id := StrVal(x["id"])

	// Print the most important attributes first
	list := []string{"displayName", "appId", "id"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" {
			fmt.Printf("%s: %s\n", i, v) // Only print non-null attributes
		}
	}

	// Print owners
	url := aza.ConstMgUrl + "/beta/applications/" + id + "/owners"
	r := ApiGet(url, z.MgHeaders, nil)
	if r["value"] != nil {
		owners := r["value"].([]interface{})
		if len(owners) > 0 {
			fmt.Printf("owners:\n")
			// PrintJson(groups) // DEBUG
			for _, i := range owners {
				o := i.(map[string]interface{})
				Type, Name := "???", "???"
				Type = utl.LastElem(StrVal(o["@odata.type"]), ".")
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
			fmt.Printf("%s: %s\n", "owners", "None")
		}
	}
	ApiErrorCheck(r, utl.Trace())

	// Print all groups/roles it is a member of
	memberOf := GetObjectMemberOfs("ap", id, z, oMap) // For this App object
	PrintMemberOfs("ap", memberOf)

	// Print API permissions
	// Just look under this object's 'requiredResourceAccess' attribute
	if x["requiredResourceAccess"] != nil && len(x["requiredResourceAccess"].([]interface{})) > 0 {
		fmt.Printf("api_permissions:\n")
		APIs := x["requiredResourceAccess"].([]interface{}) // Assert to JSON array
		for _, a := range APIs {
			api := a.(map[string]interface{})
			// Getting this API's name and permission value such as Directory.Read.All is a 2-step process:
			// 1) Get all the roles for given API and put their id/value pairs in a map, then
			// 2) Use that map to enumerate and print them

			// Let's drill down into the permissions for this API
			if api["resourceAppId"] == nil {
				fmt.Printf("  %-50s %s\n", "Unknown API", "Missing resourceAppId")
				continue // Skip this API, move on to next one
			}

			// Let's drill down into the permissions for this API
			resAppId := StrVal(api["resourceAppId"])

			// Get this API's SP object with all relevant attributes
			url := aza.ConstMgUrl + "/beta/servicePrincipals?filter=appId+eq+'" + resAppId + "'"
			r := ApiGet(url, z.MgHeaders, nil)
			// Unclear why result is a list instead of a single entry
			if r["value"] == nil {
				fmt.Printf("  %-50s %s\n", resAppId, "Unable to get Resource App object. Skipping this API.")
				continue
			}
			ApiErrorCheck(r, utl.Trace())

			SPs := r["value"].([]interface{})
			if len(SPs) > 1 { utl.Die("  %-50s %s\n", resAppId, "Error. Multiple SPs for this AppId. Aborting.") }

			sp := SPs[0].(map[string]interface{}) // The only expected entry

			// 1. Put all API role id:name pairs into roleMap list
			roleMap := make(map[string]string)
			if sp["appRoles"] != nil {
				for _, i := range sp["appRoles"].([]interface{}) { // Iterate through all roles
					// These are for Application types
					role := i.(map[string]interface{})
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
				fmt.Printf("  %-50s %s\n", resAppId, "Error getting list of appRoles.")
				continue
			}

			// 2. Parse this app permissions, and use roleMap to display permission value
			if api["resourceAccess"] != nil && len(api["resourceAccess"].([]interface{})) > 0 {
				Perms := api["resourceAccess"].([]interface{})
				apiName := StrVal(sp["displayName"]) // This API's name
				for _, i := range Perms {            // Iterate through perms
					perm := i.(map[string]interface{})
					pid := StrVal(perm["id"]) // JSON string
					fmt.Printf("  %-50s %s\n", apiName, roleMap[pid])
				}
			} else {
				fmt.Printf("  %-50s %s\n", resAppId, "Error getting list of appRoles.")
			}
		}
	}
}
