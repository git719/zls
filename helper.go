// helper.go

package main

import (
	//"fmt"
	// "log"	
	"os"
	"path/filepath"
	"time"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func StrVal(x interface{}) string {
	return utl.StrVal(x)		// Shorthand
}

func GetAzRbacScopes(z aza.AzaBundle, oMap MapString) (scopes []string) {
	// Get all scopes from the Azure RBAC hierarchy
	scopes = nil
	// Let's start with all managementGroups scopes
	managementGroups := GetObjects("m", "", true, z, oMap) // true = force a call to Azure
	for _, i := range managementGroups {
		x := i.(map[string]interface{})
		scopes = append(scopes, StrVal(x["id"]))
	}
	// Now add all the subscription scopes
	subscriptions := GetObjects("s", "", true, z, oMap) // true = force a call to Azure
	for _, i := range subscriptions {
		x := i.(map[string]interface{})
		// Skip legacy subscriptions, since they have no role definitions and calling them causes an error
		if StrVal(x["displayName"]) == "Access to Azure Active Directory" {
			continue
		}
		scopes = append(scopes, StrVal(x["id"]))
	}
	return scopes
}

func CheckLocalCache(cacheFile string, cachePeriod int64) (usable bool, cachedList []interface{}) {
	// Return locally cached list of objects if it exists *and* it is within the specified cachePeriod 
	cacheFileAge := int64(0)
	if utl.FileUsable(cacheFile) {
		cacheFileEpoc := int64(utl.FileModTime(cacheFile))
		cacheFileAge = int64(time.Now().Unix()) - cacheFileEpoc
		rawList, _ := utl.LoadFileJson(cacheFile)
		if rawList != nil {
			cachedList = rawList.([]interface{})
			if len(cachedList) > 0 && cacheFileAge < cachePeriod {
				return false, cachedList // Cache is usable, returning cached list
			}
		}
	}
	return true, nil // Cache is not usable, returning nil
}

// ****** REFACTORING *********************************
// 1. Let individual object Get functions, like GetSubscriptions and GetMgGroups, do their own
//    detetermination on whether to use local cache or call Azure
// 2. Generize the call to return matches on a filter, with empty "" filter meaning ALL
// 3. Allow forcing a call to Azure with boolean 'force' variables
// ****** REFACTORING *********************************
func GetObjects(t, filter string, force bool, z aza.AzaBundle, oMap MapString) (list []interface{}) {
	// Get objects of type t whose attributes match on filter
	// If filter is the "" empty string then return ALL
	list = nil
	switch t {
	case "d":
		return GetRoleDefinitions(filter, force, true, z, oMap) // true = verbose, to print progress while getting
	case "a":
		// kljaskjdkfl
	case "s":
		return GetSubscriptions(filter, force, z)
	case "m":
		return GetMgGroups(filter, force, z)
	}

	// // This function determines whether to retrieve all the objects from Azure or from the local
	// // cache based on the age of the local cache.
	// //
	// // Also, note that we are dealing with two (2) classes of objects:
	// // 1) AZ types = Azure RBAC definitions and assignments, subscriptions, and mgmt group objects.
	// // 2) MG types = MS Graph users, groups, service principals, and applications objects.
	// // The set of all AZ types per tenant can be relatively small and fast to retrieve from Azure, so
	// // we'll use a crude cachePeriod for local cache. On the other hand, the set of MG types can
	// // can be very large, so we will rely on MG delta query methods to keep the local data store
	// // cached and in sync.

	// oList = nil
	// cachePeriod := int64(3660 * 24 * 7) // One week cache period (TODO: Make this a configurable variable?)
	// cacheFile := filepath.Join(confDir, tenantId + "_" + oMap[t] + ".json")
	// cacheIsUsable, oList := CheckLocalCache(cacheFile, cachePeriod)
	// if cacheIsUsable && (t == "d" || t == "a" || t == "s" || t == "m") {
	// 	return oList // Shortcut for AZ types, just use local cache if its within the cachePeriod
	// }

	// // Get all objects of type t
	// switch t {
	// case "d":
	// 	oList = GetAzRoleDefinitionAll(true, confDir, tenantId)  // verbose = true
	// 	utl.SaveFileJson(oList, cacheFile)    // Cache it to local file
	// case "a":
	// 	oList = GetAzRoleAssignmentAll(confDir, tenantId)
	// 	utl.SaveFileJson(oList, cacheFile)
	// case "s":
	// 	oList = GetAzSubscriptionAll()
	// 	utl.SaveFileJson(oList, cacheFile)
	// 	utl.SaveFileJson(oList, cacheFile)
	// case "rd":
	// 	// Azure AD Role Definitions are under 'roleManagement/directory' so we're forced to process them
	// 	// differently than other MS Graph objects. Microsoft, this is not pretty :-(
	// 	url := mg_url + "/v1.0/roleManagement/directory/roleDefinitions"
	// 	r := ApiGet(url, nil, nil)
	// 	if r["value"] != nil {
	// 		oList := r["value"].(JsonArray)
	// 		utl.SaveFileJson(oList, cacheFile) // Cache it to local file
	// 	}
	// 	ApiErrorCheck(r, utl.Trace())
	// case "u", "g", "sp", "ap", "ra":
	// 	// Use this file to keep track of the delta link for doing delta queries
	// 	// See https://docs.microsoft.com/en-us/graph/delta-query-overview
	// 	deltaLinkFile := filepath.Join(confDir, tenantId + "_" + oMap[t] + "_deltaLink.json")
	// 	var deltaLinkMap JsonObject = nil

	// 	var fullQuery bool = true
	// 	url := mg_url + "/v1.0/" + oMap[t] + "/delta?$select=" // Base URL
	// 	deltaAge := int64(time.Now().Unix()) - int64(utl.FileModTime(deltaLinkFile))
	// 	// DeltaLink files cannot be older than 30 days (using 27)
	// 	if (deltaAge < int64(3660 * 24 * 27)) && utl.FileUsable(deltaLinkFile) && len(oList) > 0 {
	// 		fullQuery = false
	// 		tmpVal, _ := utl.LoadFileJson(deltaLinkFile)
	// 		deltaLinkMap = tmpVal.(JsonObject)
	// 		url = StrVal(deltaLinkMap["@odata.deltaLink"])  // Delta URL
	// 	} else {
	// 		switch t { // Build attribute select URL depending on type
	// 		case "u":
	// 			url = url + "displayName,mailNickname,userPrincipalName,onPremisesSamAccountName,onPremisesDomainName,onPremisesUserPrincipalName"
	// 		case "g":
	// 			url = url + "displayName,mailNickname,description,isAssignableToRole,mailEnabled"
	// 		case "sp":
	// 			url = url + "displayName,appId,accountEnabled,servicePrincipalType,appOwnerOrganizationId"
	// 		case "ap":
	// 			url = url + "displayName,appId,requiredResourceAccess"
	// 		case "ra":
	// 			url = url + "displayName,description,roleTemplateId"
	// 		}
	// 	}

	// 	azureCount := ObjectCountAzure(t)  // Get number of objects in Azure right at this moment
	// 	if fullQuery {
	// 		log.Printf("%d objects to get\n", azureCount)
	// 	}

	// 	calls := 1 // Track how often we call API before getting the deltaLink

	// 	var deltaSet JsonArray = nil // Assume zero new delta objects
	// 	headers := map[string]string{"Prefer": "return=minimal"} // Additional required header

	// 	r := ApiGet(url, headers, nil)
	// 	ApiErrorCheck(r, utl.Trace())
	// 	for {
	// 		// Infinite loop until deltalLink appears (meaning we're done getting current delta set)
	// 		if r["value"] != nil {
	// 			// Continue building deltaSet
	// 			thisBatch := r["value"].(JsonArray)
	// 			// Now concatenate this set to growing list
	// 			if len(thisBatch) > 0 { deltaSet = append(deltaSet, thisBatch...) }
	// 		}

	// 		if fullQuery {
	// 			// Using global var rUp to overwrite last line. Defer newline until done
	// 			fmt.Printf("%s%d (API calls = %d)", rUp, len(deltaSet), calls) // Progress count indicator
	// 		}

	// 		if r["@odata.deltaLink"] != nil {
	// 			// If deltaLink appears it means we're done retrieving initial set and we can break out of for-loop
	// 			deltaLinkMap = JsonObject{"@odata.deltaLink": StrVal(r["@odata.deltaLink"])}
	// 			utl.SaveFileJson(deltaLinkMap, deltaLinkFile) // Save new deltaLink for next call

	// 			// fmt.Printf("\nLocal count = %d (before merge/cleanup)\n", len(oList))
	// 			// fmt.Printf("Delta count = %d\n", len(deltaSet))

	// 			// New objects returned, let's run our merge logic
	// 			oList = NormalizeCache(oList, deltaSet)

	// 			// fmt.Printf("Local count = %d (after merge/cleanup)\n", len(oList))
	// 			// fmt.Printf("Azure count = %d\n", azureCount)

	// 			utl.SaveFileJson(oList, cacheFile) // Cache it to local file
	// 			break                          // from infinite for-loop
	// 		}
	// 		r = ApiGet(StrVal(r["@odata.nextLink"]), headers, nil)  // Get next batch
	// 		ApiErrorCheck(r, utl.Trace())
	// 		calls++
	// 	}
	// 	if fullQuery {
	// 		fmt.Printf("\n")
	// 		log.Printf("%d objects fetched\n", len(oList))
	// 	}
	// }
	return list
}

func GetIdNameMap(t, filter string, force bool, z aza.AzaBundle, oMap MapString) (idNameMap map[string]string) {
	// Return uuid:name map for given object type t
	idNameMap = make(map[string]string)
	allObjects := GetObjects(t, "", false, z, oMap) // false = do NOT force a call to Azure
	for _, i := range allObjects {
		x := i.(map[string]interface{})
		switch t {
		case "d": // Role definitions
			if x["name"] != nil {
				xProp := x["properties"].(map[string]interface{})
				if xProp["roleName"] != nil {
					idNameMap[StrVal(x["name"])] = StrVal(xProp["roleName"]) 
				}
			}
		case "s": // Subscriptions
			if x["subscriptionId"] != nil && x["displayName"] != nil {
				idNameMap[StrVal(x["subscriptionId"])] = StrVal(x["displayName"])
			}
		case "u", "g", "sp", "ap", "ra", "rd": 	// All MS Graph objects use same Id and displayName attributes
			if x["id"] != nil && x["displayName"] != nil {
				idNameMap[StrVal(x["id"])] = StrVal(x["displayName"])
			}
		}
	}
	return idNameMap
}

// ##### NO LONGER NEEDED 
// REUSE MATCHING LOGIC
func GetMatching(t, filter string, z aza.AzaBundle, oMap MapString) (list []interface{}) {
	// List all objects of type t whose attributes match on filter
	list = nil
	switch t {
	// case "a":
	// 	roleMap := GetIdNameMap("d")
	// 	for _, i := range GetAllObjects(t, confDir, tenantId) {
	// 		x := i.(JsonObject)
	// 		xProps := x["properties"].(JsonObject)
	// 		if xProps != nil && xProps["roleDefinitionId"] != nil {
	// 			Rid := StrVal(xProps["roleDefinitionId"])
	// 			roleName := roleMap[utl.LastElem(Rid, "/")]
	// 			if utl.SubString(roleName, name) {
	// 				oList = append(oList, x)
	// 			}
	// 		}
	// 	}
	case "s":
		return GetSubscriptions(filter, false, z) // false = Ok to first try local cache
	case "m":
		return GetMgGroups(filter, false, z) // false = Ok to first try local cache
	// case "u":
	// 	for _, i := range GetAllObjects(t, confDir, tenantId) {
	// 		x := i.(JsonObject)
	// 		// Search relevant attributes
	// 		searchList := []string{"displayName", "userPrincipalName", "mailNickname", "onPremisesSamAccountName", "onPremisesUserPrincipalName"}
	// 		for _, i := range searchList {
	// 			if utl.SubString(StrVal(x[i]), name) {
	// 				oList = append(oList, x)
	// 				break // on first match
	// 			}
	// 		}
	// 	}
	// case "g", "sp", "ap", "ra", "rd":
	// 	for _, i := range GetAllObjects(t, confDir, tenantId) {
	// 		x := i.(JsonObject)
	// 		if x != nil && x["displayName"] != nil {
	// 			if utl.SubString(StrVal(x["displayName"]), name) {
	// 				oList = append(oList, x)
	// 			}
	// 		}
	// 	}
	}
	return list
}

// func GetObjectMemberOfs(t, id string) (oList []interface{}) {
// 	// Get all group/role objects this object of type 't' with 'id' is a memberof
// 	oList = nil
// 	url := mg_url + "/beta/" + oMap[t] + "/" + id + "/memberof"
// 	r := ApiGet(url, mg_headers, nil)
// 	if r["value"] != nil { oList = r["value"].([]interface{}) }  // Assert as JSON array type
// 	ApiErrorCheck(r, utl.Trace())
// 	return oList
// }

// func GetObjectFromFile(filePath string) (formatType, t string, obj JsonObject) {
// 	// Returns 3 values: File format type, oMap type, and the object itself

// 	// Because JSON is essentially a subset of YAML, we have to check JSON first
// 	// As an aside, see https://news.ycombinator.com/item?id=31406473
// 	objRaw, _ := utl.LoadFileJson(filePath)  // Ignore the errors
// 	formatType = "JSON"
// 	if objRaw == nil {
// 		objRaw, _ = utl.LoadFileYaml(filePath)  // See if it's YAML, ignoring the error
// 		if objRaw == nil { return "", "", nil }  // It's neither, return null values
// 		formatType = "YAML"
// 	}
// 	obj = objRaw.(JsonObject)  // Henceforht, assert as single JSON object

// 	// Continue unpacking the object to see what it is
// 	xProps, ok := obj["properties"].(JsonObject) // See if it has a top-level 'properties' section
// 	if !ok { return formatType, "", nil }  // Assertion failed, also return null values
// 	roleName := StrVal(xProps["roleName"])        // Assert and assume it's a definition
// 	roleId := StrVal(xProps["roleDefinitionId"])  // assert and assume it's an assignment

//     if roleName != "" {
//         return formatType, "d", obj  // It's a role definition. Type can neatly be used with oMap
//     } else if roleId != "" {
//         return formatType, "a", obj  // It's a role assignment 
//     } else {
//         return formatType, "", obj
//     }
// }

func RemoveCacheFile(t string, z aza.AzaBundle, oMap MapString) {
	// Remove cache file for objects of type t, or all of them
	switch t {
	case "t": // Token file is a little special: It doesn't use tenant ID
		utl.RemoveFile(filepath.Join(z.ConfDir, z.TokenFile))
	case "d", "a", "s", "m", "u", "g", "sp", "ap", "ra", "rd":
		utl.RemoveFile(filepath.Join(z.ConfDir, z.TenantId + "_" + oMap[t] + ".json"))
		// Below does not apply to d, a, s, & m, since deltaLink files are only for MS Graph
		utl.RemoveFile(filepath.Join(z.ConfDir, z.TenantId + "_" + oMap[t] + "_deltaLink.json"))
	case "all":
		for _, i := range oMap {
			utl.RemoveFile(filepath.Join(z.ConfDir, z.TenantId + "_" + i + ".json"))
			utl.RemoveFile(filepath.Join(z.ConfDir, z.TenantId + "_" + i + "_deltaLink.json"))
		}
	}
	os.Exit(0)
}

// func CompareSpecfile(filePath string) {
// 	if utl.FileNotExist(filePath) || utl.FileSize(filePath) < 1 {
// 		utl.Die("File does not exist, or is zero size\n")
// 	}
//     ft, t, x := GetObjectFromFile(filePath)
// 	if ft != "JSON" && ft != "YAML" {
//         utl.Die("File is not in JSON nor YAML format\n")
//     }
//     if t != "d" && t != "a" {
//         utl.Die("This " + ft + " file is not a role definition nor an assignment specfile\n")
//     }
	
//     fmt.Printf("==== SPECFILE ============================\n")
//     PrintObject(t, x)
//     fmt.Printf("==== AZURE ===============================\n")
//     if t == "d" {
//         y := GetAzRoleDefinition(x)
//         if y == nil {
//             fmt.Printf("Above definition does NOT exist in current Azure tenant\n")
//         } else {
//             PrintObject("d", y)
//         }
//     } else {
//         y := GetAzRoleAssignment(x)
//         if y == nil {
//             fmt.Printf("Above assignment does NOT exist in current Azure tenant\n")
//         } else {
//             PrintObject("a", y)
//         }
//     }
//     os.Exit(0)	
// }
