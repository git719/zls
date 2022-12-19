// api_calls.go

package main

import (
	"encoding/json"
	"io"
	"net/http"
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
		l := LoadFileJson(localData) // Load cache file
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
		l := LoadFileJson(localData) // Load cache file
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
	definitions := GetRoleDefinitions("silent")
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
		l := LoadFileJson(localData) // Load cache file
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
		r := APIGet(mg_url+"/v1.0/"+oMap[t]+"/$count", mg_headers, nil, false)
		if r["value"] != nil {
			// Expect result to be a single int64 value for the count
			return r["value"].(int64) // Assert as int64
		}
	case "rd":
		// There is no $count filter option for AD role definitions so we have to get them all do length count
		r := APIGet(mg_url+"/v1.0/roleManagement/directory/roleDefinitions", mg_headers, nil, false)
		if r["value"] != nil {
			rds := r["value"].([]interface{}) // Assert as JSON array type
			return int64(len(rds))
		}
	}
	return 0
}

func GetResourceGroupIds(subId string) (resGroupIds []string) {
	// Get the fully qualified ID of each Resource Group in given subscription ID
	resGroupIds = nil
	r := APIGet(az_url+"/subscriptions/"+subId+"/resourcegroups", az_headers, nil, false)
	if r["value"] != nil {
		resourceGroups := r["value"].([]interface{}) // Assert as JSON array type
		for _, obj := range resourceGroups {
			x := obj.(map[string]interface{}) // Assert as JSON object
			id := StrVal(x["id"])
			if id != "" {
				resGroupIds = append(resGroupIds, id)
			}
		}
	}
	return resGroupIds
}

func GetObjectById(t, id string) (x map[string]interface{}) {
	// Retrieve Azure object by UUID
	x = nil
	switch t {
	case "d", "a":
		// First, search for this object at Tenant root level
		url := az_url + "/providers/Microsoft.Management/managementGroups/" + tenant_id
		url += "/providers/Microsoft.Authorization/" + oMap[t] + "/" + id
		x = APIGet(url, az_headers, nil, false)
		if x["error"] != nil {
			// Finally, search for it under each Subscription scope
			for _, subId := range GetSubIds() {
				url = az_url + "/subscriptions/" + subId + "/providers/Microsoft.Authorization/" + oMap[t] + "/" + id
				x2 := APIGet(url, az_headers, nil, false)
				if x2["id"] != nil {
					x = x2
					break // As soon as we find it
				}
			}
		}
	case "s":
		x = APIGet(az_url+"/"+oMap[t]+"/"+id, az_headers, nil, false)
	case "m":
		x = APIGet(az_url+"/providers/Microsoft.Management/managementGroups/"+id, az_headers, nil, false)
	case "u", "g", "sp", "ap", "ra":
		x = APIGet(mg_url+"/v1.0/"+oMap[t]+"/"+id, mg_headers, nil, false)
	case "rd":
		// Again, AD role definitions are under a different area, until they are activated
		x = APIGet(mg_url+"/v1.0/roleManagement/directory/roleDefinitions/"+id, mg_headers, nil, false)
	}
	return x
}

func APIGet(url string, headers, params map[string]string, verbose bool) (result map[string]interface{}) {
	// Make API call and return JSON object
	// This functions adds default headers and params for the respective AZ or MG APIs, so the expected
	// headers and params options are for ADDITIONAL values of these.

	// The unknown JSON object that's returned can then be parsed by asserting any of the following types
	// and checking values ( see https://eager.io/blog/go-and-json/ )
	//   nil                     for JSON null
	//   bool                    for JSON boolean
	//   string                  for JSON string
	//   float64                 for JSON number
	//   map[string]interface{}  for JSON object
	//   []interface{}           for JSON array

	// Set up parameters and headers according to API being called (AZ or MG)
	if strings.HasPrefix(url, az_url) {
		az_params := map[string]string{"api-version": "2018-07-01"}
		params = MergeMaps(az_params, params)
		headers = MergeMaps(az_headers, headers)
	} else if strings.HasPrefix(url, mg_url) {
		headers = MergeMaps(mg_headers, headers)
	}

	// Set up new HTTP client
	client := &http.Client{Timeout: time.Second * 60} // One minute timeout
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err.Error())
	}

	// Update headers
	for h, v := range headers {
		req.Header.Add(h, v)
	}

	// Update query parameters
	q := req.URL.Query()
	for p, v := range params {
		q.Add(p, v)
	}
	req.URL.RawQuery = q.Encode()

	// === Make the call ============
	r, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err.Error())
	}
	// Note that variable 'body' is of type []uint8 which is essentially a long string
	// that evidently can be either A) a count integer number, or B) a JSON object string.
	// This interpretation needs confirmation, and then better handling.

	if verbose {
		print("\nGET %s => %d %s => %s\n\n", url, r.StatusCode, http.StatusText(r.StatusCode), string(body))
		// p, _ := Prettify(params)
		// h, _ := Prettify(headers)
		// print("REQUEST_PARAMS: %s\n", p)
		// print("REQUEST_HEADERS: %s\n", h)
		// r_headers, err := httputil.DumpResponse(r, false)
		// if err != nil {
		// 	panic(err.Error())
		// }
		// print("RESPONSE_HEADERS:\n%s\n", string(r_headers))
	}

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
	return result
}
