// helper.go

package main

import (
	"os"
	"fmt"
	"log"	
	"path/filepath"
	"time"
	"runtime"

	"github.com/google/uuid"
)

func exit(code int) {
	os.Exit(code)       // Syntactic sugar. Easier to type
}

func print(format string, args ...interface{}) (n int, err error) {
	return fmt.Printf(format, args...) // More syntactic sugar
}

func die(format string, args ...interface{}) {
	fmt.Printf(format, args...) // Same as print function but does not return
	os.Exit(1)                  // Always exit with return code 1
}

func sprint(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)	// More syntactic sugar
}

func trace() (string) {
	// Return string showing current "File_path [line number] function_name"
	// https://stackoverflow.com/questions/25927660/how-to-get-the-current-function-name
    progCounter, fp, ln, ok := runtime.Caller(1)
    if !ok { return sprint("%s\n    %s:%d\n", "?", "?", 0) }
    funcPointer := runtime.FuncForPC(progCounter)
    if funcPointer == nil { return sprint("%s\n    %s:%d\n", "?", fp, ln) }
	return sprint("%s\n    %s:%d\n", funcPointer.Name(), fp, ln)
}

func ValidUuid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func SameType(a, b interface{}) bool {
	// Check if two variables are of the same type
	a_type := sprint("%T", a)
	b_type := sprint("%T", b)
	return a_type == b_type
}

func VarType(v interface{}) string {
	return sprint("%T", v)
}

