// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Experimental.
type FnGenerated interface {
}

// The jsii proxy struct for FnGenerated
type jsiiProxy_FnGenerated struct {
	_ byte // padding
}

// Experimental.
func NewFnGenerated() FnGenerated {
	_init_.Initialize()

	j := jsiiProxy_FnGenerated{}

	_jsii_.Create(
		"cdktf.FnGenerated",
		nil, // no parameters
		&j,
	)

	return &j
}

// Experimental.
func NewFnGenerated_Override(f FnGenerated) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.FnGenerated",
		nil, // no parameters
		f,
	)
}

// {@link https://developer.hashicorp.com/terraform/language/functions/abs abs} returns the absolute value of the given number. In other words, if the number is zero or positive then it is returned as-is, but if it is negative then it is multiplied by -1 to make it positive before returning it.
// Experimental.
func FnGenerated_Abs(num *float64) *float64 {
	_init_.Initialize()

	if err := validateFnGenerated_AbsParameters(num); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"abs",
		[]interface{}{num},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/abspath abspath} takes a string containing a filesystem path and converts it to an absolute path. That is, if the path is not absolute, it will be joined with the current working directory.
// Experimental.
func FnGenerated_Abspath(path *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_AbspathParameters(path); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"abspath",
		[]interface{}{path},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/alltrue alltrue} returns `true` if all elements in a given collection are `true` or `"true"`. It also returns `true` if the collection is empty.
// Experimental.
func FnGenerated_Alltrue(list *[]interface{}) IResolvable {
	_init_.Initialize()

	if err := validateFnGenerated_AlltrueParameters(list); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"alltrue",
		[]interface{}{list},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/anytrue anytrue} returns `true` if any element in a given collection is `true` or `"true"`. It also returns `false` if the collection is empty.
// Experimental.
func FnGenerated_Anytrue(list *[]interface{}) IResolvable {
	_init_.Initialize()

	if err := validateFnGenerated_AnytrueParameters(list); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"anytrue",
		[]interface{}{list},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/base64decode base64decode} takes a string containing a Base64 character sequence and returns the original string.
// Experimental.
func FnGenerated_Base64decode(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Base64decodeParameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"base64decode",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/base64encode base64encode} applies Base64 encoding to a string.
// Experimental.
func FnGenerated_Base64encode(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Base64encodeParameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"base64encode",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/base64gzip base64gzip} compresses a string with gzip and then encodes the result in Base64 encoding.
// Experimental.
func FnGenerated_Base64gzip(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Base64gzipParameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"base64gzip",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/base64sha256 base64sha256} computes the SHA256 hash of a given string and encodes it with Base64. This is not equivalent to `base64encode(sha256("test"))` since `sha256()` returns hexadecimal representation.
// Experimental.
func FnGenerated_Base64sha256(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Base64sha256Parameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"base64sha256",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/base64sha512 base64sha512} computes the SHA512 hash of a given string and encodes it with Base64. This is not equivalent to `base64encode(sha512("test"))` since `sha512()` returns hexadecimal representation.
// Experimental.
func FnGenerated_Base64sha512(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Base64sha512Parameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"base64sha512",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/basename basename} takes a string containing a filesystem path and removes all except the last portion from it.
// Experimental.
func FnGenerated_Basename(path *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_BasenameParameters(path); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"basename",
		[]interface{}{path},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/can can} evaluates the given expression and returns a boolean value indicating whether the expression produced a result without any errors.
// Experimental.
func FnGenerated_Can(expression interface{}) IResolvable {
	_init_.Initialize()

	if err := validateFnGenerated_CanParameters(expression); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"can",
		[]interface{}{expression},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/ceil ceil} returns the closest whole number that is greater than or equal to the given value, which may be a fraction.
// Experimental.
func FnGenerated_Ceil(num *float64) *float64 {
	_init_.Initialize()

	if err := validateFnGenerated_CeilParameters(num); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"ceil",
		[]interface{}{num},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/chomp chomp} removes newline characters at the end of a string.
