Snowball English
================

This package implements the English language
[Snowball stemmer](http://snowball.tartarus.org/algorithms/english/stemmer.html).

## Implementation

The English language stemmer comprises preprocessing, a number of steps,
and postprocessing.  Each of these is defined in a separate file in this
package.  All of the steps operate on a `SnowballWord` from the
`snowballword` package and *modify the word in place*.

## Caveats

There is a single difference between this implementation and the original.
Here, all apostrophes on the left hand side of a word are stripped off before
the word is stemmed.  