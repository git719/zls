// objects.go

package main

import (
	"github.com/git719/utl"
)

func MergeMaps(m1, m2 MapString) (result MapString) {
	result = MapString{}
	for k, v := range m1 {
		result[k] = v
	}
	for k, v := range m2 {
		result[k] = v
	}
	return result
}

func MergeObjects(x, y JsonObject) (obj JsonObject) {
	// Merge JSON object y into x
	// NOTES:
	// 1. Non-recursive, only works attributes at first level
	// 2. If attribute exists in y, we assume it's new and x needs to be updated with it
	obj = x
	for k, v := range x { // Update existing x values with updated y values
		obj[k] = v
		if y[k] != nil {
			obj[k] = y[k]
		}
	}
	for k, _ := range y { // Add new y values to x
		if x[k] == nil {
			obj[k] = y[k]
		}
	}
	return obj
}

func SelectObject(id string, objSet JsonArray) (x JsonObject) {
	// Select JSON object with given ID from slice
	for _, obj := range objSet {
		x = obj.(map[string]interface{})
		objId := StrVal(x["id"])
		if id == objId { return x }
	}
	return nil
}

func NormalizeCache(baseSet, deltaSet JsonArray) (oList JsonArray) {
	// Build JSON mergeSet from deltaSet, and build list of deleted IDs
	var deletedIds []string
	var uniqueIds []string
	var mergeSet JsonArray = nil
	for _, i := range deltaSet {
		x := i.(map[string]interface{})
		id := StrVal(x["id"])
		if x["@removed"] == nil && x["members@delta"] == nil {
			// Only add to mergeSet if '@remove' and 'members@delta' are missing
			if !utl.ItemInList(id, uniqueIds) {
				// Only add if it's unique
				mergeSet = append(mergeSet, x)
				uniqueIds = append(uniqueIds, id) // Track unique IDs
			}
		} else {
			deletedIds = append(deletedIds, id)
		}
	}

	// Remove recently deleted entries (deletedIs) from baseSet
	oList = nil
	var baseIds []string = nil // Track all the IDs in the base cache set
	for _, i := range baseSet {
		x := i.(map[string]interface{})
		id := StrVal(x["id"])
		if utl.ItemInList(id, deletedIds) { continue }
		oList = append(oList, x)
		baseIds = append(baseIds, id)
	}

	// Merge new entries in deltaSet into baseSet
	var duplicates JsonArray = nil
	var duplicateIds []string = nil
	for _, obj := range mergeSet {
		x := obj.(map[string]interface{})
		id := StrVal(x["id"])
		if utl.ItemInList(id, baseIds) {
			duplicates = append(duplicates, x)
			duplicateIds = append(duplicateIds, id)
			continue // Skip duplicates (these are updates)
		}
		oList = append(oList, x) // Merge all others (these are new entries)
	}

	// Merge updated entries in deltaSet into baseSet
	oList2 := oList
	oList = nil
	for _, obj := range oList2 {
		x := obj.(map[string]interface{})
		id := StrVal(x["id"])
		if !utl.ItemInList(id, duplicateIds) {
			// If this object is not a duplicate, add it to our growing list
			oList = append(oList, x)
		} else {
			// Merge object updates, then add it to our growing list
			y := SelectObject(id, duplicates)
			x = MergeObjects(x, y)
			oList = append(oList, x)
		}
	}

	return oList
}