// Experimental.
func FnGenerated_Chomp(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_ChompParameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"chomp",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/chunklist chunklist} splits a single list into fixed-size chunks, returning a list of lists.
// Experimental.
func FnGenerated_Chunklist(list *[]interface{}, size *float64) *[]*string {
	_init_.Initialize()

	if err := validateFnGenerated_ChunklistParameters(list, size); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"chunklist",
		[]interface{}{list, size},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/cidrhost cidrhost} calculates a full host IP address for a given host number within a given IP network address prefix.
// Experimental.
func FnGenerated_Cidrhost(prefix *string, hostnum *float64) *string {
	_init_.Initialize()

	if err := validateFnGenerated_CidrhostParameters(prefix, hostnum); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"cidrhost",
		[]interface{}{prefix, hostnum},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/cidrnetmask cidrnetmask} converts an IPv4 address prefix given in CIDR notation into a subnet mask address.
// Experimental.
func FnGenerated_Cidrnetmask(prefix *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_CidrnetmaskParameters(prefix); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"cidrnetmask",
		[]interface{}{prefix},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/cidrsubnet cidrsubnet} calculates a subnet address within given IP network address prefix.
// Experimental.
func FnGenerated_Cidrsubnet(prefix *string, newbits *float64, netnum *float64) *string {
	_init_.Initialize()

	if err := validateFnGenerated_CidrsubnetParameters(prefix, newbits, netnum); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"cidrsubnet",
		[]interface{}{prefix, newbits, netnum},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/cidrsubnets cidrsubnets} calculates a sequence of consecutive IP address ranges within a particular CIDR prefix.
// Experimental.
func FnGenerated_Cidrsubnets(prefix *string, newbits *[]*float64) *[]*string {
	_init_.Initialize()

	if err := validateFnGenerated_CidrsubnetsParameters(prefix, newbits); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"cidrsubnets",
		[]interface{}{prefix, newbits},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/coalesce coalesce} takes any number of arguments and returns the first one that isn't null or an empty string.
// Experimental.
func FnGenerated_Coalesce(vals *[]interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_CoalesceParameters(vals); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"coalesce",
		[]interface{}{vals},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/coalescelist coalescelist} takes any number of list arguments and returns the first one that isn't empty.
// Experimental.
func FnGenerated_Coalescelist(vals *[]interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_CoalescelistParameters(vals); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"coalescelist",
		[]interface{}{vals},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/compact compact} takes a list of strings and returns a new list with any empty string elements removed.
// Experimental.
func FnGenerated_Compact(list *[]*string) *[]*string {
	_init_.Initialize()

	if err := validateFnGenerated_CompactParameters(list); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"compact",
		[]interface{}{list},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/concat concat} takes two or more lists and combines them into a single list.
// Experimental.
func FnGenerated_Concat(seqs *[]interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_ConcatParameters(seqs); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"concat",
		[]interface{}{seqs},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/contains contains} determines whether a given list or set contains a given single value as one of its elements.
// Experimental.
func FnGenerated_Contains(list interface{}, value interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_ContainsParameters(list, value); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"contains",
		[]interface{}{list, value},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/csvdecode csvdecode} decodes a string containing CSV-formatted data and produces a list of maps representing that data.
// Experimental.
func FnGenerated_Csvdecode(str *string) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_CsvdecodeParameters(str); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"csvdecode",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/dirname dirname} takes a string containing a filesystem path and removes the last portion from it.
// Experimental.
func FnGenerated_Dirname(path *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_DirnameParameters(path); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"dirname",
		[]interface{}{path},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/distinct distinct} takes a list and returns a new list with any duplicate elements removed.
// Experimental.
func FnGenerated_Distinct(list *[]interface{}) *[]*string {
	_init_.Initialize()

	if err := validateFnGenerated_DistinctParameters(list); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"distinct",
		[]interface{}{list},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/element element} retrieves a single element from a list.
// Experimental.
func FnGenerated_Element(list interface{}, index *float64) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_ElementParameters(list, index); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"element",
		[]interface{}{list, index},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/endswith endswith} takes two values: a string to check and a suffix string. The function returns true if the first string ends with that exact suffix.
