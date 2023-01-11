// sps.go

package main

import (
	"fmt"
	"strings"
	"path/filepath"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintSp(x JsonObject, z aza.AzaBundle, oMap MapString) {
	// Print service principal object in YAML-like format
	if x == nil {
		return
	}
	id := StrVal(x["id"])

	// Print the most important attributes
	list := []string{"displayName", "id", "appId", "accountEnabled", "servicePrincipalType"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" {
			fmt.Printf("%s: %s\n", i, v) // Only print non-null attributes
		}
	}

	// Print owners
	url := aza.ConstMgUrl + "/beta/servicePrincipals/" + id + "/owners"
	r := ApiGet(url, z.MgHeaders, nil)
	ApiErrorCheck(r, utl.Trace())
	if r["value"] != nil {
		owners := r["value"].([]interface{}) // JSON array
		if len(owners) > 0 {
			fmt.Printf("owners:\n")
			for _, i := range owners {
				o := i.(map[string]interface{}) // JSON object
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

	// Print members and their roles
	url = aza.ConstMgUrl + "/beta/servicePrincipals/" + id + "/appRoleAssignedTo"
	r = ApiGet(url, z.MgHeaders, nil)
	ApiErrorCheck(r, utl.Trace())
	if r["value"] != nil {
		members := r["value"].([]interface{}) // JSON array
		if len(members) > 0 {
			fmt.Printf("members:\n")

			// Build roleMap
			roleMap := make(map[string]string)
			if x["appRoles"] != nil {
				objAppRoles := x["appRoles"].([]interface{})
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
			fmt.Printf("%s: %s\n", "members", "None")
		}
	}

	// Print groups & roles it is a member of
	memberOf := GetObjectMemberOfs("sp", id, z, oMap) // For this SP object
	PrintMemberOfs("sp", memberOf)

	// Print API permissions 
	url = aza.ConstMgUrl + "/v1.0/servicePrincipals/" + id + "/oauth2PermissionGrants"
	r = ApiGet(url, z.MgHeaders, nil)
	ApiErrorCheck(r, utl.Trace())
	if r["value"] != nil && len(r["value"].([]interface{})) > 0 {
		fmt.Printf("api_permissions:\n")
		apiPerms := r["value"].([]interface{}) // Assert as JSON array

		// Print OAuth 2.0 scopes for each API
		for _, i := range apiPerms {
			api := i.(map[string]interface{}) // Assert as JSON object
			apiName := "Unknown"
			id := StrVal(api["resourceId"])   // Get API's SP to get its displayName
			url2 := aza.ConstMgUrl + "/v1.0/servicePrincipals/" + id
			r2 := ApiGet(url2, z.MgHeaders, nil)
			if r2["appDisplayName"] != nil {
				apiName = StrVal(r2["appDisplayName"])
			}
			ApiErrorCheck(r2, utl.Trace())

			// Print each delegated claim for this API
			scope := strings.TrimSpace(StrVal(api["scope"]))
            claims := strings.Split(scope, " ")
			for _, j := range claims {
				fmt.Printf("  %-50s %s\n", apiName, j)
			}
		}
	}
}

func SpsCountLocal(z aza.AzaBundle) (microsoft, native int64) {
	// Dedicated SPs local cache counter able to discern if SP is owned by native tenant or it's a Microsoft default SP 
	var microsoftList []interface{} = nil
	var nativeList []interface{} = nil
	localData := filepath.Join(z.ConfDir, z.TenantId + "_servicePrincipals.json")
    if utl.FileUsable(localData) {
		rawList, _ := utl.LoadFileJson(localData) // Load cache file
		if rawList != nil {
			sps := rawList.([]interface{}) // Assert as JSON array type
			for _, i := range sps {
				x := i.(map[string]interface{}) // Assert as JSON object type
				owner := StrVal(x["appOwnerOrganizationId"])
				if owner == z.TenantId {  // If owned by current tenant ...
					nativeList = append(nativeList, x)
				} else {
					microsoftList = append(microsoftList, x)
				}
			}
			return int64(len(microsoftList)), int64(len(nativeList))
		}
	}
	return 0, 0
}

func SpsCountAzure(z aza.AzaBundle, oMap MapString) (microsoft, native int64) {
	// Dedicated SPs Azure counter able to discern if SP is owned by native tenant or it's a Microsoft default SP
	// NOTE: Not entirely accurate yet because function GetAllObjects still checks local cache. Need to refactor
	// that function into 2 diff versions GetAllObjectsLocal and GetAllObjectsAzure and have this function the latter.
	var microsoftList []interface{} = nil
	var nativeList []interface{} = nil
	sps := GetAzObjects("sp", false, z, oMap) // false = be silent
	if sps != nil {
		for _, i := range sps {
			x := i.(map[string]interface{}) // Assert as JSON object type
			owner := StrVal(x["appOwnerOrganizationId"])
			if owner == z.TenantId {  // If owned by current tenant ...
				nativeList = append(nativeList, x)
			} else {
				microsoftList = append(microsoftList, x)
			}
		}
	}
	return int64(len(microsoftList)), int64(len(nativeList))
}