func GetAllObjects(t string) (oList []interface{}) {
	// This function determines whether to retrieve all the objects from Azure or from the local
	// cache based on the age of the local cache.
	//
	// Also, note that we are dealing with two (2) classes of objects:
	// 1) AZ types = Azure RBAC definitions and assignments, subscriptions, and mgmt group objects.
	// 2) MG types = MS Graph users, groups, service principals, and applications objects.
	// The set of all AZ types per tenant can be relatively small and fast to retrieve from Azure, so
	// we'll use a crude cachePeriod for local cache. On the other hand, the set of MG types can
	// can be very large, so we will rely on MG delta query methods to keep the local data store
	// cached and in sync.

	oList = nil                         // Start with an empty list
	cachePeriod := int64(3660 * 24 * 7) // One week cache period
	// TODO: Make above variable a $HOME/.$PROGMAN/config file option

	localData := filepath.Join(confdir, tenant_id+"_"+oMap[t]+".json") // Define local data store file
	cacheFileAge := int64(0)
	if FileUsable(localData) {
		cacheFileEpoc := int64(FileModTime(localData))
		cacheFileAge = int64(time.Now().Unix()) - cacheFileEpoc
		l, _ := LoadFileJson(localData) // Load cache file
		if l != nil {
			oList = l.([]interface{}) // Get cached objects
			if (t == "d" || t == "a" || t == "s" || t == "m") && cacheFileAge < cachePeriod {
				// Shortcut. For AZ types just use local cache if within cachePeriod.
				return oList
				// NOTE: Getting all subscriptions below is a quick API call,
				// so maybe its cache period should be shorter.
			}
		}
	}

	// Get all objects of type t
	switch t {
	case "d":
		oList = GetRoleDefinitions("verbose")   // Get all role definitions from Azure
		SaveFileJson(oList, localData) // Cache it to local file
	case "a":
		oList = GetRoleAssignments()   // Get all role assignments from Azure
		SaveFileJson(oList, localData) // Cache it to local file
	case "s":
		oList = GetSubscriptions()     // Get all subscriptions from Azure
		SaveFileJson(oList, localData) // Cache it to local file
	case "m":
		oList = GetManagementGroups()  // Get all Management Groups from Azure
		SaveFileJson(oList, localData) // Cache it to local file
	case "rd":
		// Azure AD Role Definitions are under 'roleManagement/directory' so we're forced to process them
		// differently than other MS Graph objects. Microsoft, this is not pretty :-(
		url := mg_url + "/v1.0/roleManagement/directory/roleDefinitions"
		r := ApiGet(url, nil, nil, false)
		if r["value"] != nil {
			oList := r["value"].([]interface{}) // Treat as JSON array type
			SaveFileJson(oList, localData) // Cache it to local file
		}
		ApiErrorCheck(r, trace())
	case "u", "g", "sp", "ap", "ra":
		// Use this file to keep track of the delta link for doing delta queries
		// See https://docs.microsoft.com/en-us/graph/delta-query-overview
		deltaLinkFile := filepath.Join(confdir, tenant_id+"_"+oMap[t]+"_deltaLink.json")
		var deltaLinkMap map[string]interface{} = nil

		var fullQuery bool = true
		url := mg_url + "/v1.0/" + oMap[t] + "/delta?$select=" // Base URL
		deltaAge := int64(time.Now().Unix()) - int64(FileModTime(deltaLinkFile))
		// DeltaLink files cannot be older than 30 days (using 27)
		if (deltaAge < int64(3660 * 24 * 27)) && FileUsable(deltaLinkFile) && len(oList) > 0 {
			// log.Println("Delta query") // DEBUG
			fullQuery = false
			tmpVal, _ := LoadFileJson(deltaLinkFile)
			deltaLinkMap = tmpVal.(map[string]interface{}) // Assert as JSON object
			url = StrVal(deltaLinkMap["@odata.deltaLink"])                      // Delta URL
		} else {
			// log.Println("Full query") // DEBUG
			switch t { // Build attribute select URL depending on type
			case "u":
				url = url + "displayName,mailNickname,userPrincipalName,onPremisesSamAccountName,onPremisesDomainName,onPremisesUserPrincipalName"
			case "g":
				url = url + "displayName,mailNickname,description,isAssignableToRole,mailEnabled"
			case "sp":
				url = url + "displayName,appId,accountEnabled,servicePrincipalType,appOwnerOrganizationId"
			case "ap":
				url = url + "displayName,appId,requiredResourceAccess"
			case "ra":
				url = url + "displayName,description,roleTemplateId"
			}
		}

		azureCount := ObjectCountAzure(t) // Get number of objects in Azure right at this moment
		if fullQuery {
			log.Printf("%d objects to get\n", azureCount)
		}

		apiCalls := 1 // Track how often we call API before getting the deltaLink

		var deltaSet []interface{} = nil                         // Assume zero new delta objects
		headers := map[string]string{"Prefer": "return=minimal"} // Additional required header

		r := ApiGet(url, headers, nil, false)
		ApiErrorCheck(r, trace())
		for {
			// Infinite loop until deltalLink appears (meaning we're done getting current delta set)
			if r["value"] != nil {
				// Continue building deltaSet
				thisBatch := r["value"].([]interface{}) // Treat as JSON array type
				if len(thisBatch) > 0 {
					deltaSet = append(deltaSet, thisBatch...) // Concatenate this set to growing list
				}
			}

			if fullQuery {
				print("\r%d (API calls = %d)", len(deltaSet), apiCalls) // Progress count indicator
			}

			if r["@odata.deltaLink"] != nil {
				// If deltaLink appears it means we're done retrieving initial set and we can break out of for-loop
				deltaLinkMap = map[string]interface{}{"@odata.deltaLink": StrVal(r["@odata.deltaLink"])}
				SaveFileJson(deltaLinkMap, deltaLinkFile) // Save new deltaLink for next call

				// print("\nLocal count = %d (before merge/cleanup)\n", len(oList))
				// print("Delta count = %d\n", len(deltaSet))

				// New objects returned, let's run our merge logic
				oList = NormalizeCache(oList, deltaSet)

				// print("Local count = %d (after merge/cleanup)\n", len(oList))
				// print("Azure count = %d\n", azureCount)

				SaveFileJson(oList, localData) // Cache it to local file
				break                          // from infinite for-loop
			}
			r = ApiGet(StrVal(r["@odata.nextLink"]), headers, nil, false) // Get next batch
			ApiErrorCheck(r, trace())
			apiCalls++
		}
		if fullQuery {
			print("\n")
			log.Printf("%d objects fetched\n", len(oList))
		}
	}
	return oList
}

func GetIdNameMap(t string) (idNameMap map[string]string) {
	// Return uuid:name map for given object type t
	idNameMap = make(map[string]string)
	for _, i := range GetAllObjects(t) { // Iterate through all objects
		x := i.(map[string]interface{}) // Assert JSON object type
		switch t {
		case "d":
			// Role definitions
			if x["name"] != nil {
				xProps := x["properties"].(map[string]interface{})
				if xProps["roleName"] != nil {
					// Assert them as JSON string type and add them to map
					idNameMap[StrVal(x["name"])] = StrVal(xProps["roleName"])
				}
			}
		case "s":
			// Subscriptions
			if x["subscriptionId"] != nil && x["displayName"] != nil {
				// Assert them as JSON string type and add them to map
				idNameMap[StrVal(x["subscriptionId"])] = StrVal(x["displayName"])
			}
		case "u", "g", "sp", "ap", "ra", "rd":
			// MS Graph objects all use same Id and displayName attributes
			if x["id"] != nil && x["displayName"] != nil {
				// Assert them as JSON string type and add them to map
				idNameMap[StrVal(x["id"])] = StrVal(x["displayName"])
			}
		}
	}
	return idNameMap
}

