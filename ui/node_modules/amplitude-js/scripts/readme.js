var fs = require('fs');
var path = require('path');

// Update the README with the minified snippet.
var cwd = process.cwd();
var readmeFilename = path.join(cwd, "README.md");
var readme = fs.readFileSync(readmeFilename, 'utf-8');

var snippetFilename = path.join(cwd, "amplitude-snippet.min.js");
var snippet = fs.readFileSync(snippetFilename, 'utf-8');
var script =
'        <script type="text/javascript">\n' +
snippet.trim().replace(/^/gm, '          ') + '\n\n' +
'          amplitude.init("YOUR_API_KEY_HERE");\n' +
'        </script>';

var updated = readme.replace(/ +<script[\s\S]+?script>/, script);
fs.writeFileSync(readmeFilename, updated);

console.log('Updated README.md');