// Experimental.
func FnGenerated_Endswith(str *string, suffix *string) IResolvable {
	_init_.Initialize()

	if err := validateFnGenerated_EndswithParameters(str, suffix); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"endswith",
		[]interface{}{str, suffix},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/file file} reads the contents of a file at the given path and returns them as a string.
// Experimental.
func FnGenerated_File(path *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_FileParameters(path); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"file",
		[]interface{}{path},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/filebase64 filebase64} reads the contents of a file at the given path and returns them as a base64-encoded string.
// Experimental.
func FnGenerated_Filebase64(path *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Filebase64Parameters(path); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"filebase64",
		[]interface{}{path},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/filebase64sha256 filebase64sha256} is a variant of `base64sha256` that hashes the contents of a given file rather than a literal string.
// Experimental.
func FnGenerated_Filebase64sha256(path *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Filebase64sha256Parameters(path); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"filebase64sha256",
		[]interface{}{path},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/filebase64sha512 filebase64sha512} is a variant of `base64sha512` that hashes the contents of a given file rather than a literal string.
// Experimental.
func FnGenerated_Filebase64sha512(path *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Filebase64sha512Parameters(path); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"filebase64sha512",
		[]interface{}{path},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/fileexists fileexists} determines whether a file exists at a given path.
// Experimental.
func FnGenerated_Fileexists(path *string) IResolvable {
	_init_.Initialize()

	if err := validateFnGenerated_FileexistsParameters(path); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"fileexists",
		[]interface{}{path},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/filemd5 filemd5} is a variant of `md5` that hashes the contents of a given file rather than a literal string.
// Experimental.
func FnGenerated_Filemd5(path *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Filemd5Parameters(path); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"filemd5",
		[]interface{}{path},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/fileset fileset} enumerates a set of regular file names given a path and pattern. The path is automatically removed from the resulting set of file names and any result still containing path separators always returns forward slash (`/`) as the path separator for cross-system compatibility.
// Experimental.
func FnGenerated_Fileset(path *string, pattern *string) *[]*string {
	_init_.Initialize()

	if err := validateFnGenerated_FilesetParameters(path, pattern); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"fileset",
		[]interface{}{path, pattern},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/filesha1 filesha1} is a variant of `sha1` that hashes the contents of a given file rather than a literal string.
// Experimental.
func FnGenerated_Filesha1(path *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Filesha1Parameters(path); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"filesha1",
		[]interface{}{path},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/filesha256 filesha256} is a variant of `sha256` that hashes the contents of a given file rather than a literal string.
// Experimental.
func FnGenerated_Filesha256(path *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Filesha256Parameters(path); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"filesha256",
		[]interface{}{path},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/filesha512 filesha512} is a variant of `sha512` that hashes the contents of a given file rather than a literal string.
// Experimental.
func FnGenerated_Filesha512(path *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Filesha512Parameters(path); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"filesha512",
		[]interface{}{path},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/flatten flatten} takes a list and replaces any elements that are lists with a flattened sequence of the list contents.
// Experimental.
func FnGenerated_Flatten(list interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_FlattenParameters(list); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"flatten",
		[]interface{}{list},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/floor floor} returns the closest whole number that is less than or equal to the given value, which may be a fraction.
// Experimental.
func FnGenerated_Floor(num *float64) *float64 {
	_init_.Initialize()

	if err := validateFnGenerated_FloorParameters(num); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"floor",
		[]interface{}{num},
		&returns,
	)

	return returns
}

// The {@link https://developer.hashicorp.com/terraform/language/functions/format format} function produces a string by formatting a number of other values according to a specification string. It is similar to the `printf` function in C, and other similar functions in other programming languages.
// Experimental.
func FnGenerated_Format(format *string, args *[]interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_FormatParameters(format, args); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"format",
		[]interface{}{format, args},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/formatdate formatdate} converts a timestamp into a different time format.
