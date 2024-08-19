// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func validateFn_AbsParameters(num *float64) error {
	return nil
}

func validateFn_AbspathParameters(path *string) error {
	return nil
}

func validateFn_AlltrueParameters(list *[]interface{}) error {
	return nil
}

func validateFn_AnytrueParameters(list *[]interface{}) error {
	return nil
}

func validateFn_Base64decodeParameters(str *string) error {
	return nil
}

func validateFn_Base64encodeParameters(str *string) error {
	return nil
}

func validateFn_Base64gzipParameters(str *string) error {
	return nil
}

func validateFn_Base64sha256Parameters(str *string) error {
	return nil
}

func validateFn_Base64sha512Parameters(str *string) error {
	return nil
}

func validateFn_BasenameParameters(path *string) error {
	return nil
}

func validateFn_BcryptParameters(str *string) error {
	return nil
}

func validateFn_CanParameters(expression interface{}) error {
	return nil
}

func validateFn_CeilParameters(num *float64) error {
	return nil
}

func validateFn_ChompParameters(str *string) error {
	return nil
}

func validateFn_ChunklistParameters(list *[]interface{}, size *float64) error {
	return nil
}

func validateFn_CidrhostParameters(prefix *string, hostnum *float64) error {
	return nil
}

func validateFn_CidrnetmaskParameters(prefix *string) error {
	return nil
}

func validateFn_CidrsubnetParameters(prefix *string, newbits *float64, netnum *float64) error {
	return nil
}

func validateFn_CidrsubnetsParameters(prefix *string, newbits *[]*float64) error {
	return nil
}

func validateFn_CoalesceParameters(vals *[]interface{}) error {
	return nil
}

func validateFn_CoalescelistParameters(vals *[]interface{}) error {
	return nil
}

func validateFn_CompactParameters(list *[]*string) error {
	return nil
}

func validateFn_ConcatParameters(seqs *[]interface{}) error {
	return nil
}

func validateFn_ConditionalParameters(condition interface{}, trueValue interface{}, falseValue interface{}) error {
	return nil
}

func validateFn_ContainsParameters(list interface{}, value interface{}) error {
	return nil
}

func validateFn_CsvdecodeParameters(str *string) error {
	return nil
}

func validateFn_DirnameParameters(path *string) error {
	return nil
}

func validateFn_DistinctParameters(list *[]interface{}) error {
	return nil
}

func validateFn_ElementParameters(list interface{}, index *float64) error {
	return nil
}

func validateFn_EndswithParameters(str *string, suffix *string) error {
	return nil
}

func validateFn_FileParameters(path *string) error {
	return nil
}

func validateFn_Filebase64Parameters(path *string) error {
	return nil
}

func validateFn_Filebase64sha256Parameters(path *string) error {
	return nil
}

func validateFn_Filebase64sha512Parameters(path *string) error {
	return nil
}

func validateFn_FileexistsParameters(path *string) error {
	return nil
}

func validateFn_Filemd5Parameters(path *string) error {
	return nil
}

func validateFn_FilesetParameters(path *string, pattern *string) error {
	return nil
}

func validateFn_Filesha1Parameters(path *string) error {
	return nil
}

func validateFn_Filesha256Parameters(path *string) error {
	return nil
}

func validateFn_Filesha512Parameters(path *string) error {
	return nil
}

func validateFn_FlattenParameters(list interface{}) error {
	return nil
}

func validateFn_FloorParameters(num *float64) error {
	return nil
}

func validateFn_FormatParameters(format *string, args *[]interface{}) error {
	return nil
}

func validateFn_FormatdateParameters(format *string, time *string) error {
	return nil
}

func validateFn_FormatlistParameters(format *string, args *[]interface{}) error {
	return nil
}

func validateFn_IndentParameters(spaces *float64, str *string) error {
	return nil
}

func validateFn_IndexParameters(list interface{}, value interface{}) error {
	return nil
}

func validateFn_JoinParameters(separator *string, list *[]*string) error {
	return nil
}

func validateFn_JsondecodeParameters(str *string) error {
	return nil
}

func validateFn_JsonencodeParameters(val interface{}) error {
	return nil
}

func validateFn_KeysParameters(inputMap interface{}) error {
	return nil
}

func validateFn_LengthOfParameters(value interface{}) error {
	return nil
}

func validateFn_LogParameters(num *float64, base *float64) error {
	return nil
}

func validateFn_LookupParameters(inputMap interface{}, key *string) error {
	return nil
}

func validateFn_LookupNestedParameters(inputMap interface{}, path *[]interface{}) error {
	return nil
}

func validateFn_LowerParameters(str *string) error {
	return nil
}

func validateFn_MatchkeysParameters(values *[]interface{}, keys *[]interface{}, searchset *[]interface{}) error {
	return nil
}

