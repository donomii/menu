package main

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
)

var results []ResultRecordTransmittable
var searchString string
var statuses map[string]string

type ResultRecordTransmittable struct {
	Filename    string
	Line        string
	Fingerprint []string
	Sample      string
	Score       string
}

func Score(a SearchPrint, b ResultRecordTransmittable) int {
	score := 0
	//We can do much better than a nested loop here

	for _, vv := range b.Fingerprint {
		for _, v := range a.wanted {
			if v == vv {
				score += 1

			}

		}
		for _, v := range a.unwanted {
			if v == vv {
				score -= 1

			}

		}
	}
	//fmt.Println("----")
	return score
}

type FingerPrint []string

func RegSplit(text string, reg *regexp.Regexp) []string {
	/*
		indexes := reg.FindAllStringIndex(text, -1)
		laststart := 0
		result := make([]string, len(indexes)+1)
		for i, element := range indexes {
			result[i] = text[laststart:element[0]]
			laststart = element[1]
		}
		result[len(indexes)] = text[laststart:len(text)]
		return result
	*/

	text = reg.ReplaceAllString(text, " ")
	text = reg.ReplaceAllString(text, " ")
	text = reg.ReplaceAllString(text, " ")
	text = reg.ReplaceAllString(text, " ")
	text = reg.ReplaceAllString(text, " ")
	text = reg.ReplaceAllString(text, " ")
	text = reg.ReplaceAllString(text, " ")
	text = reg.ReplaceAllString(text, " ")
	text = reg.ReplaceAllString(text, " ")
	text = reg.ReplaceAllString(text, " ")

	result := strings.Split(text, " ")
	//log.Println("Split results:", result)
	return result

}

var FragsRegex = regexp.MustCompile(`(\s+|,+|;+|:+|-+|"+|'+|\.+|/+|\+|\\+|/+|_+|=+|}+|{+|>+|<+|!+|\)+|\(+)`) //regexp.MustCompile("(\\/+|\\.+|\\\\+|\\-+|\\_+|_+|\\\\(|\\\\|\\p{Z}+|\\p{C}|\\s+|:|,|\"|{|}|>|<|!|))")
var SearchFragsRegex = regexp.MustCompile(`(\s+|,+|;+|:+|"+|'+|\.+|/+|\+|_+|=+|}+|{+|>+|<+|!+|\)+|\(+)`)     //regexp.MustCompile("(\\/+|\\.+|\\\\+|\\-+|\\_+|_+|\\\\(|\\\\|\\p{Z}+|\\p{C}|\\s+|:|,|\"|{|}|>|<|!|))")

func MakeFingerprintFromData(aStr string) FingerPrint {
	//seps := []string{"\\\\", "\\.", " ", "\\(", "\\)", "/", "_", "\\b"} //\\b|\\p{Z}+|\\p{C}|\\s+|\\/+|\\.+|\\\\+|_+
	//return makeFingerprint(ReSplit(seps, strings.Fields(strings.ToLower(aStr))))

	return MakeFingerprint(RegSplit(strings.ToLower(aStr), FragsRegex))
}

func MakeFingerprint(fragments []string) FingerPrint {
	sort.Strings(fragments)
	frags := map[string]int{}

	index := 0
	for _, f := range fragments {

		if len(f) > 1 && len(f) < maxTagLength {

			key := f
			_, ok := frags[key]
			if !ok {
				frags[key] = index
				index = index + 1
			}
		} else {
			if len(f) > maxTagLength {
				//				s.LogChan["warning"] <- fmt.Sprintln("Rejecting tag as too long: ", f)
			}
		}

	}
	//log.Printf("Fingerprint of length: %v", index)
	fingerprint := make(FingerPrint, index)
	for k, i := range frags {
		//fingerprint = append(fingerprint, k)
		//log.Printf("Assigning to index %v", i)
		fingerprint[i] = k
	}

	return fingerprint
}

type SearchPrint struct {
	wanted   FingerPrint
	unwanted FingerPrint
}

func CalcRawScore(aStr string) (string, int) {
	lastChar := aStr[len(aStr)-1 : len(aStr)]
	//log.Printf("Last char: '%v' ", lastChar)
	if lastChar == "-" {
		return aStr[0 : len(aStr)-1], -1
	} else {
		return aStr, 1
	}
}

var maxTagLength = 100

func MakeSearchPrint(fragments []string) SearchPrint {

	//Pull this out into a separate function FIXME
	frags := map[string]int{}
	for _, f := range fragments {
		if len(f) > 1 && len(f) < maxTagLength {

			key, rawScore := CalcRawScore(f)

			frags[key] = rawScore

		} else {
			log.Println("Rejected tag as too short or too long:", f)
		}
	}
	searchP := SearchPrint{}
	for k, v := range frags {
		//fmt.Printf("k: %v, v: %v\n", k, v)
		if v > 0 {
			searchP.wanted = append(searchP.wanted, k)
			//fmt.Printf("Storing k: %v in wanted\n", k)
		} else {
			//fmt.Printf("Storing k: %v in unwanted\n", k)
			searchP.unwanted = append(searchP.unwanted, k)
		}
	}
	statuses["SearchPrint"] = fmt.Sprintf("%+v", searchP)
	return searchP
}