// Experimental.
func FnGenerated_Formatdate(format *string, time *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_FormatdateParameters(format, time); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"formatdate",
		[]interface{}{format, time},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/formatlist formatlist} produces a list of strings by formatting a number of other values according to a specification string.
// Experimental.
func FnGenerated_Formatlist(format *string, args *[]interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_FormatlistParameters(format, args); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"formatlist",
		[]interface{}{format, args},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/indent indent} adds a given number of spaces to the beginnings of all but the first line in a given multi-line string.
// Experimental.
func FnGenerated_Indent(spaces *float64, str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_IndentParameters(spaces, str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"indent",
		[]interface{}{spaces, str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/index index} finds the element index for a given value in a list.
// Experimental.
func FnGenerated_Index(list interface{}, value interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_IndexParameters(list, value); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"index",
		[]interface{}{list, value},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/jsondecode jsondecode} interprets a given string as JSON, returning a representation of the result of decoding that string.
// Experimental.
func FnGenerated_Jsondecode(str *string) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_JsondecodeParameters(str); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"jsondecode",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/jsonencode jsonencode} encodes a given value to a string using JSON syntax.
// Experimental.
func FnGenerated_Jsonencode(val interface{}) *string {
	_init_.Initialize()

	if err := validateFnGenerated_JsonencodeParameters(val); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"jsonencode",
		[]interface{}{val},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/keys keys} takes a map and returns a list containing the keys from that map.
// Experimental.
func FnGenerated_Keys(inputMap interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_KeysParameters(inputMap); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"keys",
		[]interface{}{inputMap},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/length length} determines the length of a given list, map, or string.
// Experimental.
func FnGenerated_LengthOf(value interface{}) *float64 {
	_init_.Initialize()

	if err := validateFnGenerated_LengthOfParameters(value); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"lengthOf",
		[]interface{}{value},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/log log} returns the logarithm of a given number in a given base.
// Experimental.
func FnGenerated_Log(num *float64, base *float64) *float64 {
	_init_.Initialize()

	if err := validateFnGenerated_LogParameters(num, base); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"log",
		[]interface{}{num, base},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/lower lower} converts all cased letters in the given string to lowercase.
// Experimental.
func FnGenerated_Lower(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_LowerParameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"lower",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/matchkeys matchkeys} constructs a new list by taking a subset of elements from one list whose indexes match the corresponding indexes of values in another list.
// Experimental.
func FnGenerated_Matchkeys(values *[]interface{}, keys *[]interface{}, searchset *[]interface{}) *[]*string {
	_init_.Initialize()

	if err := validateFnGenerated_MatchkeysParameters(values, keys, searchset); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"matchkeys",
		[]interface{}{values, keys, searchset},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/max max} takes one or more numbers and returns the greatest number from the set.
// Experimental.
func FnGenerated_Max(numbers *[]*float64) *float64 {
	_init_.Initialize()

	if err := validateFnGenerated_MaxParameters(numbers); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"max",
		[]interface{}{numbers},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/md5 md5} computes the MD5 hash of a given string and encodes it with hexadecimal digits.
// Experimental.
func FnGenerated_Md5(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Md5Parameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"md5",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/merge merge} takes an arbitrary number of maps or objects, and returns a single map or object that contains a merged set of elements from all arguments.
// Experimental.
func FnGenerated_Merge(maps *[]interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_MergeParameters(maps); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"merge",
		[]interface{}{maps},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/min min} takes one or more numbers and returns the smallest number from the set.
// Experimental.
func FnGenerated_Min(numbers *[]*float64) *float64 {
	_init_.Initialize()

	if err := validateFnGenerated_MinParameters(numbers); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"min",
		[]interface{}{numbers},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/nonsensitive nonsensitive} takes a sensitive value and returns a copy of that value with the sensitive marking removed, thereby exposing the sensitive value.
// Experimental.
func FnGenerated_Nonsensitive(value interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_NonsensitiveParameters(value); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"nonsensitive",
		[]interface{}{value},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/one one} takes a list, set, or tuple value with either zero or one elements. If the collection is empty, `one` returns `null`. Otherwise, `one` returns the first element. If there are two or more elements then `one` will return an error.
// Experimental.
func FnGenerated_One(list interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_OneParameters(list); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"one",
		[]interface{}{list},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/parseint parseint} parses the given string as a representation of an integer in the specified base and returns the resulting number. The base must be between 2 and 62 inclusive.
// Experimental.
func FnGenerated_Parseint(number interface{}, base *float64) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_ParseintParameters(number, base); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"parseint",
		[]interface{}{number, base},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/pathexpand pathexpand} takes a filesystem path that might begin with a `~` segment, and if so it replaces that segment with the current user's home directory path.
