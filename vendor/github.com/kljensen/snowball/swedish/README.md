Snowball Swedish
================

This package implements the Swedish language
[Snowball stemmer](http://snowball.tartarus.org/algorithms/swedish/stemmer.html).

## Implementation

The Swedish language stemmer comprises preprocessing and 3 steps.
Each of these is defined in a separate file in this
package.  All of the steps operate on a `SnowballWord` from the
`snowballword` package and *modify the word in place*.

## Caveats

None