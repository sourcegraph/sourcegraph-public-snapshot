# What is it

* Simple - Adds a .score() method to the JavaScript String object... "String".score("str");
* Fast - fastest that I can find, often drastically faster... run the tests yourself
* Small - We are talking (520 Bytes)
* Portable - Works in 100% of the browsers I've tested on multiple platforms
* Independent - Doesn't require any other JavaScript - should work with any framework.
* Tested - Not everyone writes tests (silly people). Testing using Qunit
* Proper - Passes jslint as well as meets the coding practices and principles of opinionated developers :-)
* Fuzzyable - Optional parameter for fuzziness (allows mismatched info to still score the string)

# See it in action
Check it out [http://joshaven.com/string_score](http://joshaven.com/string_score)

## Installation Notes
Simply include one of the string score JavaScript files and call the .score() method on any string.

### NodeJS Installation
    npm install --save string_score
    require("string_score");

Thats it! It will automatically add a .score() method to all JavaScript String object... "String".score("str");

# Examples
(results are for example only... I may change the scoring algorithm without updating examples)

    "hello world".score("axl") //=> 0
    "hello world".score("ow")  //=> 0.35454545454545455

    "hello world".score("e")           //=>0.1090909090909091 (single letter match)
    "hello world".score("h")           //=>0.5363636363636364 (single letter match plus bonuses for beginning of word and beginning of phrase)
    "hello world".score("he")          //=>0.5727272727272728
    "hello world".score("hel")         //=>0.6090909090909091
    "hello world".score("hell")        //=>0.6454545454545455
    "hello world".score("hello")       //=>0.6818181818181818
    ...
    "hello world".score("hello worl")  //=>0.8636363636363635
    "hello world".score("hello world") //=> 1

    // And then there is fuzziness
    "hello world".score("hello wor1")  //=>0  (the "1" in place of the "l" makes a mismatch)
    "hello world".score("hello wor1",0.5)  //=>0.6081818181818182 (fuzzy)

    // Considers string length
    'Hello'.score('h') //=>0.52
    'He'.score('h')    //=>0.6249999999999999  (better match becaus string length is closer)

    // Same case matches better than wrong case
    'Hello'.score('h') //=>0.52
    'Hello'.score('H') //=>0.5800000000000001

    // Acronyms are given a little more weight
    "Hillsdale Michigan".score("HiMi") > "Hillsdale Michigan".score("Hills")
    "Hillsdale Michigan".score("HiMi") < "Hillsdale Michigan".score("Hillsd")

# Tested And Operational Under these environments

Fully functional in the 100% of the tested browsers:

* Firefox 3 & Newer (Mac & Windows)
* Safari 4 & Newer (Mac & Windows)
* IE: 7 & Newer (Windows) **
* Chrome: 2 & Newer (Windows)
* Opera: 9.64 & Newer (Windows)

** IE 7 fails (stop running this script message) with 4000 iterations
of the benchmark test. All other browsers tested survived this test,
and in fact survive a larger number of iterations.  The benchmark
that is causing IE to choke is: 4000 iterations of 446 character
string scoring a 70 character match.

# Benchmarks
This is the fastest and smallest javascript string scoring plugin
that I am aware of.  I have taken great joy in squeezing every
millisecond I can out of this script.  If you are aware of any
ways to improve this script, please let me know.

string_score.js is faster and smaller and does more than either liquidmetal.js or quicksilver.js

The test: 4000 iterations of 446 character string scoring a 70-character match

* string_score.js:
  * Firefox 3.6 (805ms)
  * Firefox 4 (245ms)
  * Chrome 9 (268ms)
  * Safari 5 (259ms)
* liquidmetal.js:
  * Firefox 3.6 (1578ms)
  * Firefox 4 (853ms)
  * Chrome 9 (339ms)
  * Safari 5 (996ms)
* quicksilver.js:
  * Firefox 3.6 (3300ms)
  * Firefox 4 (1994ms)
  * Chrome 9 (2835ms)
  * Safari 5 (3252ms)
* fuzzy_string.js
  * Firefox 4 (OUCH! I am not sure it heats up my laptop and asks if I want to stop the script... fuzzy_string, nice idea but it doesn't like large strings matches.)

** Tests run with jQuery 1.5 on Mac Book Pro 2.4GHz Core 2 Duo running Snow Leopard
*** quicksilver & string_score both use the same test file because they are used in the
same way, LiquidMetal has to be called differently so the test file was modified to work
with the LiquidMetal Syntax.

# Ports
Please notify me of any ports so I can have them listed here.
Please also keep track of the string score version that you have ported from. For example, in your readme include a note like: ported from version 0.2

* C# port: [ScoreSharp Bruno Lara Tavares](https://github.com/bltavares/scoresharp)
* C port: [string_score by kurige](https://github.com/kurige/string_score)
* Python port: [stringslipper by Yesudeep Mangalapilly](https://github.com/gorakhargosh/stringslipper)
* Ruby ports:
  * [scorer by Matt Duncan](https://github.com/mrduncan/scorer)
  * [string_score_ruby by James Lindley](https://github.com/jlindley/string_score_ruby)
* Java: [string_score by Shingo Omura](https://github.com/everpeace/string-score)
* 4GL: [string_score by Antonio Pérez](https://github.com/skarcha/string_score)
* Objective-C [StringScore by Nicholas Bruning](https://github.com/thetron/StringScore)

# Notes
string_score.js does not have any external dependencies
other than a reasonably new browser.

The tests located in the tests folder rely on the files
located in the tests folder.

Please share your testing results with me if you are
able to test under an unlisted browser.

# Credits
Author [Joshaven Potter](mailto:yourtech@gmail.com)

Thank you Lachie Cox and Quicksilver for inspiration.

Special thanks to all who contribute... and if you're not listed here please email me.
##Contributors
[Yesudeep Mangalapilly](mailto:yesudeep@gmail.com) - Collaborator
Eric Celeste
[Matt Duncan](https://github.com/mrduncan)
[Bruno Lara Tavares](https://github.com/bltavares)

# License
Licensed under the [MIT license](http://www.opensource.org/licenses/mit-license.php).