func GetMatching(t, name string) (oList []interface{}) {
	// List all objects of type t whose displayName matches name
	oList = nil
	switch t {
	case "d":
		for _, i := range GetAllObjects(t) {
			x := i.(map[string]interface{}) // Assert JSON object type
			xProps := x["properties"].(map[string]interface{})
			if xProps != nil && xProps["roleName"] != nil {
				if SubString(StrVal(xProps["roleName"]), name) {
					oList = append(oList, x)
				}
			}
		}
	case "a":
		roleMap := GetIdNameMap("d")
		for _, i := range GetAllObjects(t) {
			x := i.(map[string]interface{}) // Assert JSON object type
			xProps := x["properties"].(map[string]interface{})
			if xProps != nil && xProps["roleDefinitionId"] != nil {
				Rid := StrVal(xProps["roleDefinitionId"])
				roleName := roleMap[LastElem(Rid, "/")]
				if SubString(roleName, name) {
					oList = append(oList, x)
				}
			}
		}
	case "m":
		for _, i := range GetAllObjects(t) {
			x := i.(map[string]interface{}) // Assert JSON object type
			xProps := x["properties"].(map[string]interface{})
			if xProps != nil && xProps["displayName"] != nil {
				if SubString(StrVal(xProps["displayName"]), name) {
					oList = append(oList, x)
				}
			}
		}
	case "u":
		for _, i := range GetAllObjects(t) {
			x := i.(map[string]interface{}) // Assert JSON object type
			// Search relevant attributes
			searchList := []string{"displayName", "userPrincipalName", "mailNickname", "onPremisesSamAccountName", "onPremisesUserPrincipalName"}
			for _, i := range searchList {
				if SubString(StrVal(x[i]), name) {
					oList = append(oList, x)
					break // on first match
				}
			}
		}
	case "s", "g", "sp", "ap", "ra", "rd":
		for _, i := range GetAllObjects(t) {
			x := i.(map[string]interface{}) // Assert JSON object type
			if x != nil && x["displayName"] != nil {
				if SubString(StrVal(x["displayName"]), name) {
					oList = append(oList, x)
				}
			}
		}
	}
	return oList
}

func GetObjectMemberOfs(t, id string) (oList []interface{}) {
	// Get all group/role objects this object of type 't' with 'id' is a memberof
	oList = nil
	r := ApiGet(mg_url+"/beta/"+oMap[t]+"/"+id+"/memberof", mg_headers, nil, false)
	if r["value"] != nil {
		oList = r["value"].([]interface{}) // Assert as JSON array type
	}
	ApiErrorCheck(r, trace())
	return oList
}

func GetObjectFromFile(filePath string) (formatType, t string, obj map[string]interface{}) {
	// Returns 3 values: File format type, oMap type, and the object itself


	// Because JSON is essentially a subset of YAML, we have to check JSON first
	// As an aside, see https://news.ycombinator.com/item?id=31406473
	objRaw, _ := LoadFileJson(filePath)  // Ignore the errors
	formatType = "JSON"
	if objRaw == nil {
		objRaw, _ = LoadFileYaml(filePath)  // See if it's YAML, ignoring the error
		if objRaw == nil {
			return "", "", nil           // It's neither, return null values
		}
		formatType = "YAML"
	}
	obj = objRaw.(map[string]interface{})  // Henceforht, assert as single JSON object

	// Continue unpacking the object to see what it is
	xProps, ok := obj["properties"].(map[string]interface{}) // See if it has a top-level 'properties' section
	if !ok {
		return formatType, "", nil // Assertion failed, also return null values
	}
	roleName := StrVal(xProps["roleName"])       // Assert and assume it's a definition
	roleId := StrVal(xProps["roleDefinitionId"]) // assert and assume it's an assignment

    if roleName != "" {
        // It's a role definition
        return formatType, "d", obj   // Type can then neatly be used with oMap generized functions
    } else if roleId != "" {
        // It's a role assignment 
        return formatType, "a", obj
    } else {
        return formatType, "", obj
    }
}
