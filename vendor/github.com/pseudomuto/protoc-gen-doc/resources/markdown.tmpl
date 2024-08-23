# Protocol Documentation
<a name="top"></a>

## Table of Contents
{{range .Files}}
{{$file_name := .Name}}- [{{.Name}}](#{{.Name | anchor}})
  {{- if .Messages }}
  {{range .Messages}}  - [{{.LongName}}](#{{.FullName | anchor}})
  {{end}}
  {{- end -}}
  {{- if .Enums }}
  {{range .Enums}}  - [{{.LongName}}](#{{.FullName | anchor}})
  {{end}}
  {{- end -}}
  {{- if .Extensions }}
  {{range .Extensions}}  - [File-level Extensions](#{{$file_name | anchor}}-extensions)
  {{end}}
  {{- end -}}
  {{- if .Services }}
  {{range .Services}}  - [{{.Name}}](#{{.FullName | anchor}})
  {{end}}
  {{- end -}}
{{end}}
- [Scalar Value Types](#scalar-value-types)

{{range .Files}}
{{$file_name := .Name}}
<a name="{{.Name | anchor}}"></a>
<p align="right"><a href="#top">Top</a></p>

## {{.Name}}
{{.Description}}

{{range .Messages}}
<a name="{{.FullName | anchor}}"></a>

### {{.LongName}}
{{.Description}}

{{if .HasFields}}
| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
{{range .Fields -}}
  | {{.Name}} | [{{.LongType}}](#{{.FullType | anchor}}) | {{.Label}} | {{if (index .Options "deprecated"|default false)}}**Deprecated.** {{end}}{{nobr .Description}}{{if .DefaultValue}} Default: {{.DefaultValue}}{{end}} |
{{end}}
{{end}}

{{if .HasExtensions}}
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
{{range .Extensions -}}
  | {{.Name}} | {{.LongType}} | {{.ContainingLongType}} | {{.Number}} | {{nobr .Description}}{{if .DefaultValue}} Default: {{.DefaultValue}}{{end}} |
{{end}}
{{end}}

{{end}} <!-- end messages -->

{{range .Enums}}
<a name="{{.FullName | anchor}}"></a>

### {{.LongName}}
{{.Description}}

| Name | Number | Description |
| ---- | ------ | ----------- |
{{range .Values -}}
  | {{.Name}} | {{.Number}} | {{nobr .Description}} |
{{end}}

{{end}} <!-- end enums -->

{{if .HasExtensions}}
<a name="{{$file_name | anchor}}-extensions"></a>

### File-level Extensions
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
{{range .Extensions -}}
  | {{.Name}} | {{.LongType}} | {{.ContainingLongType}} | {{.Number}} | {{nobr .Description}}{{if .DefaultValue}} Default: `{{.DefaultValue}}`{{end}} |
{{end}}
{{end}} <!-- end HasExtensions -->

{{range .Services}}
<a name="{{.FullName | anchor}}"></a>

### {{.Name}}
{{.Description}}

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
{{range .Methods -}}
  | {{.Name}} | [{{.RequestLongType}}](#{{.RequestFullType | anchor}}){{if .RequestStreaming}} stream{{end}} | [{{.ResponseLongType}}](#{{.ResponseFullType | anchor}}){{if .ResponseStreaming}} stream{{end}} | {{nobr .Description}} |
{{end}}
{{end}} <!-- end services -->

{{end}}

## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
{{range .Scalars -}}
  | <a name="{{.ProtoType | anchor}}" /> {{.ProtoType}} | {{.Notes}} | {{.CppType}} | {{.JavaType}} | {{.PythonType}} | {{.GoType}} | {{.CSharp}} | {{.PhpType}} | {{.RubyType}} |
{{end}}
