// helper.go

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func StrVal(x interface{}) string {
	return utl.StrVal(x)		// Shorthand
}

func GetAzRbacScopes(z aza.AzaBundle) (scopes []string) {
	// Get all scopes from the Azure RBAC hierarchy
	scopes = nil
	managementGroups := GetAzMgGroups(z) // Let's start with all managementGroups scopes
	for _, i := range managementGroups {
		x := i.(map[string]interface{})
		scopes = append(scopes, StrVal(x["id"]))
	}
	subscriptions := GetAzSubscriptions(z) // Now add all the subscription scopes
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

func CheckLocalCache(cacheFile string, cachePeriod int64) (usable bool, cachedList JsonArray) {
	// Return locally cached list of objects if it exists *and* it is within the specified cachePeriod in seconds 
	if utl.FileUsable(cacheFile) {
		cacheFileEpoc := int64(utl.FileModTime(cacheFile))
		cacheFileAge := int64(time.Now().Unix()) - cacheFileEpoc
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

func GetObjects(t, filter string, force bool, z aza.AzaBundle, oMap MapString) (list JsonArray) {
	// Generic function to get objects of type t whose attributes match on filter. If filter is
	// the "" empty string, then return ALL of the objects of this type.
	switch t {
	case "d":
		return GetRoleDefinitions(filter, force, true, z, oMap) // true = verbose, to print progress while getting
	case "a":
		return GetRoleAssignments(filter, force, true, z, oMap) // true = verbose, to print progress while getting
	case "s":
		return GetSubscriptions(filter, force, z)
	case "m":
		return GetMgGroups(filter, force, z)
	case "u":
		return GetUsers(filter, force, z)
	case "g":
		return GetGroups(filter, force, z)
	case "sp":
		return GetSps(filter, force, z)
	case "ap":
		return GetApps(filter, force, z)
	case "ad":
		return GetAdRoles(filter, force, z)
	}
	return nil
}

func GetAzObjectsLooper(url, cacheFile string, headers aza.MapString, verbose bool) (list JsonArray) {
	// Generic Azure object retriever function
	// This function implements the pattern described at https://docs.microsoft.com/en-us/graph/delta-query-overview
	
	var deltaSet JsonArray = nil // Assume zero entries have been updated
	calls := 1 // Track number of API calls
	r := ApiGet(url, headers, nil)
	ApiErrorCheck(r, utl.Trace())
	for {
		// Infinite for-loop until deltalLink appears (meaning we're done getting current delta set)
		var thisBatch JsonArray = nil // Assume zero entries in this batch
		if r["value"] != nil {
			thisBatch = r["value"].([]interface{})
			if len(thisBatch) > 0 {
				deltaSet = append(deltaSet, thisBatch...) // Continue growing deltaSet
			}
		}
		if verbose {
			// Progress count indicator. Using global var rUp to overwrite last line. Defer newline until done
			fmt.Printf("%s(API calls = %d) %d objects in set %d", rUp, calls, len(thisBatch), calls)
		}
		if r["@odata.deltaLink"] != nil {
			// BREAK infinite for-loop when deltaLink appears
			deltaLinkFile := cacheFile[:len(cacheFile)-len(filepath.Ext(cacheFile))] + "_deltaLink.json"
			deltaLinkMap := JsonObject{"@odata.deltaLink": StrVal(r["@odata.deltaLink"])}
			utl.SaveFileJson(deltaLinkMap, deltaLinkFile) // Save new deltaLink for future call
			list = NormalizeCache(list, deltaSet) // New objects returned, run our MERGE LOGIC
			utl.SaveFileJson(list, cacheFile) // Update the local cache
			break
		}
		r = ApiGet(StrVal(r["@odata.nextLink"]), headers, nil)  // Get next batch
		ApiErrorCheck(r, utl.Trace())
		calls++
	}
	if verbose {
		fmt.Printf("\n")
	}
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
		case "u", "g", "sp", "ap", "ad": // All MS Graph objects use same Id and displayName attributes
			if x["id"] != nil && x["displayName"] != nil {
				idNameMap[StrVal(x["id"])] = StrVal(x["displayName"])
			}
		}
	}
	return idNameMap
}

func GetObjectMemberOfs(t, id string, z aza.AzaBundle, oMap MapString) (list JsonArray) {
	// Get all group/role objects this object of type 't' with 'id' is a memberof
	list = nil
	url := aza.ConstMgUrl  + "/beta/" + oMap[t] + "/" + id + "/memberof"
	r := ApiGet(url, z.MgHeaders, nil)
	ApiErrorCheck(r, utl.Trace())
	if r["value"] != nil {
		list = r["value"].([]interface{})
	}
	return list
}

func RemoveCacheFile(t string, z aza.AzaBundle, oMap MapString) {
	// Remove cache file for objects of type t, or all of them
	switch t {
	case "t": // Token file is a little special: It doesn't use tenant ID
		utl.RemoveFile(filepath.Join(z.ConfDir, z.TokenFile))
	case "d", "a", "s", "m", "u", "g", "sp", "ap", "ad":
		utl.RemoveFile(filepath.Join(z.ConfDir, z.TenantId + "_" + oMap[t] + ".json"))
		// Types d, a, s, m, & ad do not have deltaLink files, but below won't balk anyway
		utl.RemoveFile(filepath.Join(z.ConfDir, z.TenantId + "_" + oMap[t] + "_deltaLink.json"))
	case "all":
		for _, i := range oMap {
			utl.RemoveFile(filepath.Join(z.ConfDir, z.TenantId + "_" + i + ".json"))
			utl.RemoveFile(filepath.Join(z.ConfDir, z.TenantId + "_" + i + "_deltaLink.json"))
		}
	}
	os.Exit(0)
}

func GetObjectFromFile(filePath string) (formatType, t string, obj JsonObject) {
	// Returns 3 values: File format type, oMap type, and the object itself

	// Because JSON is essentially a subset of YAML, we have to check JSON first
	// As an aside, see https://news.ycombinator.com/item?id=31406473
	objRaw, _ := utl.LoadFileJson(filePath) // Ignore the errors
	formatType = "JSON"
	if objRaw == nil {
		objRaw, _ = utl.LoadFileYaml(filePath) // See if it's YAML, ignoring the error
		if objRaw == nil {
			return "", "", nil // It's neither, return null values
		}
		formatType = "YAML"
	}
	obj = objRaw.(map[string]interface{})  // Henceforht, assert as single JSON object

	// Continue unpacking the object to see what it is
	xProp, ok := obj["properties"].(map[string]interface{})
	if !ok {
		return formatType, "", nil // Assertion failed, also return null values
	}
	roleName := StrVal(xProp["roleName"]) // Assert and assume it's a definition
	roleId := StrVal(xProp["roleDefinitionId"]) // assert and assume it's an assignment

    if roleName != "" {
        return formatType, "d", obj // It's a role definition. Type can neatly be used with oMap
    } else if roleId != "" {
        return formatType, "a", obj // It's a role assignment 
    } else {
        return formatType, "", obj
    }
}

func CompareSpecfile(filePath string, z aza.AzaBundle, oMap MapString) {
	if utl.FileNotExist(filePath) || utl.FileSize(filePath) < 1 {
		utl.Die("File does not exist, or is zero size\n")
	}
    ft, t, x := GetObjectFromFile(filePath)
	if ft != "JSON" && ft != "YAML" {
        utl.Die("File is not in JSON nor YAML format\n")
    }
    if t != "d" && t != "a" {
        utl.Die("This " + ft + " file is not a role definition nor an assignment specfile\n")
    }
	
    fmt.Printf("==== SPECFILE ============================\n")
    PrintObject(t, x, z, oMap) // Use generic print function
    fmt.Printf("==== AZURE ===============================\n")
    if t == "d" {
        y := GetAzRoleDefinition(x, z)
        if y == nil {
            fmt.Printf("Above definition does NOT exist in current Azure tenant\n")
        } else {
			PrintRoleDefinition(y, z, oMap) // Use specific role def print function
        }
    } else {
        y := GetAzRoleAssignment(x, z)
        if y == nil {
            fmt.Printf("Above assignment does NOT exist in current Azure tenant\n")
        } else {
			PrintRoleAssignment(y, z, oMap) // Use specific role assgmnt print function
        }
    }
    os.Exit(0)	
}