func validateFn_MaxParameters(numbers *[]*float64) error {
	return nil
}

func validateFn_Md5Parameters(str *string) error {
	return nil
}

func validateFn_MergeParameters(maps *[]interface{}) error {
	return nil
}

func validateFn_MinParameters(numbers *[]*float64) error {
	return nil
}

func validateFn_NonsensitiveParameters(value interface{}) error {
	return nil
}

func validateFn_OneParameters(list interface{}) error {
	return nil
}

func validateFn_ParseintParameters(number interface{}, base *float64) error {
	return nil
}

func validateFn_PathexpandParameters(path *string) error {
	return nil
}

func validateFn_PowParameters(num *float64, power *float64) error {
	return nil
}

func validateFn_RangeParameters(start *float64, limit *float64) error {
	return nil
}

func validateFn_RawStringParameters(str *string) error {
	return nil
}

func validateFn_RegexParameters(pattern *string, str *string) error {
	return nil
}

func validateFn_RegexallParameters(pattern *string, str *string) error {
	return nil
}

func validateFn_ReplaceParameters(str *string, substr *string, replace *string) error {
	return nil
}

func validateFn_ReverseParameters(list interface{}) error {
	return nil
}

func validateFn_RsadecryptParameters(ciphertext *string, privatekey *string) error {
	return nil
}

func validateFn_SensitiveParameters(value interface{}) error {
	return nil
}

func validateFn_SetintersectionParameters(first_set *[]interface{}, other_sets *[]*[]interface{}) error {
	return nil
}

func validateFn_SetproductParameters(sets *[]interface{}) error {
	return nil
}

func validateFn_SetsubtractParameters(a *[]interface{}, b *[]interface{}) error {
	return nil
}

func validateFn_SetunionParameters(first_set *[]interface{}, other_sets *[]*[]interface{}) error {
	return nil
}

func validateFn_Sha1Parameters(str *string) error {
	return nil
}

func validateFn_Sha256Parameters(str *string) error {
	return nil
}

func validateFn_Sha512Parameters(str *string) error {
	return nil
}

func validateFn_SignumParameters(num *float64) error {
	return nil
}

func validateFn_SliceParameters(list interface{}, start_index *float64, end_index *float64) error {
	return nil
}

func validateFn_SortParameters(list *[]*string) error {
	return nil
}

func validateFn_SplitParameters(separator *string, str *string) error {
	return nil
}

func validateFn_StartswithParameters(str *string, prefix *string) error {
	return nil
}

func validateFn_StrcontainsParameters(str *string, substr *string) error {
	return nil
}

func validateFn_StrrevParameters(str *string) error {
	return nil
}

func validateFn_SubstrParameters(str *string, offset *float64, length *float64) error {
	return nil
}

func validateFn_SumParameters(list interface{}) error {
	return nil
}

func validateFn_TemplatefileParameters(path *string, vars interface{}) error {
	return nil
}

func validateFn_Textdecodebase64Parameters(source *string, encoding *string) error {
	return nil
}

func validateFn_Textencodebase64Parameters(str *string, encoding *string) error {
	return nil
}

func validateFn_TimeaddParameters(timestamp *string, duration *string) error {
	return nil
}

func validateFn_TimecmpParameters(timestamp_a *string, timestamp_b *string) error {
	return nil
}

func validateFn_TitleParameters(str *string) error {
	return nil
}

func validateFn_ToboolParameters(v interface{}) error {
	return nil
}

func validateFn_TolistParameters(v interface{}) error {
	return nil
}

func validateFn_TomapParameters(v interface{}) error {
	return nil
}

func validateFn_TonumberParameters(v interface{}) error {
	return nil
}

func validateFn_TosetParameters(v interface{}) error {
	return nil
}

func validateFn_TostringParameters(v interface{}) error {
	return nil
}

func validateFn_TransposeParameters(values interface{}) error {
	return nil
}

func validateFn_TrimParameters(str *string, cutset *string) error {
	return nil
}

func validateFn_TrimprefixParameters(str *string, prefix *string) error {
	return nil
}

func validateFn_TrimspaceParameters(str *string) error {
	return nil
}

func validateFn_TrimsuffixParameters(str *string, suffix *string) error {
	return nil
}

func validateFn_TryParameters(expressions *[]interface{}) error {
	return nil
}

func validateFn_UpperParameters(str *string) error {
	return nil
}

func validateFn_UrlencodeParameters(str *string) error {
	return nil
}

func validateFn_Uuidv5Parameters(namespace *string, name *string) error {
	return nil
}

func validateFn_ValuesParameters(mapping interface{}) error {
	return nil
}

func validateFn_YamldecodeParameters(src *string) error {
	return nil
}

func validateFn_YamlencodeParameters(value interface{}) error {
	return nil
}

func validateFn_ZipmapParameters(keys *[]*string, values interface{}) error {
	return nil
}

