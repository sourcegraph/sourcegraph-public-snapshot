// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build no_runtime_type_checking

package cdktf

// Building without runtime type checking enabled, so all the below just return nil

func validateFnGenerated_AbsParameters(num *float64) error {
	return nil
}

func validateFnGenerated_AbspathParameters(path *string) error {
	return nil
}

func validateFnGenerated_AlltrueParameters(list *[]interface{}) error {
	return nil
}

func validateFnGenerated_AnytrueParameters(list *[]interface{}) error {
	return nil
}

func validateFnGenerated_Base64decodeParameters(str *string) error {
	return nil
}

func validateFnGenerated_Base64encodeParameters(str *string) error {
	return nil
}

func validateFnGenerated_Base64gzipParameters(str *string) error {
	return nil
}

func validateFnGenerated_Base64sha256Parameters(str *string) error {
	return nil
}

func validateFnGenerated_Base64sha512Parameters(str *string) error {
	return nil
}

func validateFnGenerated_BasenameParameters(path *string) error {
	return nil
}

func validateFnGenerated_CanParameters(expression interface{}) error {
	return nil
}

func validateFnGenerated_CeilParameters(num *float64) error {
	return nil
}

func validateFnGenerated_ChompParameters(str *string) error {
	return nil
}

func validateFnGenerated_ChunklistParameters(list *[]interface{}, size *float64) error {
	return nil
}

func validateFnGenerated_CidrhostParameters(prefix *string, hostnum *float64) error {
	return nil
}

func validateFnGenerated_CidrnetmaskParameters(prefix *string) error {
	return nil
}

func validateFnGenerated_CidrsubnetParameters(prefix *string, newbits *float64, netnum *float64) error {
	return nil
}

func validateFnGenerated_CidrsubnetsParameters(prefix *string, newbits *[]*float64) error {
	return nil
}

func validateFnGenerated_CoalesceParameters(vals *[]interface{}) error {
	return nil
}

func validateFnGenerated_CoalescelistParameters(vals *[]interface{}) error {
	return nil
}

func validateFnGenerated_CompactParameters(list *[]*string) error {
	return nil
}

func validateFnGenerated_ConcatParameters(seqs *[]interface{}) error {
	return nil
}

func validateFnGenerated_ContainsParameters(list interface{}, value interface{}) error {
	return nil
}

func validateFnGenerated_CsvdecodeParameters(str *string) error {
	return nil
}

func validateFnGenerated_DirnameParameters(path *string) error {
	return nil
}

func validateFnGenerated_DistinctParameters(list *[]interface{}) error {
	return nil
}

func validateFnGenerated_ElementParameters(list interface{}, index *float64) error {
	return nil
}

func validateFnGenerated_EndswithParameters(str *string, suffix *string) error {
	return nil
}

func validateFnGenerated_FileParameters(path *string) error {
	return nil
}

func validateFnGenerated_Filebase64Parameters(path *string) error {
	return nil
}

func validateFnGenerated_Filebase64sha256Parameters(path *string) error {
	return nil
}

func validateFnGenerated_Filebase64sha512Parameters(path *string) error {
	return nil
}

func validateFnGenerated_FileexistsParameters(path *string) error {
	return nil
}

func validateFnGenerated_Filemd5Parameters(path *string) error {
	return nil
}

func validateFnGenerated_FilesetParameters(path *string, pattern *string) error {
	return nil
}

func validateFnGenerated_Filesha1Parameters(path *string) error {
	return nil
}

func validateFnGenerated_Filesha256Parameters(path *string) error {
	return nil
}

func validateFnGenerated_Filesha512Parameters(path *string) error {
	return nil
}

func validateFnGenerated_FlattenParameters(list interface{}) error {
	return nil
}

func validateFnGenerated_FloorParameters(num *float64) error {
	return nil
}

func validateFnGenerated_FormatParameters(format *string, args *[]interface{}) error {
	return nil
}

func validateFnGenerated_FormatdateParameters(format *string, time *string) error {
	return nil
}

func validateFnGenerated_FormatlistParameters(format *string, args *[]interface{}) error {
	return nil
}

func validateFnGenerated_IndentParameters(spaces *float64, str *string) error {
	return nil
}

func validateFnGenerated_IndexParameters(list interface{}, value interface{}) error {
	return nil
}

func validateFnGenerated_JsondecodeParameters(str *string) error {
	return nil
}

func validateFnGenerated_JsonencodeParameters(val interface{}) error {
	return nil
}

func validateFnGenerated_KeysParameters(inputMap interface{}) error {
	return nil
}

func validateFnGenerated_LengthOfParameters(value interface{}) error {
	return nil
}

func validateFnGenerated_LogParameters(num *float64, base *float64) error {
	return nil
}

func validateFnGenerated_LowerParameters(str *string) error {
	return nil
}

func validateFnGenerated_MatchkeysParameters(values *[]interface{}, keys *[]interface{}, searchset *[]interface{}) error {
	return nil
}

func validateFnGenerated_MaxParameters(numbers *[]*float64) error {
	return nil
}

func validateFnGenerated_Md5Parameters(str *string) error {
	return nil
}

