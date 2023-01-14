// users.go

package main

import (
	"fmt"
	"path/filepath"
	"time"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintUser(x map[string]interface{}, z aza.AzaBundle, oMap map[string]string) {
	// Print user object in YAML-like format
	if x == nil {
		return
	}
	id := StrVal(x["id"])

	// First, print the most important attributes for this user
	list := []string{"displayName", "id", "userPrincipalName", "mailNickname", "onPremisesSamAccountName",
		"onPremisesDomainName", "onPremisesUserPrincipalName"}
	for _, i := range list {
		v := StrVal(x[i])
		if v != "" { // Only print non-null attributes
			fmt.Printf("%s: %s\n", i, v)
		}
	}

	// Print other mails this user has configured
	if x["otherMails"] != nil {
		otherMails := x["otherMails"].([]interface{})
		if len(otherMails) > 0 {
			fmt.Printf("otherMails:\n")
			for _, i := range otherMails {
				email := i.(string)
				fmt.Printf("  %s\n", email)
			}
		} else {
			fmt.Printf("  %s: %s\n", "otherMails", "None")
		}
	}

	// Print groups & roles this group is a member of
	memberOf := GetObjectMemberOfs("u", id, z, oMap) // For this User object
	PrintMemberOfs("u", memberOf)
}

func UsersCountLocal(z aza.AzaBundle) (int64) {
	// Return number of entries in local cache file
	var cachedList []interface{} = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_users.json")
	if utl.FileUsable(cacheFile) {
		rawList, _ := utl.LoadFileJson(cacheFile)
		if rawList != nil {
			cachedList = rawList.([]interface{})
			return int64(len(cachedList))
		}
	}
	return 0
}	

func UsersCountAzure(z aza.AzaBundle) (int64) {
	// Return number of entries in Azure tenant
	z.MgHeaders["ConsistencyLevel"] = "eventual"
	url := aza.ConstMgUrl + "/v1.0/users/$count"
	r := ApiGet(url, z.MgHeaders, nil)
	ApiErrorCheck(r, utl.Trace())
	if r["value"] != nil {
		return r["value"].(int64) // Expected result is a single int64 value for the count
	}
	return 0	
}

func GetUsers(filter string, force bool, z aza.AzaBundle) (list []interface{}) {
	// Get all Azure AD users that match on 'filter'. An empty "" filter returns all.
	// Uses local cache if it's less than cachePeriod old. The 'force' option forces calling Azure query.
	list = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_users.json")
	cacheNoGood, list := CheckLocalCache(cacheFile, 86400) // cachePeriod = 1 day in seconds
	if cacheNoGood || force {
		list = GetAzUsers(cacheFile, z.MgHeaders, true) // Get all from Azure and show progress (verbose = true)
	}
	
	// Do filter matching
	if filter == "" {
		return list
	}
	var matchingList []interface{} = nil
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

func GetAzUsers(cacheFile string, headers aza.MapString, verbose bool) (list []interface{}) {
	// Get all Azure AD users in current tenant AND save them to local cache file. Show progress if verbose = true.
	
	// We will first try doing a delta query. See https://docs.microsoft.com/en-us/graph/delta-query-overview
	var deltaLinkMap map[string]string = nil
	deltaLinkFile := cacheFile[:len(cacheFile)-len(filepath.Ext(cacheFile))] + "_deltaLink.json"
	deltaAge := int64(time.Now().Unix()) - int64(utl.FileModTime(deltaLinkFile))

	baseUrl := aza.ConstMgUrl + "/v1.0/users"
	// Get delta updates only when below selection of attributes are modified
	selection := "?$select=displayName,mailNickname,userPrincipalName,onPremisesSamAccountName,"
	selection += "onPremisesDomainName,onPremisesUserPrincipalName"
	url := baseUrl + "/delta" + selection + "&$top=999"
	headers["Prefer"] = "return=minimal" // This tells API to focus only on specific 'select' attributes
	
	// But first, double-check the base set again to avoid running a delta query on an empty set
	listIsEmpty, list := CheckLocalCache(cacheFile, 86400) // cachePeriod = 1 day in seconds
	if  utl.FileUsable(deltaLinkFile) && deltaAge < (3660 * 24 * 27) && listIsEmpty == false {
		// Note that deltaLink file age has to be within 30 days (we do 27)
		tmpVal, _ := utl.LoadFileJson(deltaLinkFile)
		deltaLinkMap = tmpVal.(map[string]string)
		url = StrVal(deltaLinkMap["@odata.deltaLink"]) // Base URL is now the cached Delta Link URL
	}

    // Now go get azure objects using the updated URL (either a full query or a deltaLink query)
	var deltaSet []interface{} = nil
	deltaSet, deltaLinkMap = GetAzObjects(url, headers, verbose) // Run generic deltaSet retriever function

	// Save new deltaLink for future call, and merge newly acquired delta set with existing list
	utl.SaveFileJson(deltaLinkMap, deltaLinkFile)
	list = NormalizeCache(list, deltaSet) // Run our MERGE LOGIC with new delta set
	utl.SaveFileJson(list, cacheFile) // Update the local cache
	return list
}

func GetAzUserById(id string, headers aza.MapString) (map[string]interface{}) {
	// Get Azure user by UUID, with extended attributes
	baseUrl := aza.ConstMgUrl + "/v1.0/users"
	selection := "?$select=id,accountEnabled,createdDateTime,creationType,displayName,id,identities,"
	selection += "lastPasswordChangeDateTime,mail,mailNickname,onPremisesDistinguishedName,"
	selection += "onPremisesDomainName,onPremisesExtensionAttributes,onPremisesImmutableId,"
	selection += "onPremisesLastSyncDateTime,onPremisesProvisioningErrors,onPremisesSamAccountName,"
	selection += "onPremisesSecurityIdentifier,onPremisesSyncEnabled,onPremisesUserPrincipalName,"
	selection += "otherMails,securityIdentifier,surname,userPrincipalName"
	url := baseUrl + "/" + id + selection
	r := ApiGet(url, headers, nil)
	ApiErrorCheck(r, utl.Trace())
	return r
}
