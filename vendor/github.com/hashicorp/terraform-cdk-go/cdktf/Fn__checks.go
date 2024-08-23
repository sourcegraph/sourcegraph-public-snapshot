// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !no_runtime_type_checking

package cdktf

import (
	"fmt"
)

func validateFn_AbsParameters(num *float64) error {
	if num == nil {
		return fmt.Errorf("parameter num is required, but nil was provided")
	}

	return nil
}

func validateFn_AbspathParameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFn_AlltrueParameters(list *[]interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFn_AnytrueParameters(list *[]interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFn_Base64decodeParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_Base64encodeParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_Base64gzipParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_Base64sha256Parameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_Base64sha512Parameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_BasenameParameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFn_BcryptParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_CanParameters(expression interface{}) error {
	if expression == nil {
		return fmt.Errorf("parameter expression is required, but nil was provided")
	}

	return nil
}

func validateFn_CeilParameters(num *float64) error {
	if num == nil {
		return fmt.Errorf("parameter num is required, but nil was provided")
	}

	return nil
}

func validateFn_ChompParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_ChunklistParameters(list *[]interface{}, size *float64) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	if size == nil {
		return fmt.Errorf("parameter size is required, but nil was provided")
	}

	return nil
}

func validateFn_CidrhostParameters(prefix *string, hostnum *float64) error {
	if prefix == nil {
		return fmt.Errorf("parameter prefix is required, but nil was provided")
	}

	if hostnum == nil {
		return fmt.Errorf("parameter hostnum is required, but nil was provided")
	}

	return nil
}

func validateFn_CidrnetmaskParameters(prefix *string) error {
	if prefix == nil {
		return fmt.Errorf("parameter prefix is required, but nil was provided")
	}

	return nil
}

func validateFn_CidrsubnetParameters(prefix *string, newbits *float64, netnum *float64) error {
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

func validateFn_CidrsubnetsParameters(prefix *string, newbits *[]*float64) error {
	if prefix == nil {
		return fmt.Errorf("parameter prefix is required, but nil was provided")
	}

	if newbits == nil {
		return fmt.Errorf("parameter newbits is required, but nil was provided")
	}

	return nil
}

func validateFn_CoalesceParameters(vals *[]interface{}) error {
	if vals == nil {
		return fmt.Errorf("parameter vals is required, but nil was provided")
	}

	return nil
}

func validateFn_CoalescelistParameters(vals *[]interface{}) error {
	if vals == nil {
		return fmt.Errorf("parameter vals is required, but nil was provided")
	}

	return nil
}

func validateFn_CompactParameters(list *[]*string) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFn_ConcatParameters(seqs *[]interface{}) error {
	if seqs == nil {
		return fmt.Errorf("parameter seqs is required, but nil was provided")
	}

	return nil
}

func validateFn_ConditionalParameters(condition interface{}, trueValue interface{}, falseValue interface{}) error {
	if condition == nil {
		return fmt.Errorf("parameter condition is required, but nil was provided")
	}

	if trueValue == nil {
		return fmt.Errorf("parameter trueValue is required, but nil was provided")
	}

	if falseValue == nil {
		return fmt.Errorf("parameter falseValue is required, but nil was provided")
	}

	return nil
}

func validateFn_ContainsParameters(list interface{}, value interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func validateFn_CsvdecodeParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_DirnameParameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFn_DistinctParameters(list *[]interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFn_ElementParameters(list interface{}, index *float64) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	if index == nil {
		return fmt.Errorf("parameter index is required, but nil was provided")
	}

	return nil
}

func validateFn_EndswithParameters(str *string, suffix *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if suffix == nil {
		return fmt.Errorf("parameter suffix is required, but nil was provided")
	}

	return nil
}

func validateFn_FileParameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFn_Filebase64Parameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFn_Filebase64sha256Parameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFn_Filebase64sha512Parameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFn_FileexistsParameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFn_Filemd5Parameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFn_FilesetParameters(path *string, pattern *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	if pattern == nil {
		return fmt.Errorf("parameter pattern is required, but nil was provided")
	}

	return nil
}

func validateFn_Filesha1Parameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFn_Filesha256Parameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFn_Filesha512Parameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFn_FlattenParameters(list interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFn_FloorParameters(num *float64) error {
	if num == nil {
		return fmt.Errorf("parameter num is required, but nil was provided")
	}

	return nil
}

func validateFn_FormatParameters(format *string, args *[]interface{}) error {
	if format == nil {
		return fmt.Errorf("parameter format is required, but nil was provided")
	}

	if args == nil {
		return fmt.Errorf("parameter args is required, but nil was provided")
	}

	return nil
}

func validateFn_FormatdateParameters(format *string, time *string) error {
	if format == nil {
		return fmt.Errorf("parameter format is required, but nil was provided")
	}

	if time == nil {
		return fmt.Errorf("parameter time is required, but nil was provided")
	}

	return nil
}

func validateFn_FormatlistParameters(format *string, args *[]interface{}) error {
	if format == nil {
		return fmt.Errorf("parameter format is required, but nil was provided")
	}

	if args == nil {
		return fmt.Errorf("parameter args is required, but nil was provided")
	}

	return nil
}

func validateFn_IndentParameters(spaces *float64, str *string) error {
	if spaces == nil {
		return fmt.Errorf("parameter spaces is required, but nil was provided")
	}

	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_IndexParameters(list interface{}, value interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func validateFn_JoinParameters(separator *string, list *[]*string) error {
	if separator == nil {
		return fmt.Errorf("parameter separator is required, but nil was provided")
	}

	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFn_JsondecodeParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_JsonencodeParameters(val interface{}) error {
	if val == nil {
		return fmt.Errorf("parameter val is required, but nil was provided")
	}

	return nil
}

func validateFn_KeysParameters(inputMap interface{}) error {
	if inputMap == nil {
		return fmt.Errorf("parameter inputMap is required, but nil was provided")
	}

	return nil
}

func validateFn_LengthOfParameters(value interface{}) error {
	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func validateFn_LogParameters(num *float64, base *float64) error {
	if num == nil {
		return fmt.Errorf("parameter num is required, but nil was provided")
	}

	if base == nil {
		return fmt.Errorf("parameter base is required, but nil was provided")
	}

	return nil
}

func validateFn_LookupParameters(inputMap interface{}, key *string) error {
	if inputMap == nil {
		return fmt.Errorf("parameter inputMap is required, but nil was provided")
	}

	if key == nil {
		return fmt.Errorf("parameter key is required, but nil was provided")
	}

	return nil
}

func validateFn_LookupNestedParameters(inputMap interface{}, path *[]interface{}) error {
	if inputMap == nil {
		return fmt.Errorf("parameter inputMap is required, but nil was provided")
	}

	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFn_LowerParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_MatchkeysParameters(values *[]interface{}, keys *[]interface{}, searchset *[]interface{}) error {
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

func validateFn_MaxParameters(numbers *[]*float64) error {
	if numbers == nil {
		return fmt.Errorf("parameter numbers is required, but nil was provided")
	}

	return nil
}

func validateFn_Md5Parameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_MergeParameters(maps *[]interface{}) error {
	if maps == nil {
		return fmt.Errorf("parameter maps is required, but nil was provided")
	}

	return nil
}

func validateFn_MinParameters(numbers *[]*float64) error {
	if numbers == nil {
		return fmt.Errorf("parameter numbers is required, but nil was provided")
	}

	return nil
}

func validateFn_NonsensitiveParameters(value interface{}) error {
	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func validateFn_OneParameters(list interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFn_ParseintParameters(number interface{}, base *float64) error {
	if number == nil {
		return fmt.Errorf("parameter number is required, but nil was provided")
	}

	if base == nil {
		return fmt.Errorf("parameter base is required, but nil was provided")
	}

	return nil
}

func validateFn_PathexpandParameters(path *string) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	return nil
}

func validateFn_PowParameters(num *float64, power *float64) error {
	if num == nil {
		return fmt.Errorf("parameter num is required, but nil was provided")
	}

	if power == nil {
		return fmt.Errorf("parameter power is required, but nil was provided")
	}

	return nil
}

func validateFn_RangeParameters(start *float64, limit *float64) error {
	if start == nil {
		return fmt.Errorf("parameter start is required, but nil was provided")
	}

	if limit == nil {
		return fmt.Errorf("parameter limit is required, but nil was provided")
	}

	return nil
}

func validateFn_RawStringParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_RegexParameters(pattern *string, str *string) error {
	if pattern == nil {
		return fmt.Errorf("parameter pattern is required, but nil was provided")
	}

	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_RegexallParameters(pattern *string, str *string) error {
	if pattern == nil {
		return fmt.Errorf("parameter pattern is required, but nil was provided")
	}

	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_ReplaceParameters(str *string, substr *string, replace *string) error {
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

func validateFn_ReverseParameters(list interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFn_RsadecryptParameters(ciphertext *string, privatekey *string) error {
	if ciphertext == nil {
		return fmt.Errorf("parameter ciphertext is required, but nil was provided")
	}

	if privatekey == nil {
		return fmt.Errorf("parameter privatekey is required, but nil was provided")
	}

	return nil
}

func validateFn_SensitiveParameters(value interface{}) error {
	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func validateFn_SetintersectionParameters(first_set *[]interface{}, other_sets *[]*[]interface{}) error {
	if first_set == nil {
		return fmt.Errorf("parameter first_set is required, but nil was provided")
	}

	if other_sets == nil {
		return fmt.Errorf("parameter other_sets is required, but nil was provided")
	}

	return nil
}

func validateFn_SetproductParameters(sets *[]interface{}) error {
	if sets == nil {
		return fmt.Errorf("parameter sets is required, but nil was provided")
	}

	return nil
}

func validateFn_SetsubtractParameters(a *[]interface{}, b *[]interface{}) error {
	if a == nil {
		return fmt.Errorf("parameter a is required, but nil was provided")
	}

	if b == nil {
		return fmt.Errorf("parameter b is required, but nil was provided")
	}

	return nil
}

func validateFn_SetunionParameters(first_set *[]interface{}, other_sets *[]*[]interface{}) error {
	if first_set == nil {
		return fmt.Errorf("parameter first_set is required, but nil was provided")
	}

	if other_sets == nil {
		return fmt.Errorf("parameter other_sets is required, but nil was provided")
	}

	return nil
}

func validateFn_Sha1Parameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_Sha256Parameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_Sha512Parameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_SignumParameters(num *float64) error {
	if num == nil {
		return fmt.Errorf("parameter num is required, but nil was provided")
	}

	return nil
}

func validateFn_SliceParameters(list interface{}, start_index *float64, end_index *float64) error {
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

func validateFn_SortParameters(list *[]*string) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFn_SplitParameters(separator *string, str *string) error {
	if separator == nil {
		return fmt.Errorf("parameter separator is required, but nil was provided")
	}

	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_StartswithParameters(str *string, prefix *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if prefix == nil {
		return fmt.Errorf("parameter prefix is required, but nil was provided")
	}

	return nil
}

func validateFn_StrcontainsParameters(str *string, substr *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if substr == nil {
		return fmt.Errorf("parameter substr is required, but nil was provided")
	}

	return nil
}

func validateFn_StrrevParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_SubstrParameters(str *string, offset *float64, length *float64) error {
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

func validateFn_SumParameters(list interface{}) error {
	if list == nil {
		return fmt.Errorf("parameter list is required, but nil was provided")
	}

	return nil
}

func validateFn_TemplatefileParameters(path *string, vars interface{}) error {
	if path == nil {
		return fmt.Errorf("parameter path is required, but nil was provided")
	}

	if vars == nil {
		return fmt.Errorf("parameter vars is required, but nil was provided")
	}

	return nil
}

func validateFn_Textdecodebase64Parameters(source *string, encoding *string) error {
	if source == nil {
		return fmt.Errorf("parameter source is required, but nil was provided")
	}

	if encoding == nil {
		return fmt.Errorf("parameter encoding is required, but nil was provided")
	}

	return nil
}

func validateFn_Textencodebase64Parameters(str *string, encoding *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if encoding == nil {
		return fmt.Errorf("parameter encoding is required, but nil was provided")
	}

	return nil
}

func validateFn_TimeaddParameters(timestamp *string, duration *string) error {
	if timestamp == nil {
		return fmt.Errorf("parameter timestamp is required, but nil was provided")
	}

	if duration == nil {
		return fmt.Errorf("parameter duration is required, but nil was provided")
	}

	return nil
}

func validateFn_TimecmpParameters(timestamp_a *string, timestamp_b *string) error {
	if timestamp_a == nil {
		return fmt.Errorf("parameter timestamp_a is required, but nil was provided")
	}

	if timestamp_b == nil {
		return fmt.Errorf("parameter timestamp_b is required, but nil was provided")
	}

	return nil
}

func validateFn_TitleParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_ToboolParameters(v interface{}) error {
	if v == nil {
		return fmt.Errorf("parameter v is required, but nil was provided")
	}

	return nil
}

func validateFn_TolistParameters(v interface{}) error {
	if v == nil {
		return fmt.Errorf("parameter v is required, but nil was provided")
	}

	return nil
}

func validateFn_TomapParameters(v interface{}) error {
	if v == nil {
		return fmt.Errorf("parameter v is required, but nil was provided")
	}

	return nil
}

func validateFn_TonumberParameters(v interface{}) error {
	if v == nil {
		return fmt.Errorf("parameter v is required, but nil was provided")
	}

	return nil
}

func validateFn_TosetParameters(v interface{}) error {
	if v == nil {
		return fmt.Errorf("parameter v is required, but nil was provided")
	}

	return nil
}

func validateFn_TostringParameters(v interface{}) error {
	if v == nil {
		return fmt.Errorf("parameter v is required, but nil was provided")
	}

	return nil
}

func validateFn_TransposeParameters(values interface{}) error {
	if values == nil {
		return fmt.Errorf("parameter values is required, but nil was provided")
	}

	return nil
}

func validateFn_TrimParameters(str *string, cutset *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if cutset == nil {
		return fmt.Errorf("parameter cutset is required, but nil was provided")
	}

	return nil
}

func validateFn_TrimprefixParameters(str *string, prefix *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if prefix == nil {
		return fmt.Errorf("parameter prefix is required, but nil was provided")
	}

	return nil
}

func validateFn_TrimspaceParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_TrimsuffixParameters(str *string, suffix *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	if suffix == nil {
		return fmt.Errorf("parameter suffix is required, but nil was provided")
	}

	return nil
}

func validateFn_TryParameters(expressions *[]interface{}) error {
	if expressions == nil {
		return fmt.Errorf("parameter expressions is required, but nil was provided")
	}

	return nil
}

func validateFn_UpperParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_UrlencodeParameters(str *string) error {
	if str == nil {
		return fmt.Errorf("parameter str is required, but nil was provided")
	}

	return nil
}

func validateFn_Uuidv5Parameters(namespace *string, name *string) error {
	if namespace == nil {
		return fmt.Errorf("parameter namespace is required, but nil was provided")
	}

	if name == nil {
		return fmt.Errorf("parameter name is required, but nil was provided")
	}

	return nil
}

func validateFn_ValuesParameters(mapping interface{}) error {
	if mapping == nil {
		return fmt.Errorf("parameter mapping is required, but nil was provided")
	}

	return nil
}

func validateFn_YamldecodeParameters(src *string) error {
	if src == nil {
		return fmt.Errorf("parameter src is required, but nil was provided")
	}

	return nil
}

func validateFn_YamlencodeParameters(value interface{}) error {
	if value == nil {
		return fmt.Errorf("parameter value is required, but nil was provided")
	}

	return nil
}

func validateFn_ZipmapParameters(keys *[]*string, values interface{}) error {
	if keys == nil {
		return fmt.Errorf("parameter keys is required, but nil was provided")
	}

	if values == nil {
		return fmt.Errorf("parameter values is required, but nil was provided")
	}

	return nil
}

