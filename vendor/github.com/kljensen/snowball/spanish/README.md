Snowball Spanish
================

This package implements the
[Spanish language Snowball stemmer](http://snowball.tartarus.org/algorithms/spanish/stemmer.html).

## Implementation

The Spanish language stemmer comprises preprocessing, a number of steps,
and postprocessing.  Each of these is defined in a separate file in this
package.  All of the steps operate on a `SnowballWord` from the
`snowballword` package and *modify the word in place*.

## Caveats

None yet.