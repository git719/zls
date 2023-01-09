// api_calls.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintCountStatus(z aza.AzaBundle, oMap MapString) {
	fmt.Printf("Note: Counting objects residing in Azure can take some time.\n")
	fmt.Printf("%-38s %-20s %s\n", "OBJECTS", "LOCAL_CACHE_COUNT","AZURE_COUNT")
	// fmt.Printf("%-38s ", "Groups")
	// fmt.Printf("%-20d %d\n", ObjectCountLocal("g", z, oMap), ObjectCountAzure("g", z, oMap))
	// fmt.Printf("%-38s ", "Users")
	// fmt.Printf("%-20d %d\n", ObjectCountLocal("u", z, oMap), ObjectCountAzure("u", z, oMap))
	// fmt.Printf("%-38s ", "App Registrations")
	// fmt.Printf("%-20d %d\n", ObjectCountLocal("ap", z, oMap), ObjectCountAzure("ap", z, oMap))
	// microsoftSpsLocal, nativeSpsLocal := SpsCountLocal(z.ConfDir, z.TenantId)
	// microsoftSpsAzure, nativeSpsAzure := SpsCountAzure(z.TenantId, z.TenantId)
	// fmt.Printf("%-38s ", "Service Principals Microsoft Default")
	// fmt.Printf("%-20d %d\n", microsoftSpsLocal, microsoftSpsAzure)
	// fmt.Printf("%-38s ", "Service Principals This Tenant")
	// fmt.Printf("%-20d %d\n", nativeSpsLocal, nativeSpsAzure)
	// fmt.Printf("%-38s ", "Azure AD Roles Definitions")
	// fmt.Printf("%-20d %d\n", ObjectCountLocal("rd", z, oMap), ObjectCountAzure("rd", z, oMap))
	// fmt.Printf("%-38s ", "Azure AD Roles Activated")
	// fmt.Printf("%-20d %d\n", ObjectCountLocal("ra", z, oMap), ObjectCountAzure("ra", z, oMap))
	fmt.Printf("%-38s ", "Management Groups")
	fmt.Printf("%-20d %d\n", ObjectCountLocal("m", z, oMap), ObjectCountAzure("m", z, oMap))
	fmt.Printf("%-38s ", "Subscriptions")
	fmt.Printf("%-20d %d\n", ObjectCountLocal("s", z, oMap), ObjectCountAzure("s", z, oMap))
	builtinLocal, customLocal := RoleDefinitionCountLocal(z)
	builtinAzure, customAzure := RoleDefinitionCountAzure(z, oMap)
	fmt.Printf("%-38s ", "RBAC Role Definitions BuiltIn")
    fmt.Printf("%-20d %d\n", builtinLocal, builtinAzure)
	fmt.Printf("%-38s ", "RBAC Role Definitions Custom")
    fmt.Printf("%-20d %d\n", customLocal, customAzure)
	// fmt.Printf("%-38s ", "RBAC Role Assignments")
	// fmt.Printf("%-20d %d\n", ObjectCountLocal("a", z, oMap), ObjectCountAzure("a", z, oMap))
}

// func SpsCountLocal(confDir, tenantId string) (microsoft, native int64) {
// 	// Dedicated SPs local cache counter able to discern if SP is owned by native tenant or it's a Microsoft default SP 
// 	var microsoftList []interface{} = nil
// 	var nativeList []interface{} = nil
// 	localData := filepath.Join(confDir, tenantId + "_servicePrincipals.json")
//     if utl.FileUsable(localData) {
// 		l, _ := utl.LoadFileJson(localData) // Load cache file
// 		if l != nil {
// 			sps := l.([]interface{}) // Assert as JSON array type
// 			for _, i := range sps {
// 				x := i.(map[string]interface{}) // Assert as JSON object type
// 				owner := StrVal(x["appOwnerOrganizationId"])
// 				if owner == tenantId {  // If owned by current tenant ...
// 					nativeList = append(nativeList, x)
// 				} else {
// 					microsoftList = append(microsoftList, x)
// 				}
// 			}
// 			return int64(len(microsoftList)), int64(len(nativeList))
// 		}
// 	}
// 	return 0, 0
// }

// func SpsCountAzure(confDir, tenantId string) (microsoft, native int64) {
// 	// Dedicated SPs Azure counter able to discern if SP is owned by native tenant or it's a Microsoft default SP
// 	// NOTE: Not entirely accurate yet because function GetAllObjects still checks local cache. Need to refactor
// 	// that function into 2 diff versions GetAllObjectsLocal and GetAllObjectsAzure and have this function the latter.
// 	var microsoftList []interface{} = nil
// 	var nativeList []interface{} = nil
// 	sps := GetAllObjects("sp", confDir, tenantId)
// 	if sps != nil {
// 		for _, i := range sps {
// 			x := i.(map[string]interface{}) // Assert as JSON object type
// 			owner := StrVal(x["appOwnerOrganizationId"])
// 			if owner == tenantId {  // If owned by current tenant ...
// 				nativeList = append(nativeList, x)
// 			} else {
// 				microsoftList = append(microsoftList, x)
// 			}
// 		}
// 	}
// 	return int64(len(microsoftList)), int64(len(nativeList))
// }

