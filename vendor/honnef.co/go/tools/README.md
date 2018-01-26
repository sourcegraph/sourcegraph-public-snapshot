# honnef.co/go/tools

`honnef.co/go/tools/...` is a collection of tools and libraries for
working with Go code, including linters and static analysis.

**These tools are supported by
[patrons on Patreon](https://www.patreon.com/dominikh) and
[sponsors](#sponsors). If you use these tools at your company,
consider purchasing
[commercial support](https://staticcheck.io/pricing).**

## Tools

All of the following tools can be found in the cmd/ directory. Each
tool is accompanied by its own README, describing it in more detail.

| Tool                                               | Description                                                      |
|----------------------------------------------------|------------------------------------------------------------------|
| [gosimple](cmd/gosimple/)                          | Detects code that could be rewritten in a simpler way.           |
| [keyify](cmd/keyify/)                              | Transforms an unkeyed struct literal into a keyed one.           |
| [rdeps](cmd/rdeps/)                                | Find all reverse dependencies of a set of packages               |
| [staticcheck](cmd/staticcheck/)                    | Detects a myriad of bugs and inefficiencies in your code.        |
| [structlayout](cmd/structlayout/)                  | Displays the layout (field sizes and padding) of structs.        |
| [structlayout-optimize](cmd/structlayout-optimize) | Reorders struct fields to minimize the amount of padding.        |
| [structlayout-pretty](cmd/structlayout-pretty)     | Formats the output of structlayout with ASCII art.               |
| [unused](cmd/unused/)                              | Reports unused identifiers (types, functions, ...) in your code. |
|                                                    |                                                                  |
| [megacheck](cmd/megacheck)                         | Run staticcheck, gosimple and unused in one go                   |

## Libraries

In addition to the aforementioned tools, this repository contains the
libraries necessary to implement these tools.

Unless otherwise noted, none of these libraries have stable APIs.
Their main purpose is to aid the implementation of the tools. If you
decide to use these libraries, please vendor them and expect regular
backwards-incompatible changes.

## Documentation

You can find more documentation on
[staticcheck.io](https://staticcheck.io).

## Sponsors

This project is sponsored by:

[<img src="images/sponsors/digitalocean.png" alt="DigitalOcean" height="35"></img>](https://digitalocean.com)

## Licenses

All original code in this repository is licensed under the following
MIT license.

> Copyright (c) 2016 Dominik Honnef
>
> Permission is hereby granted, free of charge, to any person obtaining
> a copy of this software and associated documentation files (the
> "Software"), to deal in the Software without restriction, including
> without limitation the rights to use, copy, modify, merge, publish,
> distribute, sublicense, and/or sell copies of the Software, and to
> permit persons to whom the Software is furnished to do so, subject to
> the following conditions:
>
> The above copyright notice and this permission notice shall be
> included in all copies or substantial portions of the Software.
>
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
> EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
> MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
> NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
> LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
> OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
> WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

In addition, some libraries reuse code owned by The Go Authors and
licensed under the following BSD 3-clause license:

> Copyright (c) 2013 The Go Authors. All rights reserved.
>
> Redistribution and use in source and binary forms, with or without
> modification, are permitted provided that the following conditions are
> met:
>
>    * Redistributions of source code must retain the above copyright
> notice, this list of conditions and the following disclaimer.
>    * Redistributions in binary form must reproduce the above
> copyright notice, this list of conditions and the following disclaimer
> in the documentation and/or other materials provided with the
> distribution.
>    * Neither the name of Google Inc. nor the names of its
> contributors may be used to endorse or promote products derived from
> this software without specific prior written permission.
>
> THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
> "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
> LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
> A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
> OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
> SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
> LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
> DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
> THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
> (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
> OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
