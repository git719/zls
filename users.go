// users.go

package main

import (
	"fmt"
	"path/filepath"
	"time"
	"github.com/git719/aza"
	"github.com/git719/utl"
)

func PrintUser(x JsonObject, z aza.AzaBundle, oMap MapString) {
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

func GetUsers(filter string, force bool, z aza.AzaBundle) (list JsonArray) {
	// Get all Azure AD users that match on 'filter'. An empty "" filter returns all.
	// Uses local cache if it's less than 1hr old. The 'force' option forces calling Azure query.
	list = nil
	cacheFile := filepath.Join(z.ConfDir, z.TenantId + "_users.json")
	cacheNoGood, list := CheckLocalCache(cacheFile, 3660) // cachePeriod = 1hr = 3600sec
	if cacheNoGood || force {
		list = GetAzUsers(cacheFile, z.MgHeaders, true) // Get all from Azure and show progress (verbose = true)
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

func GetAzUsers(cacheFile string, headers aza.MapString, verbose bool) (list JsonArray) {
	// Get all Azure AD users in current tenant AND save them to local cache file. Show progress if verbose = true.
	
	// We will first try doing doing a delta query. See https://docs.microsoft.com/en-us/graph/delta-query-overview
	deltaLinkFile := cacheFile[:len(cacheFile)-len(filepath.Ext(cacheFile))] + "_deltaLink.json"
	deltaAge := int64(time.Now().Unix()) - int64(utl.FileModTime(deltaLinkFile))
	// Base URL below will retrieve only the attributes we care about using 'select'. Also adding 'top' paging option
	url := aza.ConstMgUrl + "/v1.0/users/delta?$select=displayName,mailNickname,userPrincipalName"
	url += ",onPremisesSamAccountName,onPremisesDomainName,onPremisesUserPrincipalName" + "&$top=999"
	headers["Prefer"] = "return=minimal" // Tell delta query to focus only on specific 'select' attributes

	// But first, double-check the base set again to avoid running a delta query on an empty set
	listIsEmpty, list := CheckLocalCache(cacheFile, 3600) // cachePeriod = 1hr = 3600sec
	if  utl.FileUsable(deltaLinkFile) && deltaAge < (3660 * 24 * 27) && listIsEmpty == false {
		// Note that deltaLink file age has to be within 30 days (we do 27)
		tmpVal, _ := utl.LoadFileJson(deltaLinkFile)
		deltaLinkMap := tmpVal.(map[string]interface{})
		url = StrVal(deltaLinkMap["@odata.deltaLink"]) // Base URL is now the cached Delta Link
	}

	// Run generic looper function to retrieve all objects from Azure
	list = GetAzObjectsLooper(url, cacheFile, deltaLinkFile, headers, verbose)

	return list
}

func GetAzUserById(id string, headers aza.MapString) (list JsonObject) {
	// Get Azure user by UUID, with extended attributes
	url := aza.ConstMgUrl + "/v1.0/users/" + id + "?$select=accountEnabled,createdDateTime,creationType,"
	url += "displayName,id,identities,lastPasswordChangeDateTime,mail,mailNickname,onPremisesDistinguishedName,"
	url += "onPremisesDomainName,onPremisesExtensionAttributes,onPremisesImmutableId,onPremisesLastSyncDateTime,"
	url += "onPremisesProvisioningErrors,onPremisesSamAccountName,onPremisesSecurityIdentifier,"
	url += "onPremisesSyncEnabled,onPremisesUserPrincipalName,otherMails,securityIdentifier,surname,userPrincipalName"
	r := ApiGet(url, headers, nil)
	ApiErrorCheck(r, utl.Trace())
	return r
}