func ObjectCountLocal(t string, z aza.AzaBundle, oMap MapString) int64 {
	var cachedList []interface{} = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_" + oMap[t] + ".json")
    if utl.FileUsable(cacheFile) {
		rawList, _ := utl.LoadFileJson(cacheFile)
		if rawList != nil {
			cachedList = rawList.([]interface{})
			return int64(len(cachedList))
		}
	}
	return 0
}

func ObjectCountAzure(t string, z aza.AzaBundle, oMap MapString) int64 {
	// Returns count of given object type (ARM or MG)
	switch t {
	case "d", "a", "s", "m":
		// Azure Resource Management (ARM) API does not have a dedicated '$count' object filter,
		// so we're forced to retrieve all objects then count them
		allObjects := GetObjects(t, "", true, z, oMap) // true = force a call to Azure
		return int64(len(allObjects))
	case "u", "g", "sp", "ap", "ra":
		// MS Graph API makes counting much easier with its dedicated '$count' filter
		z.MgHeaders["ConsistencyLevel"] = "eventual"
		url := aza.ConstMgUrl + "/v1.0/" + oMap[t] + "/$count"
		r := ApiGet(url, z.MgHeaders, nil)
		ApiErrorCheck(r, utl.Trace())
		if r["value"] != nil {
			return r["value"].(int64) // Expected result is a single int64 value for the count
		}		
	case "rd":
		// There is no $count filter option for adRoleDefinitions so we have to get them all and do a length count
		url := aza.ConstMgUrl + "/v1.0/roleManagement/directory/roleDefinitions"
		r := ApiGet(url, z.MgHeaders, nil)
		ApiErrorCheck(r, utl.Trace())
		if r["value"] != nil {
			adRoleDefinitions := r["value"].([]interface{})
			return int64(len(adRoleDefinitions))
		}
	}
	return 0
}

// func GetAzObjectById(t, id string, z aza.AzaBundle, oMap MapString) (x JsonObject) {
// 	// Retrieve Azure object by Object Id
// 	x = nil
// 	switch t {
// 	case "d", "a":
// 		scopes := GetAzRbacScopes(z.ConfDir, z.TenantId)  // Look for objects under all the RBAC hierarchy scopes
// 		// Add null string below to represent the root "/" scope, else we miss any role assignments for it
// 		scopes = append(scopes, "")
// 		params := aza.MapString{ "api-version": "2022-04-01"}  // roleDefinitions and roleAssignments
// 		for _, scope := range scopes {
// 			url := aza.ConstAzUrl + scope + "/providers/Microsoft.Authorization/" + oMap[t] + "/" + id
// 			r := ApiGet(url, z.AzHeaders, params)
// 			if r != nil && r["id"] != nil { return r }  // Returns as soon as we find a match
// 			//ApiErrorCheck(r, utl.Trace()) // # DEBUG
// 		}
// 	case "s":
// 		params := aza.MapString{"api-version": "2022-09-01"}  // subscriptions
// 		url := aza.ConstAzUrl + "/subscriptions/" + id
// 		r := ApiGet(url, z.AzHeaders, params)
// 		ApiErrorCheck(r, utl.Trace())
// 		x = r
// 	case "m":
// 		params := aza.MapString{"api-version": "2022-04-01"} // managementGroups
// 		url := aza.ConstAzUrl + "/providers/Microsoft.Management/managementGroups/" + id
// 		r := ApiGet(url, z.AzHeaders, params)
// 		ApiErrorCheck(r, utl.Trace())
// 		x = r
// 	case "u", "g", "ra":
// 		url := aza.ConstMgUrl + "/v1.0/" + oMap[t] + "/" + id
// 		r := ApiGet(url, z.MgHeaders, nil)
// 		ApiErrorCheck(r, utl.Trace())
// 		x = r
// 	case "ap", "sp":
// 		url := aza.ConstMgUrl + "/v1.0/" + oMap[t]
// 		r := ApiGet(url + "/" + id, z.MgHeaders, nil)  // First search is for direct Object Id
// 		if r != nil && r["error"] != nil {
// 			// Also look for this app or SP using its App/Client Id
// 			params := aza.MapString{"$filter": "appId eq '" + id + "'"}
// 			r := ApiGet(url, z.MgHeaders, params)
// 			if r != nil && r["value"] != nil {
// 				list := r["value"].([]interface{})
// 				count := len(list)
// 				if count == 1 {
// 					x = list[0].(map[string]interface{})  // Return single value found
// 					return x
// 				} else if count > 1 {
// 					print("Found %d entries with this appId\n", count)  // Not sure this would ever happen, but just in case
// 					return nil
// 				} else {
// 					return nil
// 				}
// 			}
// 			//ApiErrorCheck(r, utl.Trace())  // DEBUG
// 		}
// 		x = r
// 	case "rd":
// 		// Again, AD role definitions are under a different area, until they are activated
// 		url := aza.ConstMgUrl + "/v1.0/roleManagement/directory/roleDefinitions/" + id
// 		r := ApiGet(url, z.MgHeaders, nil)
// 		ApiErrorCheck(r, utl.Trace())
// 		x = r
// 	}
// 	return x
// }

