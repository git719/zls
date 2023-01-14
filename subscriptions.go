// subscriptions.go

package main

import (
	"fmt"
	"path/filepath"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintSubscription(x map[string]interface{}) {
	// Print subscription object in YAML-like
	if x == nil { return }
	list := []string{"displayName", "subscriptionId", "state", "tenantId"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" { fmt.Printf("%s: %s\n", i, v) } // Only print non-null attributes
	}
}

func SubsCountLocal(z aza.AzaBundle) (int64) {
	var cachedList []interface{} = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_subscriptions.json")
    if utl.FileUsable(cacheFile) {
		rawList, _ := utl.LoadFileJson(cacheFile)
		if rawList != nil {
			cachedList = rawList.([]interface{})
			return int64(len(cachedList))
		}
	}
	return 0
}

func SubsCountAzure(z aza.AzaBundle) (int64) {
	list := GetAzSubscriptions(z)
	return int64(len(list))
}

func GetAzSubscriptionsIds(z aza.AzaBundle) (scopes []string) {
	// Get all subscription full IDs, i.e. "/subscriptions/UUID", which are commonly
	// used as scopes for Azure resource RBAC role definitions and assignments
	scopes = nil
	subscriptions := GetAzSubscriptions(z)
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

func GetSubscriptions(filter string, force bool, z aza.AzaBundle) (list []interface{}) {
	// Get all subscriptions that match on provided filter. An empty "" filter means return
	// all subscription objects. It always uses local cache if it's within the cache retention
	// period, else it gets them from Azure. Also gets them from Azure if force is specified.
	// TODO: Make cachePeriod a configurable variable
	list = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_subscriptions.json")
	cacheNoGood, list := CheckLocalCache(cacheFile, 604800) // cachePeriod = 1 week in seconds
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

func GetAzSubscriptions(z aza.AzaBundle) (list []interface{}) {
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

func GetAzSubscriptionById(id string, headers aza.MapString) (map[string]interface{}) {
	// Get Azure subscription by Object Id
	params := aza.MapString{"api-version": "2022-09-01"}  // subscriptions
	url := aza.ConstAzUrl + "/subscriptions/" + id
	r := ApiGet(url, headers, params)
	ApiErrorCheck(r, utl.Trace())
	return r
}
