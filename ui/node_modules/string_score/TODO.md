# TODO
* Add global config... ability to set adjust weights or change 
  ranges 0..1 to 1..100 or some such as well as set turn on fuzziness.
* Update tests to be more story like to ensure good coverage.
* qunit submodule is broken. > Need to remove submodule and add files back in.
* Write a Jaro-Winkler string score in JS to compare see: http://bit.ly/e5khQC & http://bit.ly/ecvzNy 
* Play nice with accented letters


## Accented Letters Example
    "Épendre".string_score("Ependre") => 0   // expecting 1 or nearly 1  É should be no more different from E then e
    "Épendre".string_score("Ependre", 0.5) => 0.009523809523809525 // expecting 1 or nearly 1
    "Amirauté".string_score("Amiraute") => 0
    "Amirauté".string_score("Amiraute", 0.5) => 0.6166666666666666