func validateFnGenerated_MergeParameters(maps *[]interface{}) error {
	return nil
}

func validateFnGenerated_MinParameters(numbers *[]*float64) error {
	return nil
}

func validateFnGenerated_NonsensitiveParameters(value interface{}) error {
	return nil
}

func validateFnGenerated_OneParameters(list interface{}) error {
	return nil
}

func validateFnGenerated_ParseintParameters(number interface{}, base *float64) error {
	return nil
}

func validateFnGenerated_PathexpandParameters(path *string) error {
	return nil
}

func validateFnGenerated_PowParameters(num *float64, power *float64) error {
	return nil
}

func validateFnGenerated_RegexParameters(pattern *string, str *string) error {
	return nil
}

func validateFnGenerated_RegexallParameters(pattern *string, str *string) error {
	return nil
}

func validateFnGenerated_ReplaceParameters(str *string, substr *string, replace *string) error {
	return nil
}

func validateFnGenerated_ReverseParameters(list interface{}) error {
	return nil
}

func validateFnGenerated_RsadecryptParameters(ciphertext *string, privatekey *string) error {
	return nil
}

func validateFnGenerated_SensitiveParameters(value interface{}) error {
	return nil
}

func validateFnGenerated_SetintersectionParameters(first_set *[]interface{}, other_sets *[]*[]interface{}) error {
	return nil
}

func validateFnGenerated_SetproductParameters(sets *[]interface{}) error {
	return nil
}

func validateFnGenerated_SetsubtractParameters(a *[]interface{}, b *[]interface{}) error {
	return nil
}

func validateFnGenerated_SetunionParameters(first_set *[]interface{}, other_sets *[]*[]interface{}) error {
	return nil
}

func validateFnGenerated_Sha1Parameters(str *string) error {
	return nil
}

func validateFnGenerated_Sha256Parameters(str *string) error {
	return nil
}

func validateFnGenerated_Sha512Parameters(str *string) error {
	return nil
}

func validateFnGenerated_SignumParameters(num *float64) error {
	return nil
}

func validateFnGenerated_SliceParameters(list interface{}, start_index *float64, end_index *float64) error {
	return nil
}

func validateFnGenerated_SortParameters(list *[]*string) error {
	return nil
}

func validateFnGenerated_SplitParameters(separator *string, str *string) error {
	return nil
}

func validateFnGenerated_StartswithParameters(str *string, prefix *string) error {
	return nil
}

func validateFnGenerated_StrcontainsParameters(str *string, substr *string) error {
	return nil
}

func validateFnGenerated_StrrevParameters(str *string) error {
	return nil
}

func validateFnGenerated_SubstrParameters(str *string, offset *float64, length *float64) error {
	return nil
}

func validateFnGenerated_SumParameters(list interface{}) error {
	return nil
}

func validateFnGenerated_TemplatefileParameters(path *string, vars interface{}) error {
	return nil
}

func validateFnGenerated_Textdecodebase64Parameters(source *string, encoding *string) error {
	return nil
}

func validateFnGenerated_Textencodebase64Parameters(str *string, encoding *string) error {
	return nil
}

func validateFnGenerated_TimeaddParameters(timestamp *string, duration *string) error {
	return nil
}

func validateFnGenerated_TimecmpParameters(timestamp_a *string, timestamp_b *string) error {
	return nil
}

func validateFnGenerated_TitleParameters(str *string) error {
	return nil
}

func validateFnGenerated_ToboolParameters(v interface{}) error {
	return nil
}

func validateFnGenerated_TolistParameters(v interface{}) error {
	return nil
}

func validateFnGenerated_TomapParameters(v interface{}) error {
	return nil
}

func validateFnGenerated_TonumberParameters(v interface{}) error {
	return nil
}

func validateFnGenerated_TosetParameters(v interface{}) error {
	return nil
}

func validateFnGenerated_TostringParameters(v interface{}) error {
	return nil
}

func validateFnGenerated_TransposeParameters(values interface{}) error {
	return nil
}

func validateFnGenerated_TrimParameters(str *string, cutset *string) error {
	return nil
}

func validateFnGenerated_TrimprefixParameters(str *string, prefix *string) error {
	return nil
}

func validateFnGenerated_TrimspaceParameters(str *string) error {
	return nil
}

func validateFnGenerated_TrimsuffixParameters(str *string, suffix *string) error {
	return nil
}

func validateFnGenerated_TryParameters(expressions *[]interface{}) error {
	return nil
}

func validateFnGenerated_UpperParameters(str *string) error {
	return nil
}

func validateFnGenerated_UrlencodeParameters(str *string) error {
	return nil
}

func validateFnGenerated_Uuidv5Parameters(namespace *string, name *string) error {
	return nil
}

func validateFnGenerated_ValuesParameters(mapping interface{}) error {
	return nil
}

func validateFnGenerated_YamldecodeParameters(src *string) error {
	return nil
}

func validateFnGenerated_YamlencodeParameters(value interface{}) error {
	return nil
}

func validateFnGenerated_ZipmapParameters(keys *[]*string, values interface{}) error {
	return nil
}

