Snowball Russian
================

This package implements the
[Russian language Snowball stemmer](http://snowball.tartarus.org/algorithms/russian/stemmer.html).

## Russian overview

Russian has 33 letters, 11 Vowels, 20 consonants
and 2 unpronounced signs.  The capital letters 
look the same as the lower case letters, with
the exception of cursive capital letter and
lower case.

## Implementation

The Russian language stemmer comprises preprocessing, a number of steps.
Each of these is defined in a separate file in this
package.  All of the steps operate on a `SnowballWord` from the
`snowballword` package and *modify the word in place*.

## Caveats

The [example vocabulary for the original Russian snowball stemmer](http://snowball.tartarus.org/algorithms/russian/voc.txt) contains the word "злейший", which means "worst" in English.
This word contains the adjectival suffix "ий" preceded by the superlative suffix "ейш".
The [output for the example vocabulary](http://snowball.tartarus.org/algorithms/russian/output.txt)
indicates that this word should be stemmed to "злейш".  However, this implementation stems
the word to "зл".
The [Python NLTK](https://github.com/nltk/nltk/blob/master/nltk/stem/snowball.py#L2879)
implementation also stems "злейший" to "зл".
It is unclear to me how the original snowball implementation would possibly produce "злейш".
So, I removed that word from the tests.