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

func ObjectCountLocal(t string, z aza.AzaBundle, oMap map[string]string) int64 {
	var cachedList JsonArray = nil
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

func ObjectCountAzure(t string, z aza.AzaBundle, oMap map[string]string) int64 {
	// Returns count of given object type (ARM or MG)
	switch t {
	case "d":
		// Azure Resource Management (ARM) API does not have a dedicated '$count' object filter,
		// so we're forced to retrieve all objects then count them
		roleDefinitions := GetRoleDefinitions("", true, false, z, oMap)
		// true = force querying Azure, false = quietly
		return int64(len(roleDefinitions))
	case "a":
		roleAssignments := GetRoleAssignments("", true, false, z, oMap)
		return int64(len(roleAssignments))
	case "s":
		subscriptions :=  GetSubscriptions("", true, z)
		return int64(len(subscriptions))
	case "m":
		mgGroups := GetMgGroups("", true, z)
		return int64(len(mgGroups))
	}
	return 0
}

func GetObjectById(t, id string, z aza.AzaBundle) (x map[string]interface{}) {
	// Retrieve Azure object by Object Id
	switch t {
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
	case "a":
		return GetAzRoleAssignmentById(id, z)
	case "s":
		return GetAzSubscriptionById(id, z.AzHeaders)
	case "u":
		return GetAzUserById(id, z.MgHeaders)
	case "g":
		return GetAzGroupById(id, z.MgHeaders)
	case "sp":
		return GetAzSpById(id, z.MgHeaders)
	case "ap":
		return GetAzAppById(id, z.MgHeaders)
	case "ad":
		return GetAzAdRoleById(id, z.MgHeaders)
	}
	return nil
}

func ApiGet(url string, headers, params aza.MapString) (result map[string]interface{}) {
	// Basic, without debugging
	return ApiCall("GET", url, nil, headers, params, false)  // Verbose = false
}

func ApiGetDebug(url string, headers, params aza.MapString) (result map[string]interface{}) {
	// Sets verbose boolean to true
	return ApiCall("GET", url, nil, headers, params, true)  // Verbose = true
}

func ApiCall(method, url string, jsonObj map[string]interface{}, headers, params aza.MapString, verbose bool) (result map[string]interface{}) {
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
		fmt.Printf("==== REQUEST =================================\n")
		fmt.Printf("GET " + url + "\n")
		fmt.Printf("HEADERS:\n")
		utl.PrintJson(req.Header); print("\n") 
		print("PARAMS:\n")
		utl.PrintJson(q); fmt.Println() 
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
		fmt.Printf("==== RESPONSE ================================\n")
	    fmt.Printf("STATUS: %d %s\n", r.StatusCode, http.StatusText(r.StatusCode)) 
		fmt.Printf("RESULT:\n")
		utl.PrintJson(result); fmt.Println() 
		resHeaders, err := httputil.DumpResponse(r, false)
		if err != nil { panic(err.Error()) }
		fmt.Printf("HEADERS:\n%s\n", string(resHeaders)) 
	}
	return result
}

func ApiErrorCheck(r map[string]interface{}, caller string) {
	if r["error"] != nil {
		e := r["error"].(map[string]interface{})
		fmt.Printf(caller + "Error: " + e["message"].(string) + "\n")
	}
}
