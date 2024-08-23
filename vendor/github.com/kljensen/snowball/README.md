Snowball
========


A [Go (golang)](http://golang.org) implementation of the
[Snowball stemmer](http://snowball.tartarus.org/)
for natural language processing.



|                      |  Status                   |
| -------------------- | ------------------------- |
| Latest release       |  [v0.6.0](https://github.com/kljensen/snowball/tags) (2018-10-29) |
| Latest build status  |  [![Build Status](https://travis-ci.org/kljensen/snowball.png)](https://travis-ci.org/kljensen/snowball)
| Latest Go versions tested   |  go1.10 darwin/amd64                 |
| Languages available  |  English, Spanish (español), French (le français), Russian (ру́сский язы́к), Swedish (svenska), Norwegian (norsk)|
| License              |  MIT                      |


## Usage


Here is a minimal Go program that uses this package in order
to stem a single word.

```go
package main
import (
	"fmt"
	"github.com/kljensen/snowball"
)
func main(){
	stemmed, err := snowball.Stem("Accumulations", "english", true)
	if err == nil{
		fmt.Println(stemmed) // Prints "accumul"
	}
}
```


## Organization & Implementation

The code is organized as follows:

* The top-level `snowball` package has a single exported function `snowball.Stem`,
  which is defined in `snowball/snowball.go`.
* The stemmer for each language is defined in a "sub-package", e.g `snowball/spanish`.
* Each language exports a `Stem` function: e.g. `spanish.Stem`,
  which is defined in `snowball/spanish/stem.go`.
* Code that is common to multiple lanuages may go in a separate package,
  e.g. the small `romance` package.

Some notes about the implementation:

* In order to ensure the code is easily extended to non-English lanuages,
  I avoided using bytes and byte arrays, and instead perform all operations
  on runes.  See `snowball/snowballword/snowballword.go` and the
  `SnowballWord` struct.
* In order to avoid casting strings into slices of runes numerous times,
  this implementation uses a single slice of runes stored in the `SnowballWord`
  struct for each word that needs to be stemmed.
* In spite of the foregoing, readability requires that some strings be
  kept around and repeatedly cast into slices of runes.  For example,
  in the Spanish stemmer, one step requires removing suffixes with accute
  accents such as "ución", "logía", and "logías".  If I were to hard-code those
  suffices as slices of runes, the code would be substantially less readable.
* Instead of carrying around the word regions R1, R2, & RV as separate strings
  (or slices or runes, or whatever), we carry around the index where each of
  these regions begins.  These are stored as `R1start`, `R2start`, & `RVstart`
  on the `SnowballWord` struct. I believe this is a relatively efficient way of
  storing R1 and R2.
* The code does not use any maps or regular expressions 1) for kicks, and 2) because
  I thought they'd negatively impact the performance. (But, mostly for #1; I realize
  #2 is silly.)
* I end up refactoring the `snowballword` package a bit every time I implement a
  new language.
* Clearly, the Go implentation of these stemmers is verbose relative to the
  Snowball language.  However, it is much better than the
  [Java version](https://github.com/weavejester/snowball-stemmer/blob/master/src/java/org/tartarus/snowball/ext/frenchStemmer.java)
  and [others](https://github.com/patch/lingua-stem-unine-pm5/blob/master/src/frenchStemmerPlus.txt).

## Testing

To run the tests, do `go test ./...` in the top-level directory.

## Future work

I'd like to implement the Snowball stemmer in more lanuages.
If you can help, I would greatly appreciate it: please fork the project and send
a pull request!

(Also, if you are interested in creating a larger NLP project for Go, please get in touch.)

## Related work

I know of a few other stemmers availble in Go:

* [stemmer](https://github.com/dchest/stemmer) by [Dmitry Chestnykh](https://github.com/dchest).
  His project also
  implements the Snowball (Porter2) English stemmer as well as the Snowball German stemmer.
* [porter-stemmer](https://github.com/a2800276/porter-stemmer.go) - an implementation of the
  original Porter stemming algorithm.
* [go-stem](https://github.com/agonopol/go-stem) by [Alex Gonopolskiy](https://github.com/agonopol).
  Also the original Porter algorithm.
* [paicehusk](https://github.com/Rookii/paicehusk) by [Aaron Groves](https://github.com/rookii).
  This package implements the
  [Paice/Husk](http://www.comp.lancs.ac.uk/computing/research/stemming/)
  stemmer.
* [golibstemmer](https://github.com/rjohnsondev/golibstemmer)
  by [Richard Johnson](https://github.com/rjohnsondev).  This provides Go bindings for the
  [libstemmer](http://snowball.tartarus.org/download.php) C library.
* [snowball](https://bitbucket.org/tebeka/snowball) by [Miki Tebeka](http://web.mikitebeka.com/).
  Also, I believe, Go bindings for the C library.

## Contributors

* Kyle Jensen (kljensen@gmail.com, [@DataKyle](http://twitter.com/datakyle))
* [Shawn Smith](https://github.com/shawnps)
* [Herman Schaaf](https://github.com/hermanschaaf)
* [Anton Södergren](https://github.com/AAAton)
* [Eivind Moland](https://github.com/eivindam)
* Your name should be here!


## License (MIT)

Copyright (c) 2013-2018 the Contributors (see above)

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
