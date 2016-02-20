# Nitro

Quick and easy performance analyzer library for golang.

## Overview

Nitro is a quick and easy performance analyzer library for golang.
It is useful for comparing A/B against different drafts of functions
or different functions.

## Implementing Nitro

Using Nitro is simple. First use go get to install the latest version
of the library.

    $ go get github.com/spf13/nitro

Next include nitro in your application.

    import "github.com/spf13/nitro"

Somewhere near the beginning of your application (or where you want to
begin profiling) call

    timer := nitro.Initialize()

Then throughout your application wherever a major division of work is
call

    timer.Step("name of step")

### Flags

Nitro automatically adds a flag to your application. If you aren't
already using flags in your application the following code is an example
of how you may use flags. *Make sure to import "flag"*.

    func main() {
        flag.Parse()
    }

## Bring your own flags implementation

If you are using your own flag system or a commander like
[cobra](http://cobra.spf13.com) you may want to enable Nitro on your
own. To enable Nitro without using the default flags, simply set the
package variable '&nitro.AnalysisOn' to true. The following example uses
a flagset:

	Flags().BoolVar(&nitro.AnalysisOn, "stepAnalysis", false, "display memory and timing of different steps of the program")

    var Timer *nitro.B

    func init() {
        Timer = nitro.Initalize()
    }

    func TrackMe() {
        // a bunch of code here
        Timer.Step("important function to track")
    }

    func TrackAnother() {
        // more code here
        Timer.Step("another function to track")
    }


## Usage

Once the library is implemented throughout your application simply run
your application and pass the "--stepAnalysis" flag to it. It does not
need to be built to run, but can be called from go run or the binary
form.

    $ go run ./my_application --stepAnalysis

### Example output
The following output comes from the [hugo](http://github.com/spf13/hugo) static site generator library.  Nitro was built as a component of hugo and was extracted into it's own library.

    $ ./main -p spf13 -b http://localhost -d --stepAnalysis

    initialize & template prep:
        4.664481ms (5.887625ms)	        0.43 MB 	4583 Allocs
    import pages:
        65.196788ms (71.107809ms)	   17.13 MB 	70151 Allocs
    build indexes:
        1.823434ms (72.960713ms)	    0.12 MB 	3720 Allocs
    render and write indexes:
        212.06721ms (285.057592ms)	   65.72 MB 	362557 Allocs
    render and write lists:
        17.796945ms (302.87847ms)	    7.76 MB 	33122 Allocs
    render pages:
        50.092756ms (352.998539ms)	   11.27 MB 	139898 Allocs
    render shortcodes:
        11.34692ms (364.386939ms)	    6.24 MB 	21260 Allocs
    render and write homepage:
        4.075194ms (368.497883ms)	    0.84 MB 	3906 Allocs
    write pages:
        8.73933ms (377.263888ms)	    0.11 MB 	1672 Allocs

## Release Notes
* **0.5.0** Oct 1, 2013
  * Now supporting non flag based enabling
* **0.4.0** June 19, 2013
  * Implement first draft

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request

## Contributors

Names in no particular order:

* [spf13](https://github.com/spf13)

## License

nitro is released under the Apache 2.0 license. See [LICENSE.txt](https://github.com/spf13/nitro/blob/master/LICENSE.txt)
