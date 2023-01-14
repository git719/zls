// apps.go

package main

import (
	"fmt"
	"path/filepath"
	"time"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintApp(x map[string]interface{}, z aza.AzaBundle, oMap map[string]string) {
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

func AppsCountLocal(z aza.AzaBundle) (int64) {
	// Return number of entries in local cache file
	var cachedList []interface{} = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_applications.json")
	if utl.FileUsable(cacheFile) {
		rawList, _ := utl.LoadFileJson(cacheFile)
		if rawList != nil {
			cachedList = rawList.([]interface{})
			return int64(len(cachedList))
		}
	}
	return 0
}	

func AppsCountAzure(z aza.AzaBundle) (int64) {
	// Return number of entries in Azure tenant
	z.MgHeaders["ConsistencyLevel"] = "eventual"
	url := aza.ConstMgUrl + "/v1.0/applications/$count"
	r := ApiGet(url, z.MgHeaders, nil)
	ApiErrorCheck(r, utl.Trace())
	if r["value"] != nil {
		return r["value"].(int64) // Expected result is a single int64 value for the count
	}
	return 0	
}

func GetApps(filter string, force bool, z aza.AzaBundle) (list []interface{}) {
	// Get all Azure AD applications whose searchAttributes match on 'filter'. An empty "" filter returns all.
	// Uses local cache if it's less than cachePeriod old. The 'force' option forces calling Azure query.
	list = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_applications.json")
	cacheNoGood, list := CheckLocalCache(cacheFile, 86400) // cachePeriod = 1 day in seconds
	if cacheNoGood || force {
		list = GetAzApps(cacheFile, z.MgHeaders, true) // Get all from Azure and show progress (verbose = true)
	}
	
	// Do filter matching
	if filter == "" {
		return list
	}
	var matchingList []interface{} = nil
	searchAttributes := []string{"id", "displayName", "appId"}
	var ids []string // Keep track of each unique objects to eliminate repeats
	for _, i := range list {
		x := i.(map[string]interface{})
		id := StrVal(x["id"])
		for _, i := range searchAttributes {
			if utl.SubString(StrVal(x[i]), filter) && !utl.ItemInList(id, ids) {
				matchingList = append(matchingList, x)
				ids = append(ids, id)
			}
		}
	}
	return matchingList	
}

func GetAzApps(cacheFile string, headers aza.MapString, verbose bool) (list []interface{}) {
	// Get all Azure AD service principal in current tenant AND save them to local cache file. Show progress if verbose = true.
	
	// We will first try doing a delta query. See https://docs.microsoft.com/en-us/graph/delta-query-overview
	var deltaLinkMap map[string]string = nil
	deltaLinkFile := cacheFile[:len(cacheFile)-len(filepath.Ext(cacheFile))] + "_deltaLink.json"
	deltaAge := int64(time.Now().Unix()) - int64(utl.FileModTime(deltaLinkFile))
	baseUrl := aza.ConstMgUrl + "/v1.0/applications"
    // Get delta updates only when below selection of attributes are modified
	selection := "?$select=displayName,appId,requiredResourceAccess"
	url := baseUrl + "/delta" + selection + "&$top=999"
	headers["Prefer"] = "return=minimal" // This tells API to focus only on specific 'select' attributes

	// But first, double-check the base set again to avoid running a delta query on an empty set
	listIsEmpty, list := CheckLocalCache(cacheFile, 86400) // cachePeriod = 1 day in seconds
	if  utl.FileUsable(deltaLinkFile) && deltaAge < (3660 * 24 * 27) && listIsEmpty == false {
		// Note that deltaLink file age has to be within 30 days (we do 27)
		tmpVal, _ := utl.LoadFileJson(deltaLinkFile)
		deltaLinkMap = tmpVal.(map[string]string)
		url = StrVal(deltaLinkMap["@odata.deltaLink"]) // Base URL is now the cached Delta Link
	}

    // Now go get azure objects using the updated URL (either a full query or a deltaLink query)
	var deltaSet []interface{} = nil
	deltaSet, deltaLinkMap = GetAzObjects(url, headers, verbose) // Run generic deltaSet retriever function

	// Save new deltaLink for future call, and merge newly acquired delta set with existing list
	utl.SaveFileJson(deltaLinkMap, deltaLinkFile)
	list = NormalizeCache(list, deltaSet) // Run our MERGE LOGIC with new delta set
	utl.SaveFileJson(list, cacheFile) // Update the local cache
	return list
}

func GetAzAppById(id string, headers aza.MapString) (map[string]interface{}) {
	// Get Azure AD application by its Object UUID or by its appId, with extended attributes
	baseUrl := aza.ConstMgUrl + "/v1.0/applications"
	selection := "?$select=id,addIns,api,appId,applicationTemplateId,appRoles,certification,createdDateTime,"
	selection += "deletedDateTime,disabledByMicrosoftStatus,displayName,groupMembershipClaims,id,identifierUris,"
	selection += "info,isDeviceOnlyAuthSupported,isFallbackPublicClient,keyCredentials,logo,notes,"
	selection += "oauth2RequiredPostResponse,optionalClaims,parentalControlSettings,passwordCredentials,"
	selection += "publicClient,publisherDomain,requiredResourceAccess,serviceManagementReference,"
	selection += "signInAudience,spa,tags,tokenEncryptionKeyId,verifiedPublisher,web"
	url := baseUrl + "/" + id + selection // First search is for direct Object Id
	r := ApiGet(url, headers, nil)
    if r != nil && r["error"] != nil {
		// Second search is for this app's application Client Id
		url = baseUrl + selection
		params := aza.MapString{"$filter": "appId eq '" + id + "'"}
		r := ApiGet(url, headers, params)
		ApiErrorCheck(r, utl.Trace())
		if r != nil && r["value"] != nil {
			list := r["value"].([]interface{})
			count := len(list)
			if count == 1 {
				return list[0].(map[string]interface{})  // Return single value found
			} else if count > 1 {
				// Not sure this would ever happen, but just in case
				fmt.Printf("Found %d entries with this appId\n", count)
				return nil
			} else {
				return nil
			}
		}
	}
	return r
}
