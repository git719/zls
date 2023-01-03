// api_calls.go

package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func PrintCountStatus() {
	print("Note: Counting objects residing in Azure can take some time.\n")
	print("%-38s %-20s %s\n", "OBJECTS", "LOCAL_CACHE_COUNT","AZURE_COUNT")
	print("%-38s ", "Groups")
	print("%-20d %d\n", ObjectCountLocal("g"), ObjectCountAzure("g"))
	print("%-38s ", "Users")
	print("%-20d %d\n", ObjectCountLocal("u"), ObjectCountAzure("u"))
	print("%-38s ", "App Registrations")
	print("%-20d %d\n", ObjectCountLocal("ap"), ObjectCountAzure("ap"))
	microsoftSpsLocal, nativeSpsLocal := SpsCountLocal()
	microsoftSpsAzure, nativeSpsAzure := SpsCountAzure()
	print("%-38s ", "Service Principals Microsoft Default")
	print("%-20d %d\n", microsoftSpsLocal, microsoftSpsAzure)
	print("%-38s ", "Service Principals This Tenant")
	print("%-20d %d\n", nativeSpsLocal, nativeSpsAzure)
	print("%-38s ", "Azure AD Roles Definitions")
	print("%-20d %d\n", ObjectCountLocal("rd"), ObjectCountAzure("rd"))
	print("%-38s ", "Azure AD Roles Activated")
	print("%-20d %d\n", ObjectCountLocal("ra"), ObjectCountAzure("ra"))
	print("%-38s ", "Management Groups")
	print("%-20d %d\n", ObjectCountLocal("m"), ObjectCountAzure("m"))
	print("%-38s ", "Subscriptions")
	print("%-20d %d\n", ObjectCountLocal("s"), ObjectCountAzure("s"))
	builtinLocal, customLocal := RoleDefinitionCountLocal()
	builtinAzure, customAzure := RoleDefinitionCountAzure()
	print("%-38s ", "RBAC Role Definitions BuiltIn")
    print("%-20d %d\n", builtinLocal, builtinAzure)
	print("%-38s ", "RBAC Role Definitions Custom")
    print("%-20d %d\n", customLocal, customAzure)
	print("%-38s ", "RBAC Role Assignments")
	print("%-20d %d\n", ObjectCountLocal("a"), ObjectCountAzure("a"))
}