// Experimental.
func FnGenerated_Pathexpand(path *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_PathexpandParameters(path); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"pathexpand",
		[]interface{}{path},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/plantimestamp plantimestamp} returns a UTC timestamp string in [RFC 3339](https://tools.ietf.org/html/rfc3339) format, fixed to a constant time representing the time of the plan.
// Experimental.
func FnGenerated_Plantimestamp() *string {
	_init_.Initialize()

	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"plantimestamp",
		nil, // no parameters
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/pow pow} calculates an exponent, by raising its first argument to the power of the second argument.
// Experimental.
func FnGenerated_Pow(num *float64, power *float64) *float64 {
	_init_.Initialize()

	if err := validateFnGenerated_PowParameters(num, power); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"pow",
		[]interface{}{num, power},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/regex regex} applies a [regular expression](https://en.wikipedia.org/wiki/Regular_expression) to a string and returns the matching substrings.
// Experimental.
func FnGenerated_Regex(pattern *string, str *string) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_RegexParameters(pattern, str); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"regex",
		[]interface{}{pattern, str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/regexall regexall} applies a [regular expression](https://en.wikipedia.org/wiki/Regular_expression) to a string and returns a list of all matches.
// Experimental.
func FnGenerated_Regexall(pattern *string, str *string) *[]*string {
	_init_.Initialize()

	if err := validateFnGenerated_RegexallParameters(pattern, str); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"regexall",
		[]interface{}{pattern, str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/replace replace} searches a given string for another given substring, and replaces each occurrence with a given replacement string.
// Experimental.
func FnGenerated_Replace(str *string, substr *string, replace *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_ReplaceParameters(str, substr, replace); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"replace",
		[]interface{}{str, substr, replace},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/reverse reverse} takes a sequence and produces a new sequence of the same length with all of the same elements as the given sequence but in reverse order.
// Experimental.
func FnGenerated_Reverse(list interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_ReverseParameters(list); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"reverse",
		[]interface{}{list},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/rsadecrypt rsadecrypt} decrypts an RSA-encrypted ciphertext, returning the corresponding cleartext.
// Experimental.
func FnGenerated_Rsadecrypt(ciphertext *string, privatekey *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_RsadecryptParameters(ciphertext, privatekey); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"rsadecrypt",
		[]interface{}{ciphertext, privatekey},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/sensitive sensitive} takes any value and returns a copy of it marked so that Terraform will treat it as sensitive, with the same meaning and behavior as for [sensitive input variables](/terraform/language/values/variables#suppressing-values-in-cli-output).
// Experimental.
func FnGenerated_Sensitive(value interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_SensitiveParameters(value); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"sensitive",
		[]interface{}{value},
		&returns,
	)

	return returns
}

// The {@link https://developer.hashicorp.com/terraform/language/functions/setintersection setintersection} function takes multiple sets and produces a single set containing only the elements that all of the given sets have in common. In other words, it computes the [intersection](https://en.wikipedia.org/wiki/Intersection_\(set_theory\)) of the sets.
// Experimental.
func FnGenerated_Setintersection(first_set *[]interface{}, other_sets *[]*[]interface{}) *[]*string {
	_init_.Initialize()

	if err := validateFnGenerated_SetintersectionParameters(first_set, other_sets); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"setintersection",
		[]interface{}{first_set, other_sets},
		&returns,
	)

	return returns
}

// The {@link https://developer.hashicorp.com/terraform/language/functions/setproduct setproduct} function finds all of the possible combinations of elements from all of the given sets by computing the [Cartesian product](https://en.wikipedia.org/wiki/Cartesian_product).
// Experimental.
func FnGenerated_Setproduct(sets *[]interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_SetproductParameters(sets); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"setproduct",
		[]interface{}{sets},
		&returns,
	)

	return returns
}

// The {@link https://developer.hashicorp.com/terraform/language/functions/setsubtract setsubtract} function returns a new set containing the elements from the first set that are not present in the second set. In other words, it computes the [relative complement](https://en.wikipedia.org/wiki/Complement_\(set_theory\)#Relative_complement) of the second set.
// Experimental.
func FnGenerated_Setsubtract(a *[]interface{}, b *[]interface{}) *[]*string {
	_init_.Initialize()

	if err := validateFnGenerated_SetsubtractParameters(a, b); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"setsubtract",
		[]interface{}{a, b},
		&returns,
	)

	return returns
}

// The {@link https://developer.hashicorp.com/terraform/language/functions/setunion setunion} function takes multiple sets and produces a single set containing the elements from all of the given sets. In other words, it computes the [union](https://en.wikipedia.org/wiki/Union_\(set_theory\)) of the sets.
// Experimental.
func FnGenerated_Setunion(first_set *[]interface{}, other_sets *[]*[]interface{}) *[]*string {
	_init_.Initialize()

	if err := validateFnGenerated_SetunionParameters(first_set, other_sets); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"setunion",
		[]interface{}{first_set, other_sets},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/sha1 sha1} computes the SHA1 hash of a given string and encodes it with hexadecimal digits.
// Experimental.
func FnGenerated_Sha1(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Sha1Parameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"sha1",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/sha256 sha256} computes the SHA256 hash of a given string and encodes it with hexadecimal digits.
// Experimental.
func FnGenerated_Sha256(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Sha256Parameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"sha256",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/sha512 sha512} computes the SHA512 hash of a given string and encodes it with hexadecimal digits.
// Experimental.
func FnGenerated_Sha512(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Sha512Parameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"sha512",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/signum signum} determines the sign of a number, returning a number between -1 and 1 to represent the sign.
// Experimental.
func FnGenerated_Signum(num *float64) *float64 {
	_init_.Initialize()

	if err := validateFnGenerated_SignumParameters(num); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"signum",
		[]interface{}{num},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/slice slice} extracts some consecutive elements from within a list.
// Experimental.
func FnGenerated_Slice(list interface{}, start_index *float64, end_index *float64) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_SliceParameters(list, start_index, end_index); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"slice",
		[]interface{}{list, start_index, end_index},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/sort sort} takes a list of strings and returns a new list with those strings sorted lexicographically.
