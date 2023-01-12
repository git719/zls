// groups.go

package main

import (
	"fmt"
	"path/filepath"
	"time"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintGroup(x JsonObject, z aza.AzaBundle, oMap MapString) {
	// Print group object in YAML-like format
	if x == nil {
		return
	}
	id := StrVal(x["id"])

	// First, print the most important attributes of this group
	list := []string{"displayName", "description", "id", "isAssignableRole", "isAssignableToRole", "mailEnabled", "mailNickname"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" {
			fmt.Printf("%s: %s\n", i, v) // Only print non-null attributes
		}
	}

	// Print owners of this group
	url := aza.ConstMgUrl + "/beta/groups/" + id + "/owners"
	r := ApiGet(url, z.MgHeaders, nil)
	ApiErrorCheck(r, utl.Trace())
	if r["value"] != nil {
		owners := r["value"].([]interface{}) // Assert as JSON array type
		if len(owners) > 0 {
			fmt.Printf("owners:\n")
			for _, i := range owners {
				o := i.(map[string]interface{}) // Assert as JSON object type
				fmt.Printf("  %-50s %s\n", StrVal(o["userPrincipalName"]), StrVal(o["id"]))
			}
		} else {
			fmt.Printf("%s: %s\n", "owners", "None")
		}
	}

	// Print groups & roles this group is a member of
	memberOf := GetObjectMemberOfs("g", id, z, oMap) // For this Group object
	PrintMemberOfs("g", memberOf)

	// Print members of this group
	url = aza.ConstMgUrl + "/beta/groups/" + id + "/members"
	r = ApiGet(url, z.MgHeaders, nil)
	ApiErrorCheck(r, utl.Trace())
	if r["value"] != nil {
		members := r["value"].([]interface{}) // Assert as JSON array type
		if len(members) > 0 {
			fmt.Printf("members:\n")
			for _, i := range members {
				m := i.(map[string]interface{}) // Assert as JSON object type
				Type, Name := "-", "-"
				Type = utl.LastElem(StrVal(m["@odata.type"]), ".")
				switch Type {
				case "group", "servicePrincipal":
					Name = StrVal(m["displayName"])
				default:
					Name = StrVal(m["userPrincipalName"])
				}
				fmt.Printf("  %-50s %s (%s)\n", Name, StrVal(m["id"]), Type)
			}
		} else {
			fmt.Printf("%s: %s\n", "members", "None")
		}
	}
}

func GroupsCountLocal(z aza.AzaBundle) (int64) {
	// Return number of entries in local cache file
	var cachedList JsonArray = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_groups.json")
	if utl.FileUsable(cacheFile) {
		rawList, _ := utl.LoadFileJson(cacheFile)
		if rawList != nil {
			cachedList = rawList.([]interface{})
			return int64(len(cachedList))
		}
	}
	return 0
}	

func GroupsCountAzure(z aza.AzaBundle) (int64) {
	// Return number of entries in Azure tenant
	z.MgHeaders["ConsistencyLevel"] = "eventual"
	url := aza.ConstMgUrl + "/v1.0/groups/$count"
	r := ApiGet(url, z.MgHeaders, nil)
	ApiErrorCheck(r, utl.Trace())
	if r["value"] != nil {
		return r["value"].(int64) // Expected result is a single int64 value for the count
	}
	return 0	
}

func GetGroups(filter string, force bool, z aza.AzaBundle) (list JsonArray) {
	// Get all Azure AD groups whose searchAttributes match on 'filter'. An empty "" filter returns all.
	// Uses local cache if it's less than 1hr old. The 'force' option forces calling Azure query.
	list = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_groups.json")
	cacheNoGood, list := CheckLocalCache(cacheFile, 3660) // cachePeriod = 1hr = 3600sec
	if cacheNoGood || force {
		list = GetAzGroups(cacheFile, z.MgHeaders, true) // Get all from Azure and show progress (verbose = true)
	}
	
	// Do filter matching
	if filter == "" {
		return list
	}
	var matchingList JsonArray = nil
	searchAttributes := []string{
		"id", "displayName", "userPrincipalName", "onPremisesSamAccountName",
		"onPremisesUserPrincipalName", "onPremisesDomainName",
	}
	var ids []string // Keep track of each unique objects to eliminate repeats
	for _, i := range list {
		x := i.(map[string]interface{})
		id := StrVal(x["id"])
		for _, i := range searchAttributes {
			if utl.SubString(StrVal(x[i]), filter) && !utl.ItemInList(id, ids) {
				matchingList = append(matchingList, x)
				ids = append(ids, id)
			}
		}
	}
	return matchingList	
}

func GetAzGroups(cacheFile string, headers aza.MapString, verbose bool) (list JsonArray) {
	// Get all Azure AD groups in current tenant AND save them to local cache file. Show progress if verbose = true.
	
	// We will first try doing a delta query. See https://docs.microsoft.com/en-us/graph/delta-query-overview
	deltaLinkFile := cacheFile[:len(cacheFile)-len(filepath.Ext(cacheFile))] + "_deltaLink.json"
	deltaAge := int64(time.Now().Unix()) - int64(utl.FileModTime(deltaLinkFile))
	baseUrl := aza.ConstMgUrl + "/v1.0/groups"
    // Get delta updates only when below selection of attributes are modified
	selection := "?$select=displayName,mailNickname,description,mailEnabled,isAssignableToRole"
	url := baseUrl + "/delta" + selection + "&$top=999"
	headers["Prefer"] = "return=minimal" // This tells API to focus only on specific 'select' attributes

	// But first, double-check the base set again to avoid running a delta query on an empty set
	listIsEmpty, list := CheckLocalCache(cacheFile, 3600) // cachePeriod = 1hr = 3600sec
	if  utl.FileUsable(deltaLinkFile) && deltaAge < (3660 * 24 * 27) && listIsEmpty == false {
		// Note that deltaLink file age has to be within 30 days (we do 27)
		tmpVal, _ := utl.LoadFileJson(deltaLinkFile)
		deltaLinkMap := tmpVal.(map[string]interface{})
		url = StrVal(deltaLinkMap["@odata.deltaLink"]) // Base URL is now the cached Delta Link
	}

	// Run generic looper function to retrieve all objects from Azure
	list = GetAzObjectsLooper(url, cacheFile, headers, verbose)

	return list
}

func GetAzGroupById(id string, headers aza.MapString) (list JsonObject) {
	// Get Azure AD group by UUID, with extended attributes
	baseUrl := aza.ConstMgUrl + "/v1.0/groups"
	selection := "?$select=id,createdDateTime,description,displayName,groupTypes,id,isAssignableToRole,"
	selection += "mail,mailNickname,onPremisesLastSyncDateTime,onPremisesProvisioningErrors,"
	selection += "onPremisesSecurityIdentifier,onPremisesSyncEnabled,renewedDateTime,securityEnabled,"
	selection += "securityIdentifier,memberOf,members,owners"
	url := baseUrl + "/" + id + selection
	r := ApiGet(url, headers, nil)
	ApiErrorCheck(r, utl.Trace())
	return r
}

func PrintPags(z aza.AzaBundle) {
	// List all Privileged Access Groups
	groups := GetGroups("", false, z) // Get all groups, false = not need to hit Azure
	for _, i := range groups {
		x := i.(map[string]interface{})
		if x["isAssignableToRole"] != nil {
			if x["isAssignableToRole"].(bool) {
				PrintTersely("g", x) // Pring group tersely
			}
		}
	}
}
