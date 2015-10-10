package godoc

var tmplHTML = `
<style>
	h2 {
		margin-top: 0;
	}
	h3 {
		margin-top: 21px;
	}
	h4 {
		margin-top: 28px;
	}

	.sourcegraph-sourcebox .sourcegraph-footer {
		display: none;
	}

	pre {
		background-color: #f5f5f5 !important;
		border: solid 1px #ccc !important;
	}
	code {
		padding-left: 0;
	}

	a.permalink {
		opacity: 0;
	}
	h3:hover a.permalink, h4:hover a.permalink {
		opacity: 1;
	}
</style>
{{with .PDoc}}
	{{$pdoc := .}}
	{{if or .Doc .Consts .Vars .Funcs .Types}}
		<h2>package {{.Name}}</h2>

		<p><code>import "{{.ImportPath}}"</code>

			{{.Doc|godoc_comment}}

			<!-- Index -->
			<h3 id="pkg-index" class="anchor section-header">Index <a class="permalink" href="#pkg-index"><i class="octicon octicon-link"></i></a></h3>

			{{if .Truncated}}
				<div class="alert">The documentation displayed here is incomplete. Use the godoc command to read the complete documentation.</div>
			{{end}}

			{{if not .IsCmd}}
				<ul class="list-unstyled">
					{{if .Consts}}<li><a href="#pkg-constants">Constants</a></li>{{end}}
					{{if .Vars}}<li><a href="#pkg-variables">Variables</a></li>{{end}}
					{{range .Funcs}}<li><a href="#{{.Name}}">{{.Decl.Text}}</a></li>{{end}}
					{{range $t := .Types}}
						<li><a href="#{{.Name}}">type {{.Name}}</a></li>
						{{if or .Funcs .Methods}}<ul>{{end}}
							{{range .Funcs}}<li><a href="#{{.Name}}">{{.Decl.Text}}</a></li>{{end}}
							{{range .Methods}}<li><a href="#{{$t.Name}}.{{.Name}}">{{.Decl.Text}}</a></li>{{end}}
							{{if or .Funcs .Methods}}</ul>{{end}}
					{{end}}
				</ul>

				<!-- Examples -->
				{{with .AllExamples}}
					<h4 class="anchor" id="pkg-examples">Examples <a class="permalink" href="#pkg-examples"><i class="octicon octicon-link"></i></a></h4>
					<ul class="list-unstyled">
						{{range . }}<li><a href="#example-{{.ID}}" onclick="$('#ex-{{.ID}}').addClass('in').removeClass('collapse').height('auto')">{{.Label}}</a></li>{{end}}
					</ul>
				{{else}}
					<span class="anchor" id="pkg-examples"></span>
				{{end}}

				<!-- Files -->
				<h4 class="anchor" id="pkg-files">
					{{with .BrowseURL}}<a href="{{.}}">Package Files</a>{{else}}Package Files{{end}}
					<a class="permalink" href="#pkg-files"><i class="octicon octicon-link"></i></a>
				</h4>

				<p>{{range .Files}}{{if .URL}}<a href="{{.URL}}">{{.Name}}</a>{{else}}{{.Name}}{{end}} {{end}}</p>

				<!-- Contants -->
				{{if .Consts}}
					<h3 class="anchor" id="pkg-constants">Constants <a class="permalink" href="#pkg-constants"><i class="octicon octicon-link"></i></a></h3>
					{{range .Consts}}{{$.PDoc.Code .Pos $.RepoRevSpec}}{{.Doc|godoc_comment}}{{end}}
				{{end}}

				<!-- Variables -->
				{{if .Vars}}
					<h3 class="anchor" id="pkg-variables">Variables <a class="permalink" href="#pkg-variables"><i class="octicon octicon-link"></i></a></h3>
					{{range .Vars}}{{$.PDoc.Code .Pos $.RepoRevSpec}}{{.Doc|godoc_comment}}{{end}}
				{{end}}

				<!-- Functions -->
				{{range .Funcs}}
					<h3 class="anchor" id="{{.Name}}" data-kind="f">
						func <a title="View Source" href="{{printf $.PDoc.LineFmt (index $pdoc.Files .Pos.File).URL .Pos.Line}}">{{.Name}}</a>
						<a class="permalink" href="#{{.Name}}"><i class="octicon octicon-link"></i></a>
					</h3>
					{{$.PDoc.Code .Pos $.RepoRevSpec}}{{.Doc|godoc_comment}}
					{{template "Examples" .|$.PDoc.ObjExamples}}
				{{end}}

				<!-- Types -->
				{{range $t := .Types}}
					<h3 class="anchor" id="{{.Name}}" data-kind="t">
						type <a title="View Source" href="{{printf $.PDoc.LineFmt (index $pdoc.Files .Pos.File).URL .Pos.Line}}">{{.Name}}</a>
						<a class="permalink" href="#{{.Name}}"><i class="octicon octicon-link"></i></a>
					</h3>
					{{$.PDoc.Code .Pos $.RepoRevSpec}}{{.Doc|godoc_comment}}
					{{range .Consts}}{{$.PDoc.Code .Pos $.RepoRevSpec}}{{.Doc|godoc_comment}}{{end}}
					{{range .Vars}}{{$.PDoc.Code .Pos $.RepoRevSpec}}{{.Doc|godoc_comment}}{{end}}
					{{template "Examples" .|$.PDoc.ObjExamples}}

					{{range .Funcs}}
						<h4 class="anchor" id="{{.Name}}" data-kind="f">
							func <a title="View Source" href="{{printf $.PDoc.LineFmt (index $pdoc.Files .Pos.File).URL .Pos.Line}}">{{.Name}}</a>
							<a class="permalink" href="#{{.Name}}"><i class="octicon octicon-link"></i></a>
						</h4>
						{{$.PDoc.Code .Pos $.RepoRevSpec}}{{.Doc|godoc_comment}}
						{{template "Examples" .|$.PDoc.ObjExamples}}
					{{end}}

					{{range .Methods}}
						<h4 class="anchor" id="{{$t.Name}}.{{.Name}}" data-kind="m">
							func ({{.Recv}}) <a title="View Source" href="{{printf $.PDoc.LineFmt (index $pdoc.Files .Pos.File).URL .Pos.Line}}">{{.Name}}</a>
							<a class="permalink" href="#{{$t.Name}}.{{.Name}}"><i class="octicon octicon-link"></i></a>
						</h4>
						{{$.PDoc.Code .Pos $.RepoRevSpec}}{{.Doc|godoc_comment}}
						{{template "Examples" .|$.PDoc.ObjExamples}}
					{{end}}
				{{end}}
			{{end}}
	{{else if not $.Subpkgs}}
			<p>No documentation found.</p>
	{{end}}

{{template "PkgCmdFooter" $}}
{{end}}

{{define "Examples"}}
{{if .}}
		<div class="panel-group">
		{{range .}}
			<div class="panel panel-default anchor" id="example-{{.ID}}">
				<div class="panel-heading"><a class="accordion-toggle" data-toggle="collapse" href="#ex-{{.ID}}">Example{{with .Example.Name}} ({{.}}){{end}}</a></div>
				<div id="ex-{{.ID}}" class="anchor panel-collapse collapse"><div class="panel-body">
					{{with .Example.Doc}}<p>{{.|godoc_comment}}{{end}}
						<p>Code:
							<pre>{{godoc_code .Example.Code nil}}</pre>
							{{with .Example.Output}}<p>Output:<pre>{{.}}</pre>{{end}}
				</div></div>
			</div>
		{{end}}
		</div>
{{end}}
{{end}}

{{define "PkgCmdFooter"}}
<!-- Bugs -->
{{with .PDoc}}
	{{$pdoc := .}}
	{{with .Notes.BUG}}
		<h3 class="anchor" id="pkg-note-bug">Bugs <a class="permalink" href="#pkg-note-bug"><i class="octicon octicon-link"></i></a></h3>{{range .}}<p><a title="View Source" href="{{printf $.PDoc.LineFmt (index $pdoc.Files .Pos.File).URL .Pos.Line}}">☞</a> {{.Body}}{{end}}
	{{end}}
{{end}}

{{if .Subpkgs}}<h3 class="anchor" id="pkg-subdirectories">Directories <a class="permalink" href="#pkg-subdirectories"><i class="octicon octicon-link"></i></a></h3>
		<table class="table table-condensed">
		<thead><tr><th>Path</th></tr></thead>
		<tbody>{{range .Subpkgs}}<tr><td><a href="{{urlToRepoGoDoc $.Repo.URI $.RepoRevSpec.Rev .Path}}">{{pathBase .Path}}</a></tr>{{end}}</tbody>
		</table>
{{end}}
<div class="anchor" id="x-pkginfo">
{{with $.PDoc}}
	{{if .DeadEndFork}}This is a dead-end fork (low number of GitHub stars).{{end}}
{{end}}
{{with $.PDoc.Errors}}
		<p>The <a href="https://golang.org/cmd/go/#Download_and_install_packages_and_dependencies">go get</a>
		command cannot install this package because of the following issues:
		<ul>
			{{range .}}<li>{{.}}{{end}}
	</ul>
{{end}}
</div>
{{end}}
`