// Experimental.
func FnGenerated_Sort(list *[]*string) *[]*string {
	_init_.Initialize()

	if err := validateFnGenerated_SortParameters(list); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"sort",
		[]interface{}{list},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/split split} produces a list by dividing a given string at all occurrences of a given separator.
// Experimental.
func FnGenerated_Split(separator *string, str *string) *[]*string {
	_init_.Initialize()

	if err := validateFnGenerated_SplitParameters(separator, str); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"split",
		[]interface{}{separator, str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/startswith startswith} takes two values: a string to check and a prefix string. The function returns true if the string begins with that exact prefix.
// Experimental.
func FnGenerated_Startswith(str *string, prefix *string) IResolvable {
	_init_.Initialize()

	if err := validateFnGenerated_StartswithParameters(str, prefix); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"startswith",
		[]interface{}{str, prefix},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/strcontains strcontains} takes two values: a string to check and an expected substring. The function returns true if the string has the substring contained within it.
// Experimental.
func FnGenerated_Strcontains(str *string, substr *string) IResolvable {
	_init_.Initialize()

	if err := validateFnGenerated_StrcontainsParameters(str, substr); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"strcontains",
		[]interface{}{str, substr},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/strrev strrev} reverses the characters in a string. Note that the characters are treated as _Unicode characters_ (in technical terms, Unicode [grapheme cluster boundaries](https://unicode.org/reports/tr29/#Grapheme_Cluster_Boundaries) are respected).
// Experimental.
func FnGenerated_Strrev(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_StrrevParameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"strrev",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/substr substr} extracts a substring from a given string by offset and (maximum) length.
// Experimental.
func FnGenerated_Substr(str *string, offset *float64, length *float64) *string {
	_init_.Initialize()

	if err := validateFnGenerated_SubstrParameters(str, offset, length); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"substr",
		[]interface{}{str, offset, length},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/sum sum} takes a list or set of numbers and returns the sum of those numbers.
// Experimental.
func FnGenerated_Sum(list interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_SumParameters(list); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"sum",
		[]interface{}{list},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/templatefile templatefile} reads the file at the given path and renders its content as a template using a supplied set of template variables.
// Experimental.
func FnGenerated_Templatefile(path *string, vars interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_TemplatefileParameters(path, vars); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"templatefile",
		[]interface{}{path, vars},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/textdecodebase64 textdecodebase64} function decodes a string that was previously Base64-encoded, and then interprets the result as characters in a specified character encoding.
// Experimental.
func FnGenerated_Textdecodebase64(source *string, encoding *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Textdecodebase64Parameters(source, encoding); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"textdecodebase64",
		[]interface{}{source, encoding},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/textencodebase64 textencodebase64} encodes the unicode characters in a given string using a specified character encoding, returning the result base64 encoded because Terraform language strings are always sequences of unicode characters.
// Experimental.
func FnGenerated_Textencodebase64(str *string, encoding *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Textencodebase64Parameters(str, encoding); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"textencodebase64",
		[]interface{}{str, encoding},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/timeadd timeadd} adds a duration to a timestamp, returning a new timestamp.
// Experimental.
func FnGenerated_Timeadd(timestamp *string, duration *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_TimeaddParameters(timestamp, duration); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"timeadd",
		[]interface{}{timestamp, duration},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/timecmp timecmp} compares two timestamps and returns a number that represents the ordering of the instants those timestamps represent.
// Experimental.
func FnGenerated_Timecmp(timestamp_a *string, timestamp_b *string) *float64 {
	_init_.Initialize()

	if err := validateFnGenerated_TimecmpParameters(timestamp_a, timestamp_b); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"timecmp",
		[]interface{}{timestamp_a, timestamp_b},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/timestamp timestamp} returns a UTC timestamp string in [RFC 3339](https://tools.ietf.org/html/rfc3339) format.
// Experimental.
func FnGenerated_Timestamp() *string {
	_init_.Initialize()

	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"timestamp",
		nil, // no parameters
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/title title} converts the first letter of each word in the given string to uppercase.
// Experimental.
func FnGenerated_Title(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_TitleParameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"title",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/tobool tobool} converts its argument to a boolean value.
// Experimental.
func FnGenerated_Tobool(v interface{}) IResolvable {
	_init_.Initialize()

	if err := validateFnGenerated_ToboolParameters(v); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"tobool",
		[]interface{}{v},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/tolist tolist} converts its argument to a list value.
// Experimental.
func FnGenerated_Tolist(v interface{}) *[]*string {
	_init_.Initialize()

	if err := validateFnGenerated_TolistParameters(v); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"tolist",
		[]interface{}{v},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/tomap tomap} converts its argument to a map value.
// Experimental.
func FnGenerated_Tomap(v interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_TomapParameters(v); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"tomap",
		[]interface{}{v},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/tonumber tonumber} converts its argument to a number value.
// Experimental.
func FnGenerated_Tonumber(v interface{}) *float64 {
	_init_.Initialize()

	if err := validateFnGenerated_TonumberParameters(v); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"tonumber",
		[]interface{}{v},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/toset toset} converts its argument to a set value.
// Experimental.
func FnGenerated_Toset(v interface{}) *[]*string {
	_init_.Initialize()

	if err := validateFnGenerated_TosetParameters(v); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"toset",
		[]interface{}{v},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/tostring tostring} converts its argument to a string value.
// Experimental.
func FnGenerated_Tostring(v interface{}) *string {
	_init_.Initialize()

	if err := validateFnGenerated_TostringParameters(v); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"tostring",
		[]interface{}{v},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/transpose transpose} takes a map of lists of strings and swaps the keys and values to produce a new map of lists of strings.
// Experimental.
func FnGenerated_Transpose(values interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_TransposeParameters(values); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"transpose",
		[]interface{}{values},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/trim trim} removes the specified set of characters from the start and end of the given string.
// Experimental.
func FnGenerated_Trim(str *string, cutset *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_TrimParameters(str, cutset); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"trim",
		[]interface{}{str, cutset},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/trimprefix trimprefix} removes the specified prefix from the start of the given string. If the string does not start with the prefix, the string is returned unchanged.
// Experimental.
func FnGenerated_Trimprefix(str *string, prefix *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_TrimprefixParameters(str, prefix); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"trimprefix",
		[]interface{}{str, prefix},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/trimspace trimspace} removes any space characters from the start and end of the given string.
// Experimental.
func FnGenerated_Trimspace(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_TrimspaceParameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"trimspace",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/trimsuffix trimsuffix} removes the specified suffix from the end of the given string.
// Experimental.
func FnGenerated_Trimsuffix(str *string, suffix *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_TrimsuffixParameters(str, suffix); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"trimsuffix",
		[]interface{}{str, suffix},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/try try} evaluates all of its argument expressions in turn and returns the result of the first one that does not produce any errors.
// Experimental.
func FnGenerated_Try(expressions *[]interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_TryParameters(expressions); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"try",
		[]interface{}{expressions},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/upper upper} converts all cased letters in the given string to uppercase.
// Experimental.
func FnGenerated_Upper(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_UpperParameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"upper",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/urlencode urlencode} applies URL encoding to a given string.
// Experimental.
func FnGenerated_Urlencode(str *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_UrlencodeParameters(str); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"urlencode",
		[]interface{}{str},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/uuid uuid} generates a unique identifier string.
// Experimental.
func FnGenerated_Uuid() *string {
	_init_.Initialize()

	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"uuid",
		nil, // no parameters
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/uuidv5 uuidv5} generates a _name-based_ UUID, as described in [RFC 4122 section 4.3](https://tools.ietf.org/html/rfc4122#section-4.3), also known as a "version 5" UUID.
// Experimental.
func FnGenerated_Uuidv5(namespace *string, name *string) *string {
	_init_.Initialize()

	if err := validateFnGenerated_Uuidv5Parameters(namespace, name); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"uuidv5",
		[]interface{}{namespace, name},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/values values} takes a map and returns a list containing the values of the elements in that map.
// Experimental.
func FnGenerated_Values(mapping interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_ValuesParameters(mapping); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"values",
		[]interface{}{mapping},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/yamldecode yamldecode} parses a string as a subset of YAML, and produces a representation of its value.
// Experimental.
func FnGenerated_Yamldecode(src *string) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_YamldecodeParameters(src); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"yamldecode",
		[]interface{}{src},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/yamlencode yamlencode} encodes a given value to a string using [YAML 1.2](https://yaml.org/spec/1.2/spec.html) block syntax.
// Experimental.
func FnGenerated_Yamlencode(value interface{}) *string {
	_init_.Initialize()

	if err := validateFnGenerated_YamlencodeParameters(value); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"yamlencode",
		[]interface{}{value},
		&returns,
	)

	return returns
}

// {@link https://developer.hashicorp.com/terraform/language/functions/zipmap zipmap} constructs a map from a list of keys and a corresponding list of values.
// Experimental.
func FnGenerated_Zipmap(keys *[]*string, values interface{}) interface{} {
	_init_.Initialize()

	if err := validateFnGenerated_ZipmapParameters(keys, values); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.StaticInvoke(
		"cdktf.FnGenerated",
		"zipmap",
		[]interface{}{keys, values},
		&returns,
	)

	return returns
}

