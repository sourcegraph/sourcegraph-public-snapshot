{{define "ROOT"}}
document.write('<link rel="stylesheet" href="{{.StylesheetURL}}">');
document.write('<script type="text/javascript" src="{{.ScriptURL}}"></script>');
document.write({{.EscapedHTML}});
{{end}}
