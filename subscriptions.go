// subscriptions.go

package main

func PrintSubscription(x map[string]interface{}) {
	// Print subscription object in YAML
	if x == nil { return }
	list := []string{"displayName", "subscriptionId", "state", "tenantId"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" { print("%-20s %s\n", i+":", v) } // Only print non-null attributes
	}
}

func GetAzSubscriptionAll() (oList []interface{}) {
	// Get all Azure subscriptions in this tenant
	oList = nil
	params := map[string]string{
		"api-version": "2022-09-01",  // subscriptions
	}
	url := az_url + "/subscriptions"
	r := ApiGet(url, az_headers, params)
	if r != nil && r["value"] != nil {
		objects := r["value"].([]interface{})  // Treat as JSON array type
		oList = append(oList, objects...)      // Elipsis means the source may be more than one
	}
	ApiErrorCheck(r, trace())
	return oList
}
