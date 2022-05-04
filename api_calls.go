// api_calls.go

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ObjectCount(t string) int64 {
	// Returns count of given object type (ARM or MG)
	switch t {
	case "d", "a", "s", "m":
		// Azure Resource Management (ARM) API does not have a dedicated '$count' object filter,
		// so we're forced to retrieve all objects then count them. For simplicity, it is
		// best to have a dedicate function that handles all that.
		return int64(len(GetAllObjects(t)))
	case "u", "g", "sp", "ap":
		// MS Graph API makes counting much easier with its dedicated '$count' filter
		mg_headers["ConsistencyLevel"] = "eventual"
		r := APIGet(mg_url+"/v1.0/"+oMap[t]+"/$count", mg_headers, nil, false)
		if r["value"] != nil {
			// Expect result to be a single int64 value for the count
			return r["value"].(int64) // Assert as int64
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
	case "u", "g", "sp", "ap":
		x = APIGet(mg_url+"/beta/"+oMap[t]+"/"+id, mg_headers, nil, false)
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
		fmt.Printf("\nGET %s => %d %s => %s\n\n", url, r.StatusCode, http.StatusText(r.StatusCode), string(body))
		// p, _ := Prettify(params)
		// h, _ := Prettify(headers)
		// fmt.Println("REQUEST_PARAMS:", p)
		// fmt.Println("REQUEST_HEADERS:", h)
		// r_headers, err := httputil.DumpResponse(r, false)
		// if err != nil {
		// 	panic(err.Error())
		// }
		// fmt.Printf("RESPONSE_HEADERS:\n%s\n", string(r_headers))
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