// func GetAzObjectByName(t, name string, z aza.AzaBundle, oMap MapString) (x map[string]interface{}) {
// 	// FUTURE, not yet in use

// 	// Retrieve Azure object by displayName, given its type t
// 	switch t {
// 	case "a":
// 		return nil // Role assignments don't have a displayName attribute
// 	case "d":
// 		scopes := GetAzRbacScopes(z.ConfDir, z.TenantId)  // Look for objects under all the RBAC hierarchy scopes
// 		params := map[string]string{
// 			"api-version": "2022-04-01",  // roleDefinitions
// 			"$filter":     "roleName eq '" + name + "'",
// 		}
// 		for _, scope := range scopes {
// 			url := az_url + scope + "/providers/Microsoft.Authorization/roleDefinitions"
// 			r := ApiGet(url, az_headers, params)
// 			if r != nil && r["value"] != nil {
// 				results := r["value"].([]interface{})  // Assert as JSON array type
// 				// NOTE: Would results ever be an array with MORE than 1 element? Is below name
// 				// confirmation even needed? Can we just return r["value"][0]
// 				for _, i := range results {
// 					x := i.(map[string]interface{})    // Assert as JSON object type
// 					xProps := x["properties"].(map[string]interface{})
// 					roleName := StrVal(xProps["roleName"])
// 					if roleName == name { return x }
// 					// Return first match we find, since roleName are unique across the tenant
// 				}
// 			}
// 			ApiErrorCheck(r, utl.Trace())
// 		}
// 	case "s":
// 		//x = ApiGet(az_url+"/"+oMap[t]+"/"+id + "?api-version=2022-04-01", az_headers, nil)
// 	case "m":
// 		//x = ApiGet(az_url+"/providers/Microsoft.Management/managementGroups/"+id + "?api-version=2022-04-01", az_headers, nil)
// 	case "u", "g", "sp", "ap", "ra":
// 		//x = ApiGet(mg_url+"/v1.0/"+oMap[t]+"/"+id, mg_headers, nil)
// 	case "rd":
// 		// Again, AD role definitions are under a different area, until they are activated
// 		//x = ApiGet(mg_url+"/v1.0/roleManagement/directory/roleDefinitions/"+id, mg_headers, nil)
// 	}
// 	return nil
// }

func ApiGet(url string, headers, params aza.MapString) (result JsonObject) {
	// Basic, without debugging
	return ApiCall("GET", url, nil, headers, params, false)  // Verbose = false
}

func ApiGetDebug(url string, headers, params aza.MapString) (result JsonObject) {
	// Sets verbose boolean to true
	return ApiCall("GET", url, nil, headers, params, true)  // Verbose = true
}

func ApiCall(method, url string, jsonObj JsonObject, headers, params aza.MapString, verbose bool) (result JsonObject) {
	// Make API call and return JSON object. Global az_headers and mg_headers are merged with additional ones called with.

	if !strings.HasPrefix(url, "http") {
		utl.Die(utl.Trace() + "Error: Bad URL, " + url + "\n")
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
		utl.Die(utl.Trace() + "Error: Unsupported HTTP method\n")
	}
	if err != nil { panic(err.Error()) }

	// Set up the headers
	for h, v := range headers {
		req.Header.Add(h, v)
	}

	// Set up the query parameters and encode 
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
		utl.PrintJson(req.Header); print("\n") 
		print("PARAMS:\n")
		utl.PrintJson(q); print("\n") 
		// print("REQUEST_PAYLOAD:\n")
		// utl.PrintJson(BODY); print("\n") 
	}
	r, err := client.Do(req)
	if err != nil { panic(err.Error()) }
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil { panic(err.Error()) }
	
	// Note that variable 'body' is of type []uint8 which is essentially a long string
	// that evidently can be either A) a count integer number, or B) a JSON object string.
	// This interpretation needs confirmation, and then better handling.

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
		utl.PrintJson(result); print("\n") 
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
