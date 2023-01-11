// subscriptions.go

package main

import (
	"fmt"
	"path/filepath"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintSubscription(x JsonObject) {
	// Print subscription object in YAML
	if x == nil { return }
	list := []string{"displayName", "subscriptionId", "state", "tenantId"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" { fmt.Printf("%-15s %s\n", i+":", v) } // Only print non-null attributes
	}
}

func GetAzSubscriptions(z aza.AzaBundle) (list JsonArray) {
	// Get ALL subscription in current Azure tenant AND save them to local cache file
	list = nil // We have to zero it out
	params := aza.MapString{"api-version": "2022-09-01"} // subscriptions
	url := aza.ConstAzUrl + "/subscriptions"
	r := ApiGet(url, z.AzHeaders, params)
	ApiErrorCheck(r, utl.Trace())
	if r != nil && r["value"] != nil {
		objects := r["value"].([]interface{})
		list = append(list, objects...)
	}
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_subscriptions.json")
	utl.SaveFileJson(list, cacheFile) // Update the local cache
	return list
}

func GetSubscriptions(filter string, force bool, z aza.AzaBundle) (list JsonArray) {
	// Get all subscriptions that match on provided filter. An empty "" filter means return
	// all subscription objects. It always uses local cache if it's within the cache retention
	// period, else it gets them from Azure. Also gets them from Azure if force is specified.
	// TODO: Make this cache retention period a configurable variable.
	list = nil
	cachePeriod := int64(3660 * 24 * 1) // 1 day cache retention period 
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_subscriptions.json")
	cacheNoGood, list := CheckLocalCache(cacheFile, cachePeriod)
	if cacheNoGood || force {
		list = GetAzSubscriptions(z) // Get the entire set from Azure
	}

	// Do filter matching
	if filter == "" {
		return list
	}
	var matchingList []interface{} = nil
	for _, i := range list { // Parse every object
		x := i.(map[string]interface{})
		// Match against relevant subscription attributes
		searchList := []string{"displayName", "subscriptionId", "state"}
		for _, i := range searchList {
			if utl.SubString(StrVal(x[i]), filter) {
				matchingList = append(matchingList, x)
			}
		}
	}
	return matchingList
}