func SpsCountLocal() (microsoft, native int64) {
	// Dedicated SPs local cache counter able to discern if SP is owned by native tenant or it's a Microsoft default SP 
	var microsoftList []interface{} = nil
	var nativeList []interface{} = nil
	localData := filepath.Join(confdir, tenant_id+"_"+oMap["sp"]+".json")
    if FileUsable(localData) {
		l, _ := LoadFileJson(localData) // Load cache file
		if l != nil {
			sps := l.([]interface{}) // Assert as JSON array type
			for _, i := range sps {
				x := i.(map[string]interface{}) // Assert as JSON object type
				owner := StrVal(x["appOwnerOrganizationId"])
				if owner == tenant_id {  // If owned by current tenant ...
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

func SpsCountAzure() (microsoft, native int64) {
	// Dedicated SPs Azure counter able to discern if SP is owned by native tenant or it's a Microsoft default SP
	// NOTE: Not entirely accurate yet because function GetAllObjects still checks local cache. Need to refactor
	// that function into 2 diff versions GetAllObjectsLocal and GetAllObjectsAzure and have this function the latter.
	var microsoftList []interface{} = nil
	var nativeList []interface{} = nil
	sps := GetAllObjects("sp")
	if sps != nil {
		for _, i := range sps {
			x := i.(map[string]interface{}) // Assert as JSON object type
			owner := StrVal(x["appOwnerOrganizationId"])
			if owner == tenant_id {  // If owned by current tenant ...
				nativeList = append(nativeList, x)
			} else {
				microsoftList = append(microsoftList, x)
			}
		}
	}
	return int64(len(microsoftList)), int64(len(nativeList))
}

func RoleDefinitionCountLocal() (builtin, custom int64) {
	// Dedicated role definition local cache counter able to discern if role is custom to native tenant or it's an Azure BuilIn role
	var customList []interface{} = nil
	var builtinList []interface{} = nil
	localData := filepath.Join(confdir, tenant_id+"_"+oMap["d"]+".json")
    if FileUsable(localData) {
		l, _:= LoadFileJson(localData) // Load cache file
		if l != nil {
			definitions := l.([]interface{}) // Assert as JSON array type
			for _, i := range definitions {
				x := i.(map[string]interface{}) // Assert as JSON object type
				xProp := x["properties"].(map[string]interface{})
				Type := StrVal(xProp["type"])
				if Type == "CustomRole" {
					customList = append(customList, x)
				} else {
					builtinList = append(builtinList, x)
				}
			}			
			return int64(len(builtinList)), int64(len(customList))
		}
	}
	return 0, 0
}

func RoleDefinitionCountAzure() (builtin, custom int64) {
	// Dedicated role definition Azure counter able to discern if role is custom to native tenant or it's an Azure BuilIn role
	var customList []interface{} = nil
	var builtinList []interface{} = nil
	quiet := false
	definitions := GetAzRoleDefinitionAll(quiet)
	for _, i := range definitions {
		x := i.(map[string]interface{}) // Assert as JSON object type
		xProp := x["properties"].(map[string]interface{})
		Type := StrVal(xProp["type"])
		if Type == "CustomRole" {
			customList = append(customList, x)
		} else {
			builtinList = append(builtinList, x)
		}
	}			
	return int64(len(builtinList)), int64(len(customList))
}


func ObjectCountLocal(t string) int64 {
	var oList []interface{} = nil // Start with an empty list
	localData := filepath.Join(confdir, tenant_id+"_"+oMap[t]+".json") // Define local data store file
    if FileUsable(localData) {
		l, _ := LoadFileJson(localData) // Load cache file
		if l != nil {
			oList = l.([]interface{})
			return int64(len(oList))
		}
	}
	return 0
}

func ObjectCountAzure(t string) int64 {
	// Returns count of given object type (ARM or MG)
	switch t {
	case "d", "a", "s", "m":
		// Azure Resource Management (ARM) API does not have a dedicated '$count' object filter,
		// so we're forced to retrieve all objects then count them. For simplicity, it is
		// best to have a dedicate function that handles all that.
		return int64(len(GetAllObjects(t)))
	case "u", "g", "sp", "ap", "ra":
		// MS Graph API makes counting much easier with its dedicated '$count' filter
		mg_headers["ConsistencyLevel"] = "eventual"
		url := mg_url + "/v1.0/" + oMap[t] + "/$count"
		r := ApiGet(url, mg_headers, nil)
		if r["value"] != nil { return r["value"].(int64) }  // Assert as int64
		// Expect result to be a single int64 value for the count
		ApiErrorCheck(r, trace())
	case "rd":
		// There is no $count filter option for AD role definitions so we have to get them all do length count
		url := mg_url + "/v1.0/roleManagement/directory/roleDefinitions"
		r := ApiGet(url, mg_headers, nil)
		if r["value"] != nil {
			rds := r["value"].([]interface{}) // Assert as JSON array type
			return int64(len(rds))
		}
		ApiErrorCheck(r, trace())
	}
	return 0
}

func GetAzObjectById(t, id string) (x map[string]interface{}) {
	// Retrieve Azure object by Object Id
	x = nil
	switch t {
	case "d", "a":
		scopes := GetAzRbacScopes()  // Look for objects under all the RBAC hierarchy scopes
		// Add null string below to represent the root "/" scope, else we miss any role assignments for it
		scopes = append(scopes, "")
		params := map[string]string{
			"api-version": "2022-04-01",  // roleDefinitions and roleAssignments
		}
		for _, scope := range scopes {
			url := az_url + scope + "/providers/Microsoft.Authorization/" + oMap[t] + "/" + id
			r := ApiGet(url, az_headers, params)
			if r != nil && r["id"] != nil { return r }  // Returns as soon as we find a match
			//ApiErrorCheck(r, trace()) // # DEBUG
		}
	case "s":
		params := map[string]string{
			"api-version": "2022-09-01",  // subscriptions
		}
		url := az_url + "/subscriptions/" + id
		r := ApiGet(url, az_headers, params)
		ApiErrorCheck(r, trace())
		x = r
	case "m":
		params := map[string]string{
			"api-version": "2022-04-01",  // managementGroups
		}
		url := az_url + "/providers/Microsoft.Management/managementGroups/" + id
		r := ApiGet(url, az_headers, params)
		ApiErrorCheck(r, trace())
		x = r
	case "u", "g", "ra":
		url := mg_url + "/v1.0/" + oMap[t] + "/" + id
		r := ApiGet(url, mg_headers, nil)
		ApiErrorCheck(r, trace())
		x = r
	case "ap", "sp":
		url := mg_url + "/v1.0/" + oMap[t]
		r := ApiGet(url + "/" + id, mg_headers, nil)  // First search is for direct Object Id
		if r != nil && r["error"] != nil {
			// Also look for this app or SP using its App/Client Id
			params := map[string]string{
				"$filter": "appId eq '" + id + "'",
			}
			r := ApiGet(url, mg_headers, params)
			if r != nil && r["value"] != nil {
				list := r["value"].([]interface{})
				count := len(list)
				if count == 1 {
					x = list[0].(map[string]interface{})  // Return single value found
					return x
				} else if count > 1 {
					print("Found %d entries with this appId\n", count)  // Not sure this would ever happen, but just in case
					return nil
				} else {
					return nil
				}
			}
			//ApiErrorCheck(r, trace())  // DEBUG
		}
		x = r
	case "rd":
		// Again, AD role definitions are under a different area, until they are activated
		url := mg_url + "/v1.0/roleManagement/directory/roleDefinitions/" + id
		r := ApiGet(url, mg_headers, nil)
		ApiErrorCheck(r, trace())
		x = r
	}
	return x
}

func GetAzObjectByName(t, name string) (x map[string]interface{}) {
	// FUTURE, not yet in use

	// Retrieve Azure object by displayName, given its type t
	switch t {
	case "a":
		return nil // Role assignments don't have a displayName attribute
	case "d":
		scopes := GetAzRbacScopes()  // Look for objects under all the RBAC hierarchy scopes
		params := map[string]string{
			"api-version": "2022-04-01",  // roleDefinitions
			"$filter":     "roleName eq '" + name + "'",
		}
		for _, scope := range scopes {
			url := az_url + scope + "/providers/Microsoft.Authorization/roleDefinitions"
			r := ApiGet(url, az_headers, params)
			if r != nil && r["value"] != nil {
				results := r["value"].([]interface{})  // Assert as JSON array type
				// NOTE: Would results ever be an array with MORE than 1 element? Is below name
				// confirmation even needed? Can we just return r["value"][0]
				for _, i := range results {
					x := i.(map[string]interface{})    // Assert as JSON object type
					xProps := x["properties"].(map[string]interface{})
					roleName := StrVal(xProps["roleName"])
					if roleName == name { return x }
					// Return first match we find, since roleName are unique across the tenant
				}
			}
			ApiErrorCheck(r, trace())
		}
	case "s":
		//x = ApiGet(az_url+"/"+oMap[t]+"/"+id + "?api-version=2022-04-01", az_headers, nil)
	case "m":
		//x = ApiGet(az_url+"/providers/Microsoft.Management/managementGroups/"+id + "?api-version=2022-04-01", az_headers, nil)
	case "u", "g", "sp", "ap", "ra":
		//x = ApiGet(mg_url+"/v1.0/"+oMap[t]+"/"+id, mg_headers, nil)
	case "rd":
		// Again, AD role definitions are under a different area, until they are activated
		//x = ApiGet(mg_url+"/v1.0/roleManagement/directory/roleDefinitions/"+id, mg_headers, nil)
	}
	return nil
}

func ApiGet(url string, headers, params map[string]string) (result map[string]interface{}) {
	// Basic, without debugging
	return ApiCall("GET", url, nil, headers, params, false)  // Verbose = false
}

func ApiGetDebug(url string, headers, params map[string]string) (result map[string]interface{}) {
	// Sets verbose boolean to true
	return ApiCall("GET", url, nil, headers, params, true)  // Verbose = true
}

func ApiCall(method, url string, jsonObj map[string]interface{}, headers, params map[string]string, verbose bool) (result map[string]interface{}) {
	// Make API call and return JSON object. Global az_headers and mg_headers are merged with additional ones called with.

	if !strings.HasPrefix(url, "http") { die(trace() + "Error: Bad URL, " + url + "\n") }

	// Set up headers according to API being called (AZ or MG)
	if strings.HasPrefix(url, az_url) {
		headers = MergeMaps(az_headers, headers)
	} else if strings.HasPrefix(url, mg_url) {
		headers = MergeMaps(mg_headers, headers)
	}

	// Set up new HTTP client
	client := &http.Client{Timeout: time.Second * 60} // One minute timeout
	var req *http.Request = nil
	var err error = nil
	switch strings.ToUpper(method) {
	case "GET":
		req, err = http.NewRequest("GET", url, nil)
	case "POST":
		jsonData, ok := json.Marshal(jsonObj)
		if ok != nil { panic(err.Error()) }
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	case "PUT":
		jsonData, ok := json.Marshal(jsonObj)
		if ok != nil { panic(err.Error()) }
		req, err = http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	case "DELETE":
		req, err = http.NewRequest("DELETE", url, nil)
	default:
		die(trace() + "Error: Unsupported HTTP method\n")
	}
	if err != nil { panic(err.Error()) }

	// Merge additional headers
	for h, v := range headers {
		req.Header.Add(h, v)
	}

	// Merge additional query parameters and encode
	q := req.URL.Query()
	for p, v := range params {
		q.Add(p, v)
	}
	req.URL.RawQuery = q.Encode()

	// === MAKE THE CALL ============
	if verbose {
		print("==== REQUEST =================================\n")
		print("GET " + url + "\n")
		print("HEADERS:\n")
		PrintJson(req.Header); print("\n") 
		print("PARAMS:\n")
		PrintJson(q); print("\n") 
		// print("REQUEST_PAYLOAD:\n")
		// PrintJson(BODY); print("\n") 
	}
	r, err := client.Do(req)
	if err != nil { panic(err.Error()) }
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil { panic(err.Error()) }
	
	// Note that variable 'body' is of type []uint8 which is essentially a long string
	// that evidently can be either A) a count integer number, or B) a JSON object string.
	// This interpretation needs confirmation, and then better handling.

	// Also, when dealing with unknown JSON object that are returned, one can assert the result to one
	// of the types below then parsing the values accordingly (see https://eager.io/blog/go-and-json/)
	//   nil                     for JSON null
	//   bool                    for JSON boolean
	//   string                  for JSON string
	//   float64                 for JSON number
	//   map[string]interface{}  for JSON object
	//   []interface{}           for JSON array

	if count, err := strconv.ParseInt(string(body), 10, 64); err == nil {
		// If entire body is a string representing an integer value, create
		// a JSON object with this count value we just converted to int64
		result = make(map[string]interface{})
		result["value"] = count
	} else {
		// Alternatively, treat entire body as a JSON object string, and unmarshalled into 'result'
		if err = json.Unmarshal([]byte(body), &result); err != nil {
			panic(err.Error())
		}
	}
	if verbose {
		print("==== RESPONSE ================================\n")
	    print("STATUS: %d %s\n", r.StatusCode, http.StatusText(r.StatusCode)) 
		print("RESULT:\n")
		PrintJson(result); print("\n") 
		resHeaders, err := httputil.DumpResponse(r, false)
		if err != nil { panic(err.Error()) }
		print("HEADERS:\n%s\n", string(resHeaders)) 
	}
	return result
}

func ApiErrorCheck(r map[string]interface{}, caller string) {
	if r["error"] != nil {
		e := r["error"].(map[string]interface{})
		print(caller + "Error: " + e["message"].(string) + "\n")
	}
}
