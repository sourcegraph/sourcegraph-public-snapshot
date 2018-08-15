// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web

import (
	"html/template"
	"log"
)

// Top provides the standard templates in parsed form
var Top = template.New("top").Funcs(Funcmap)

// TemplateText contains the text of the standard templates.
var TemplateText = map[string]string{

	"didyoumean": `
<html>
<head>
  <title>Error</title>
</head>
<body>
  <p>{{.Message}}. Did you mean <a href="/search?q={{.Suggestion}}">{{.Suggestion}}</a> ?
</body>
</html>
`,

	"head": `
<head>
<meta charset="utf-8">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<meta name="viewport" content="width=device-width, initial-scale=1">
<!-- Licensed under MIT (https://github.com/twbs/bootstrap/blob/master/LICENSE) -->
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">
<style>
  #navsearchbox { width: 350px !important; }
  #maxhits { width: 100px !important; }
  .label-dup {
    border-width: 1px !important;
    border-style: solid !important;
    border-color: #aaa !important;
    color: black;
  }
  .noselect {
    user-select: none;
  }
  a.label-dup:hover {
    color: black;
    background: #ddd;
  }
  .result {
    display: block;
    content: " ";
    visibility: hidden;
  }
  .container-results {
     overflow: auto;
     max-height: calc(100% - 72px);
  }
  .inline-pre {
     border: unset;
     background-color: unset;
     margin: unset;
     padding: unset;
     overflow: unset;
  }
  :target { background-color: #ccf; }
  table tbody tr td { border: none !important; padding: 2px !important; }
</style>
</head>
  `,

	"jsdep": `
<script src="https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.min.js"></script>
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>
`,

	// the template for the search box.
	"searchbox": `
<form action="search">
  <div class="form-group form-group-lg">
    <div class="input-group input-group-lg">
      <input class="form-control" placeholder="Search for some code..." autofocus
              {{if .Query}}
              value={{.Query}}
              {{end}}
              id="searchbox" type="text" name="q">
      <div class="input-group-btn">
        <button class="btn btn-primary">Search</button>
      </div>
    </div>
  </div>
</form>
`,

	"navbar": `
<nav class="navbar navbar-default">
  <div class="container-fluid">
    <div class="navbar-header">
      <a class="navbar-brand" href="/">Zoekt</a>
      <button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#navbar-collapse" aria-expanded="false">
        <span class="sr-only">Toggle navigation</span>
        <span class="icon-bar"></span>
        <span class="icon-bar"></span>
        <span class="icon-bar"></span>
      </button>
    </div>
    <div class="navbar-collapse collapse" id="navbar-collapse" aria-expanded="false" style="height: 1px;">
      <form class="navbar-form navbar-left" action="search">
        <div class="form-group">
          <input class="form-control"
                placeholder="Search for some code..." role="search"
                id="navsearchbox" type="text" name="q" autofocus
                {{if .Query}}
                value={{.Query}}
                {{end}}>
          <div class="input-group">
            <div class="input-group-addon">Max Results</div>
            <input class="form-control" type="number" id="maxhits" name="num" value="{{.Num}}">
          </div>
          <button class="btn btn-primary">Search</button>
        </div>
      </form>
    </div>
  </div>
</nav>
`,
	// search box for the entry page.
	"search": `
<html>
{{template "head"}}
<title>Zoekt, en gij zult spinazie eten</title>
<body>
  <div class="jumbotron">
    <div class="container">
      {{template "searchbox" .Last}}
    </div>
  </div>

  <div class="container">
    <div class="row">
      <div class="col-md-8">
        <h3>Search examples:</h3>
        <dl class="dl-horizontal">
          <dt><a href="search?q=needle">needle</a></dt><dd>search for "needle"</dd>
          <dt><a href="search?q=thread+or+needle">thread or needle</a></dt><dd>search for either "thread" or "needle"</dd>
          <dt><a href="search?q=class+needle">class needle</a></span></dt><dd>search for files containing both "class" and "needle"</dd>
          <dt><a href="search?q=class+Needle">class Needle</a></dt><dd>search for files containing both "class" (case insensitive) and "Needle" (case sensitive)</dd>
          <dt><a href="search?q=class+Needle+case:yes">class Needle case:yes</a></dt><dd>search for files containing "class" and "Needle", both case sensitively</dd>
          <dt><a href="search?q=%22class Needle%22">"class Needle"</a></dt><dd>search for files with the phrase "class Needle"</dd>
          <dt><a href="search?q=needle+-hay">needle -hay</a></dt><dd>search for files with the word "needle" but not the word "hay"</dd>
          <dt><a href="search?q=path+file:java">path file:java</a></dt><dd>search for the word "path" in files whose name contains "java"</dd>
          <dt><a href="search?q=needle+lang%3Apython&num=50">needle lang:python</a></dt><dd>search for "needle" in Python source code</dd>
          <dt><a href="search?q=f:%5C.c%24">f:\.c$</a></dt><dd>search for files whose name ends with ".c"</dd>
          <dt><a href="search?q=path+-file:java">path -file:java</a></dt><dd>search for the word "path" excluding files whose name contains "java"</dd>
          <dt><a href="search?q=foo.*bar">foo.*bar</a></dt><dd>search for the regular expression "foo.*bar"</dd>
          <dt><a href="search?q=-%28Path File%29 Stream">-(Path File) Stream</a></dt><dd>search "Stream", but exclude files containing both "Path" and "File"</dd>
          <dt><a href="search?q=-Path%5c+file+Stream">-Path\ file Stream</a></dt><dd>search "Stream", but exclude files containing "Path File"</dd>
          <dt><a href="search?q=sym:data">sym:data</a></span></dt><dd>search for symbol definitions containing "data"</dd>
          <dt><a href="search?q=phone+r:droid">phone r:droid</a></dt><dd>search for "phone" in repositories whose name contains "droid"</dd>
          <dt><a href="search?q=phone+b:master">phone b:master</a></dt><dd>for Git repos, find "phone" in files in branches whose name contains "master".</dd>
          <dt><a href="search?q=phone+b:HEAD">phone b:HEAD</a></dt><dd>for Git repos, find "phone" in the default ('HEAD') branch.</dd>
        </dl>
      </div>
      <div class="col-md-4">
        <h3>To list repositories, try:</h3>
        <dl class="dl-horizontal">
          <dt><a href="search?q=r:droid">r:droid</a></dt><dd>list repositories whose name contains "droid".</dd>
          <dt><a href="search?q=r:go+-r:google">r:go -r:google</a></dt><dd>list repositories whose name contains "go" but not "google".</dd>
        </dl>
      </div>
    </div>
  </div>
  <nav class="navbar navbar-default navbar-bottom">
    <div class="container">
      {{template "footerBoilerplate"}}
      <p class="navbar-text navbar-right">
        Used {{HumanUnit .Stats.IndexBytes}} mem for
        {{.Stats.Documents}} documents ({{HumanUnit .Stats.ContentBytes}})
        from {{.Stats.Repos}} repositories.
      </p>
    </div>
  </nav>
</body>
</html>
`,
	"footerBoilerplate": `<a class="navbar-text" href="about">About</a>`,
	"results": `
<html>
{{template "head"}}
<title>Results for {{.QueryStr}}</title>
<body id="results">
  {{template "navbar" .Last}}
  <div class="container-fluid container-results">
    <h5>
      {{if .Stats.Crashes}}<br><b>{{.Stats.Crashes}} shards crashed</b><br>{{end}}
      {{ $fileCount := len .FileMatches }}
      Found {{.Stats.MatchCount}} results in {{.Stats.FileCount}} files{{if or (lt $fileCount .Stats.FileCount) (or (gt .Stats.ShardsSkipped 0) (gt .Stats.FilesSkipped 0)) }},
        showing top {{ $fileCount }} files (<a rel="nofollow"
           href="search?q={{.Last.Query}}&num={{More .Last.Num}}">show more</a>).
      {{else}}.{{end}}
    </h5>
    {{range .FileMatches}}
    <table class="table table-hover table-condensed">
      <thead>
        <tr>
          <th>
            {{if .URL}}<a name="{{.ResultID}}" class="result"></a><a href="{{.URL}}" >{{else}}<a name="{{.ResultID}}">{{end}}
            <small>
              {{.Repo}}:{{.FileName}}</a>:
              <span style="font-weight: normal">[ {{if .Branches}}{{range .Branches}}<span class="label label-default">{{.}}</span>,{{end}}{{end}} ]</span>
              {{if .Language}}<span class="label label-primary">{{.Language}}</span>{{end}}
              {{if .DuplicateID}}<a class="label label-dup" href="#{{.DuplicateID}}">Duplicate result</a>{{end}}
            </small>
          </th>
        </tr>
      </thead>
      {{if not .DuplicateID}}
      <tbody>
        {{range .Matches}}
        <tr>
          <td style="background-color: rgba(238, 238, 255, 0.6);">
            <pre class="inline-pre"><span class="noselect">{{if .URL}}<a href="{{.URL}}">{{end}}<u>{{.LineNum}}</u>{{if .URL}}</a>{{end}}: </span>{{range .Fragments}}{{LimitPre 100 .Pre}}<b>{{.Match}}</b>{{LimitPost 100 .Post}}{{end}}</pre>
          </td>
        </tr>
        {{end}}
      </tbody>
      {{end}}
    </table>
    {{end}}

  <nav class="navbar navbar-default navbar-bottom">
    <div class="container">
      {{template "footerBoilerplate"}}
      <p class="navbar-text navbar-right">
      Took {{.Stats.Duration}}{{if .Stats.Wait}}(queued: {{.Stats.Wait}}){{end}} for
      {{HumanUnit .Stats.IndexBytesLoaded}}B index data,
      {{.Stats.NgramMatches}} ngram matches,
      {{.Stats.FilesConsidered}} docs considered,
      {{.Stats.FilesLoaded}} docs ({{HumanUnit .Stats.ContentBytesLoaded}}B)
      loaded{{if or .Stats.FilesSkipped .Stats.ShardsSkipped}},
      {{.Stats.FilesSkipped}} docs and {{.Stats.ShardsSkipped}} shards skipped{{else}}.{{end}}
      </p>
    </div>
  </nav>
  </div>
  {{ template "jsdep"}}
</body>
</html>
`,

	"repolist": `
<html>
{{template "head"}}
<body id="results">
  <div class="container">
    {{template "navbar" .Last}}
    <div><b>
    Found {{.Stats.Repos}} repositories ({{.Stats.Documents}} files, {{HumanUnit .Stats.ContentBytes}}b content)
    </b></div>
    <table class="table table-hover table-condensed">
      <thead>
	<tr>
	  <th>Name <a href="/search?q={{.Last.Query}}&order=name">▼</a><a href="/search?q={{.Last.Query}}&order=revname">▲</a></th>
	  <th>Last updated <a href="/search?q={{.Last.Query}}&order=revtime">▼</a><a href="/search?q={{.Last.Query}}&order=time">▲</a></th>
	  <th>Branches</th>
	  <th>Size <a href="/search?q={{.Last.Query}}&order=revsize">▼</a><a href="/search?q={{.Last.Query}}&order=size">▲</a></th>
	</tr>
      </thead>
      <tbody>
	{{range .Repos}}
	<tr>
	  <td>{{if .URL}}<a href="{{.URL}}">{{end}}{{.Name}}{{if .URL}}</a>{{end}}</td>
	  <td><small>{{.IndexTime.Format "Jan 02, 2006 15:04"}}</small></td>
	  <td style="vertical-align: middle;">
	    {{range .Branches}}
	    {{if .URL}}<tt><a class="label label-default small" href="{{.URL}}">{{end}}{{.Name}}{{if .URL}}</a> </tt>{{end}}&nbsp;
	    {{end}}
	  </td>
	  <td><small>{{HumanUnit .Files}} files ({{HumanUnit .Size}})</small></td>
	</tr>
	{{end}}
      </tbody>
    </table>
  </div>

  <nav class="navbar navbar-default navbar-bottom">
    <div class="container">
      {{template "footerBoilerplate"}}
      <p class="navbar-text navbar-right">
      </p>
    </div>
  </nav>

  {{ template "jsdep"}}
</body>
</html>
`,

	"print": `
<html>
  {{template "head"}}
  <title>{{.Repo}}:{{.Name}}</title>
<body id="results">
  {{template "navbar" .Last}}
  <div class="container-fluid container-results" >
     <div class="table table-hover table-condensed" style="overflow:auto; background: #eef;">
       {{ range $index, $ln := .Lines}}
	 <pre id="l{{Inc $index}}" class="inline-pre"><span class="noselect"><a href="#l{{Inc $index}}">{{Inc $index}}</a>: </span>{{$ln}}</pre>
       {{end}}
     </div>
  <nav class="navbar navbar-default navbar-bottom">
    <div class="container">
      {{template "footerBoilerplate"}}
      <p class="navbar-text navbar-right">
      </p>
    </div>
  </nav>
  </div>
 {{ template "jsdep"}}
</body>
</html>
`,

	"about": `

<html>
  {{template "head"}}
  <title>About <em>zoekt</em></title>
<body>


  <div class="jumbotron">
    <div class="container">
      {{template "searchbox" .Last}}
    </div>
  </div>

  <div class="container">
    <p>
      This is <a href="http://github.com/google/zoekt"><em>zoekt</em> (IPA: /zukt/)</a>,
      an open-source full text search engine. It's pronounced roughly as you would
      pronounce "zooked" in English.
    </p>
    <p>
    {{if .Version}}<em>Zoekt</em> version {{.Version}}, uptime{{else}}Uptime{{end}} {{.Uptime}}
    </p>

    <p>
    Used {{HumanUnit .Stats.IndexBytes}} memory for
    {{.Stats.Documents}} documents ({{HumanUnit .Stats.ContentBytes}})
    from {{.Stats.Repos}} repositories.
    </p>
  </div>

  <nav class="navbar navbar-default navbar-bottom">
    <div class="container">
      {{template "footerBoilerplate"}}
      <p class="navbar-text navbar-right">
      </p>
    </div>
  </nav>
`,
}

func init() {
	for k, v := range TemplateText {
		_, err := Top.New(k).Parse(v)
		if err != nil {
			log.Panicf("parse(%s): %v:", k, err)
		}
	}
}
