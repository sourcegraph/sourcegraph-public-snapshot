# Afero

A FileSystem Abstraction System for Go

[![Build Status](https://travis-ci.org/spf13/afero.png)](https://travis-ci.org/spf13/afero)

## Overview

Package Afero provides types and methods for interacting with the filesystem,
as an abstraction layer.

It provides a few implementations that are largely interoperable. One that
uses the operating system filesystem, one that uses memory to store files
(cross platform) and an interface that should be implemented if you want to
provide your own filesystem.

It is suitable for use in a any situation where you would consider using
the OS package as it provides an additional abstraction that makes it
easy to use a memory backed file system during testing. It also adds
support for the http filesystem for full interoperability.

Afero has an exceptionally clean interface and simple design without needless
constructors or initialization methods.

## The name

Initially this project was called fs. Unfortunately as I used it, the
name proved confusing, there were too many fs’. In looking for
alternatives I looked up the word 'abstract' in a variety of different
languages. Afero is the Greek word for abstract and it seemed to be a
fitting name for the project. It also means ‘to do’ or ‘thing’ in
Esperanto which is also fitting.

## Interface

Afero simply takes the interfaces already defined throughout the standard
library and unifies them into a pair of interfaces that satisfy all
known uses. One interface for a file and one for a filesystem.

## Filesystems

Afero additionally comes with a few filesystems and file implementations
ready to use.

### OsFs

The first is simply a wrapper around the native OS calls. This makes it
very easy to use as all of the calls are the same as the existing OS
calls. It also makes it trivial to have your code use the OS during
operation and a mock filesystem during testing or as needed.

### MemMapFs

Afero also provides a fully atomic memory backed filesystem perfect for use in
mocking and to speed up unnecessary disk io when persistence isn’t
necessary. It is fully concurrent and will work within go routines
safely.

### MemFile

As part of MemMapFs, Afero also provides an atomic, fully concurrent memory
backed file implementation. This can be used in other memory backed file
systems with ease. Plans are to add a radix tree memory stored file
system using MemFile. 

## Usage


### Installing
Using Afero is easy. First use go get to install the latest version
of the library.

    $ go get github.com/spf13/afero

Next include Afero in your application.

    import "github.com/spf13/afero"

## Using Afero

There are a few different ways to use Afero. You could use the
interfaces alone to define you own file system. You could use it as a
wrapper for the OS packages. You could use it to define different
filesystems for different parts of your application. Here we will
demonstrate a basic usage.

First define a package variable and set it to a pointer to a filesystem.

    var AppFs afero.Fs = &afero.MemMapFs{}

    or

    var AppFs afero.Fs = &afero.OsFs{}

It is important to note that if you repeat the composite literal you
will be using a completely new and isolated filesystem. In the case of
OsFs it will still use the same underlying filesystem but will reduce
the ability to drop in other filesystems as desired.

Then throughout your functions and methods use the methods familiar
already from the OS package.

Methods Available:

       Create(name string) : File, error
       Mkdir(name string, perm os.FileMode) : error
       MkdirAll(path string, perm os.FileMode) : error
       Name() : string
       Open(name string) : File, error
       OpenFile(name string, flag int, perm os.FileMode) : File, error
       Remove(name string) : error
       RemoveAll(path string) : error
       Rename(oldname, newname string) : error
       Stat(name string) : os.FileInfo, error

In our case we would call `AppFs.Open()` as an example because that is how we’ve defined to
access our filesystem.

In some applications it may make sense to define a new package that
simply exports the file system variable for easy access from anywhere.


### Using a Mock Filesystem for Testing

There is a large benefit to using a mock filesystem for testing. It has
a completely blank state every time it is initialized and can be easily
reproducible regardless of OS. It is also faster than disk which makes
the tests run faster. Lastly it doesn’t require any clean up after tests
are run.

One way to accomplish this is to define a variable as mentioned above.
In your application this will be set to &afero.OsFs{} during testing you
can set it to &afero.MemMapFs{}.

### Using with Http

The Http package requires a slightly specific version of Open which
returns an http.File type.

Afero provides an httpFs file system which satisfies this requirement.
Any Afero FileSystem can be used as an httpFs.

	httpFs := &afero.HttpFs{SourceFs: <ExistingFS>}
	fileserver := http.FileServer(httpFs.Dir(<PATH>)))
    http.Handle("/", fileserver)


## Release Notes
* **0.8.0** 2014.10.28
  * First public version
  * Interfaces feel ready for people to build using
  * Interfaces satisfy all known uses
  * MemMapFs passes the majority of the OS test suite
  * OsFs passes the majority of the OS test suite

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request

## Contributors

Names in no particular order:

* [spf13](https://github.com/spf13)
* [jaqx0r](https://github.com/jaqx0r)

## License

Afero is released under the Apache 2.0 license. See [LICENSE.txt](https://github.com/spf13/afero/blob/master/LICENSE.txt)
