// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"
)

func validateFnGenerated_AbsParameters(num *float64) error {
	if num == nil {
		return fmt.Errorf("parameter num is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_AbspathParameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_AlltrueParameters(list *[]interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_AnytrueParameters(list *[]interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Base64decodeParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Base64encodeParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Base64gzipParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Base64sha256Parameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Base64sha512Parameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_BasenameParameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_CanParameters(expression interface{}) error {
	if expression == nil {
		return fmt.Errorf("parameter expression is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_CeilParameters(num *float64) error {
	if num == nil {
		return fmt.Errorf("parameter num is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_ChompParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_ChunklistParameters(list *[]interface{}, size *float64) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	if size == nil {
		return fmt.Errorf("parameter size is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_CidrhostParameters(prefix *string, hostnum *float64) error {
	if prefix == nil {
		return fmt.Errorf("parameter prefix is required, but nil was provided")
	}

	if hostnum == nil {
		return fmt.Errorf("parameter hostnum is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_CidrnetmaskParameters(prefix *string) error {
	if prefix == nil {
		return fmt.Errorf("parameter prefix is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_CidrsubnetParameters(prefix *string, newbits *float64, netnum *float64) error {
	if prefix == nil {
		return fmt.Errorf("parameter prefix is required, but nil was provided")
	}

	if newbits == nil {
		return fmt.Errorf("parameter newbits is required, but nil was provided")
	}

	if netnum == nil {
		return fmt.Errorf("parameter netnum is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_CidrsubnetsParameters(prefix *string, newbits *[]*float64) error {
	if prefix == nil {
		return fmt.Errorf("parameter prefix is required, but nil was provided")
	}

	if newbits == nil {
		return fmt.Errorf("parameter newbits is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_CoalesceParameters(vals *[]interface{}) error {
	if vals == nil {
		return fmt.Errorf("parameter vals is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_CoalescelistParameters(vals *[]interface{}) error {
	if vals == nil {
		return fmt.Errorf("parameter vals is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_CompactParameters(list *[]*string) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_ConcatParameters(seqs *[]interface{}) error {
	if seqs == nil {
		return fmt.Errorf("parameter seqs is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_ContainsParameters(list interface{}, value interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_CsvdecodeParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_DirnameParameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_DistinctParameters(list *[]interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_ElementParameters(list interface{}, index *float64) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	if index == nil {
		return fmt.Errorf("parameter index is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_EndswithParameters(str *string, suffix *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if suffix == nil {
		return fmt.Errorf("parameter suffix is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_FileParameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Filebase64Parameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Filebase64sha256Parameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Filebase64sha512Parameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_FileexistsParameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Filemd5Parameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_FilesetParameters(path *string, pattern *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	if pattern == nil {
		return fmt.Errorf("parameter pattern is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Filesha1Parameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Filesha256Parameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Filesha512Parameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_FlattenParameters(list interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_FloorParameters(num *float64) error {
	if num == nil {
		return fmt.Errorf("parameter num is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_FormatParameters(format *string, args *[]interface{}) error {
	if format == nil {
		return fmt.Errorf("parameter format is required, but nil was provided")
	}

	if args == nil {
		return fmt.Errorf("parameter args is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_FormatdateParameters(format *string, time *string) error {
	if format == nil {
		return fmt.Errorf("parameter format is required, but nil was provided")
	}

	if time == nil {
		return fmt.Errorf("parameter time is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_FormatlistParameters(format *string, args *[]interface{}) error {
	if format == nil {
		return fmt.Errorf("parameter format is required, but nil was provided")
	}

	if args == nil {
		return fmt.Errorf("parameter args is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_IndentParameters(spaces *float64, str *string) error {
	if spaces == nil {
		return fmt.Errorf("parameter spaces is required, but nil was provided")
	}

	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_IndexParameters(list interface{}, value interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_JsondecodeParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_JsonencodeParameters(val interface{}) error {
	if val == nil {
		return fmt.Errorf("parameter val is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_KeysParameters(inputMap interface{}) error {
	if inputMap == nil {
		return fmt.Errorf("parameter inputMap is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_LengthOfParameters(value interface{}) error {
	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_LogParameters(num *float64, base *float64) error {
	if num == nil {
		return fmt.Errorf("parameter num is required, but nil was provided")
	}

	if base == nil {
		return fmt.Errorf("parameter base is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_LowerParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_MatchkeysParameters(values *[]interface{}, keys *[]interface{}, searchset *[]interface{}) error {
	if values == nil {
		return fmt.Errorf("parameter values is required, but nil was provided")
	}

	if keys == nil {
		return fmt.Errorf("parameter keys is required, but nil was provided")
	}

	if searchset == nil {
		return fmt.Errorf("parameter searchset is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_MaxParameters(numbers *[]*float64) error {
	if numbers == nil {
		return fmt.Errorf("parameter numbers is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Md5Parameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_MergeParameters(maps *[]interface{}) error {
	if maps == nil {
		return fmt.Errorf("parameter maps is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_MinParameters(numbers *[]*float64) error {
	if numbers == nil {
		return fmt.Errorf("parameter numbers is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_NonsensitiveParameters(value interface{}) error {
	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_OneParameters(list interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_ParseintParameters(number interface{}, base *float64) error {
	if number == nil {
		return fmt.Errorf("parameter number is required, but nil was provided")
	}

	if base == nil {
		return fmt.Errorf("parameter base is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_PathexpandParameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_PowParameters(num *float64, power *float64) error {
	if num == nil {
		return fmt.Errorf("parameter num is required, but nil was provided")
	}

	if power == nil {
		return fmt.Errorf("parameter power is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_RegexParameters(pattern *string, str *string) error {
	if pattern == nil {
		return fmt.Errorf("parameter pattern is required, but nil was provided")
	}

	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_RegexallParameters(pattern *string, str *string) error {
	if pattern == nil {
		return fmt.Errorf("parameter pattern is required, but nil was provided")
	}

	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_ReplaceParameters(str *string, substr *string, replace *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if substr == nil {
		return fmt.Errorf("parameter substr is required, but nil was provided")
	}

	if replace == nil {
		return fmt.Errorf("parameter replace is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_ReverseParameters(list interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_RsadecryptParameters(ciphertext *string, privatekey *string) error {
	if ciphertext == nil {
		return fmt.Errorf("parameter ciphertext is required, but nil was provided")
	}

	if privatekey == nil {
		return fmt.Errorf("parameter privatekey is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_SensitiveParameters(value interface{}) error {
	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_SetintersectionParameters(first_set *[]interface{}, other_sets *[]*[]interface{}) error {
	if first_set == nil {
		return fmt.Errorf("parameter first_set is required, but nil was provided")
	}

	if other_sets == nil {
		return fmt.Errorf("parameter other_sets is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_SetproductParameters(sets *[]interface{}) error {
	if sets == nil {
		return fmt.Errorf("parameter sets is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_SetsubtractParameters(a *[]interface{}, b *[]interface{}) error {
	if a == nil {
		return fmt.Errorf("parameter a is required, but nil was provided")
	}

	if b == nil {
		return fmt.Errorf("parameter b is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_SetunionParameters(first_set *[]interface{}, other_sets *[]*[]interface{}) error {
	if first_set == nil {
		return fmt.Errorf("parameter first_set is required, but nil was provided")
	}

	if other_sets == nil {
		return fmt.Errorf("parameter other_sets is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Sha1Parameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Sha256Parameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Sha512Parameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_SignumParameters(num *float64) error {
	if num == nil {
		return fmt.Errorf("parameter num is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_SliceParameters(list interface{}, start_index *float64, end_index *float64) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	if start_index == nil {
		return fmt.Errorf("parameter start_index is required, but nil was provided")
	}

	if end_index == nil {
		return fmt.Errorf("parameter end_index is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_SortParameters(list *[]*string) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_SplitParameters(separator *string, str *string) error {
	if separator == nil {
		return fmt.Errorf("parameter separator is required, but nil was provided")
	}

	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_StartswithParameters(str *string, prefix *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if prefix == nil {
		return fmt.Errorf("parameter prefix is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_StrcontainsParameters(str *string, substr *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if substr == nil {
		return fmt.Errorf("parameter substr is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_StrrevParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_SubstrParameters(str *string, offset *float64, length *float64) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if offset == nil {
		return fmt.Errorf("parameter offset is required, but nil was provided")
	}

	if length == nil {
		return fmt.Errorf("parameter length is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_SumParameters(list interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TemplatefileParameters(path *string, vars interface{}) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	if vars == nil {
		return fmt.Errorf("parameter vars is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Textdecodebase64Parameters(source *string, encoding *string) error {
	if source == nil {
		return fmt.Errorf("parameter source is required, but nil was provided")
	}

	if encoding == nil {
		return fmt.Errorf("parameter encoding is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Textencodebase64Parameters(str *string, encoding *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if encoding == nil {
		return fmt.Errorf("parameter encoding is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TimeaddParameters(timestamp *string, duration *string) error {
	if timestamp == nil {
		return fmt.Errorf("parameter timestamp is required, but nil was provided")
	}

	if duration == nil {
		return fmt.Errorf("parameter duration is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TimecmpParameters(timestamp_a *string, timestamp_b *string) error {
	if timestamp_a == nil {
		return fmt.Errorf("parameter timestamp_a is required, but nil was provided")
	}

	if timestamp_b == nil {
		return fmt.Errorf("parameter timestamp_b is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TitleParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_ToboolParameters(v interface{}) error {
	if v == nil {
		return fmt.Errorf("parameter v is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TolistParameters(v interface{}) error {
	if v == nil {
		return fmt.Errorf("parameter v is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TomapParameters(v interface{}) error {
	if v == nil {
		return fmt.Errorf("parameter v is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TonumberParameters(v interface{}) error {
	if v == nil {
		return fmt.Errorf("parameter v is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TosetParameters(v interface{}) error {
	if v == nil {
		return fmt.Errorf("parameter v is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TostringParameters(v interface{}) error {
	if v == nil {
		return fmt.Errorf("parameter v is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TransposeParameters(values interface{}) error {
	if values == nil {
		return fmt.Errorf("parameter values is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TrimParameters(str *string, cutset *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if cutset == nil {
		return fmt.Errorf("parameter cutset is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TrimprefixParameters(str *string, prefix *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if prefix == nil {
		return fmt.Errorf("parameter prefix is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TrimspaceParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TrimsuffixParameters(str *string, suffix *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if suffix == nil {
		return fmt.Errorf("parameter suffix is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_TryParameters(expressions *[]interface{}) error {
	if expressions == nil {
		return fmt.Errorf("parameter expressions is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_UpperParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_UrlencodeParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_Uuidv5Parameters(namespace *string, name *string) error {
	if namespace == nil {
		return fmt.Errorf("parameter namespace is required, but nil was provided")
	}

	if name == nil {
		return fmt.Errorf("parameter name is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_ValuesParameters(mapping interface{}) error {
	if mapping == nil {
		return fmt.Errorf("parameter mapping is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_YamldecodeParameters(src *string) error {
	if src == nil {
		return fmt.Errorf("parameter src is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_YamlencodeParameters(value interface{}) error {
	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func validateFnGenerated_ZipmapParameters(keys *[]*string, values interface{}) error {
	if keys == nil {
		return fmt.Errorf("parameter keys is required, but nil was provided")
	}

	if values == nil {
		return fmt.Errorf("parameter values is required, but nil was provided")
	}

	return nil
}

