package spanish

import (
	"github.com/kljensen/snowball/snowballword"
	"log"
)

// Step 1 is the removal of standard suffixes
//
func step1(word *snowballword.SnowballWord) bool {

	// Possible suffixes, longest first
	suffix, suffixRunes := word.FirstSuffix(
		"amientos", "imientos", "aciones", "amiento", "imiento",
		"uciones", "logías", "idades", "encias", "ancias", "amente",
		"adores", "adoras", "ución", "mente", "logía", "istas",
		"ismos", "ibles", "encia", "anzas", "antes", "ancia",
		"adora", "ación", "ables", "osos", "osas", "ivos", "ivas",
		"ista", "ismo", "idad", "icos", "icas", "ible", "anza",
		"ante", "ador", "able", "oso", "osa", "ivo", "iva",
		"ico", "ica",
	)

	isInR1 := (word.R1start <= len(word.RS)-len(suffixRunes))
	isInR2 := (word.R2start <= len(word.RS)-len(suffixRunes))

	// Deal with special cases first.  All of these will
	// return if they are hit.
	//
	switch suffix {
	case "":

		// Nothing to do
		return false

	case "amente":

		if isInR1 {
			// Delete if in R1
			word.RemoveLastNRunes(len(suffixRunes))

			// if preceded by iv, delete if in R2 (and if further preceded by at,
			// delete if in R2), otherwise,
			// if preceded by os, ic or ad, delete if in R2
			newSuffix, _ := word.RemoveFirstSuffixIfIn(word.R2start, "iv", "os", "ic", "ad")
			if newSuffix == "iv" {
				word.RemoveFirstSuffixIfIn(word.R2start, "at")
			}
			return true
		}
		return false
	}

	// All the following cases require the found suffix
	// to be in R2.
	if isInR2 == false {
		return false
	}

	// Compound replacement cases.  All these cases return
	// if they are hit.
	//
	compoundReplacement := func(otherSuffixes ...string) bool {
		word.RemoveLastNRunes(len(suffixRunes))
		word.RemoveFirstSuffixIfIn(word.R2start, otherSuffixes...)
		return true
	}

	switch suffix {
	case "adora", "ador", "ación", "adoras", "adores", "aciones", "ante", "antes", "ancia", "ancias":
		return compoundReplacement("ic")
	case "mente":
		return compoundReplacement("ante", "able", "ible")
	case "idad", "idades":
		return compoundReplacement("abil", "ic", "iv")
	case "iva", "ivo", "ivas", "ivos":
		return compoundReplacement("at")
	}

	// Simple replacement & deletion cases are all that remain.
	//
	simpleReplacement := func(repl string) bool {
		word.ReplaceSuffixRunes(suffixRunes, []rune(repl), true)
		return true
	}
	switch suffix {
	case "logía", "logías":
		return simpleReplacement("log")
	case "ución", "uciones":
		return simpleReplacement("u")
	case "encia", "encias":
		return simpleReplacement("ente")
	case "anza", "anzas", "ico", "ica", "icos", "icas",
		"ismo", "ismos", "able", "ables", "ible", "ibles",
		"ista", "istas", "oso", "osa", "osos", "osas",
		"amiento", "amientos", "imiento", "imientos":
		word.RemoveLastNRunes(len(suffixRunes))
		return true
	}

	log.Panicln("Unhandled suffix:", suffix)
	return false
}
