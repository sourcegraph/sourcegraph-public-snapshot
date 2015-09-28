package config

import (
	"encoding/json"
	"log"
	"os"

	"strings"

	"sourcegraph.com/sourcegraph/srclib/unit"
)

func init() {
	additionalOverrides := os.Getenv("SRCLIB_ADDITIONAL_OVERRIDES")
	if additionalOverrides != "" {
		var o map[string]*Repository
		if err := json.Unmarshal([]byte(additionalOverrides), &o); err != nil {
			// HACK: In upstart, escaped double quotes include the backslash
			// in the string. To get around this, we encode json using
			// single quotes instead of double quotes, and make the switch
			// here.
			additionalOverrides = strings.Replace(additionalOverrides, `'`, `"`, -1)
			if err := json.Unmarshal([]byte(additionalOverrides), &o); err != nil {
				log.Fatalf("config/overrides.go init(): %s", err)
			}
		}
		for k, v := range o {
			Overrides[k] = v
		}
	}
}

var Overrides = map[string]*Repository{
	"sourcegraph.com/sourcegraph/sourcegraph": {
		URI: "sourcegraph.com/sourcegraph/sourcegraph",
		Tree: Tree{
			SkipDirs: []string{"app/node_modules", "app/bower_components"},
			SkipUnits: []struct{ Name, Type string }{
				{Name: ".", Type: "ruby"},
				{Name: "app", Type: "CommonJSPackage"},
			},
		},
	},
	"code.google.com/p/rsc": {
		URI: "code.google.com/p/rsc",
		Tree: Tree{
			SkipDirs: []string{"cmd/numbers", "cc"},
		},
	},
	"github.com/emicklei/go-restful": {
		URI: "github.com/emicklei/go-restful",
		Tree: Tree{
			SkipDirs: []string{"examples"},
		},
	},
	"github.com/golang/go": {
		URI: "github.com/golang/go",
		Tree: Tree{
			Config:   map[string]interface{}{"GOROOT": "."},
			SkipDirs: []string{"test", "misc", "doc", "lib", "include"},
			PreConfigCommands: []string{`
if [ -d /home/srclib ]; then
curl -L 'https://storage.googleapis.com/golang/go1.4.1.linux-amd64.tar.gz' > /tmp/go1.4.tgz &&
mkdir /home/srclib/go1.4 &&
tar -xf /tmp/go1.4.tgz -C /home/srclib/go1.4 --strip-components=1
fi &&
cd src
./make.bash
`},
		},
	},
	"code.google.com/p/go": {
		URI: "code.google.com/p/go",
		Tree: Tree{
			Config:            map[string]interface{}{"GOROOT": "."},
			SkipDirs:          []string{"test", "misc", "doc", "lib", "include"},
			PreConfigCommands: []string{"echo devel > VERSION && cd src && ./make.bash"},
		},
	},
	"github.com/joyent/node": {
		URI: "github.com/joyent/node",
		Tree: Tree{
			SkipDirs: []string{"tools", "deps", "test", "src"},
			SourceUnits: []*unit.SourceUnit{
				{
					Name:  "node",
					Type:  "CommonJSPackage",
					Dir:   ".",
					Files: []string{"lib/*.js"},
					Config: map[string]interface{}{
						"jsg": map[string]interface{}{
							"plugins": map[string]interface{}{
								"node": map[string]string{"coreModulesDir": "lib/"},
								"$(JSG_DIR)/node_modules/tern-node-api-doc/node-api-doc": map[string]string{
									"apiDocDir":      "doc/api/",
									"apiSrcDir":      "lib/",
									"generateJSPath": "tools/doc/generate.js",
								},
							},
						},
					},
				},
			},
		},
	},
	"github.com/ruby/ruby": {
		URI: "github.com/ruby/ruby",
		Tree: Tree{
			SkipDirs: []string{"test", "ext"},
			SourceUnits: []*unit.SourceUnit{
				{
					Name:   ".",
					Type:   "ruby",
					Dir:    ".",
					Config: map[string]interface{}{"noCachedStdlibYardoc": false},
					Files: []string{
						"*.c",
						"lib/*.rb",
						"lib/cgi/*.rb",
						"lib/cgi/session/*.rb",
						"lib/drb/*.rb",
						"lib/irb/*.rb",
						"lib/irb/cmd/*.rb",
						"lib/irb/ext/*.rb",
						"lib/irb/lc/*.rb",
						"lib/irb/lc/ja/*.rb",
						"lib/matrix/*.rb",
						"lib/minitest/*.rb",
						"lib/net/*.rb",
						"lib/net/http/*.rb",
						"lib/optparse/*.rb",
						"lib/racc/*.rb",
						"lib/rake/*.rb",
						"lib/rake/contrib/*.rb",
						"lib/rake/ext/*.rb",
						"lib/rake/loaders/*.rb",
						"lib/rbconfig/*.rb",
						"lib/rdoc/*.rb",
						"lib/rdoc/context/*.rb",
						"lib/rdoc/generator/*.rb",
						"lib/rdoc/markdown/*.rb",
						"lib/rdoc/markup/*.rb",
						"lib/rdoc/parser/*.rb",
						"lib/rdoc/rd/*.rb",
						"lib/rdoc/ri/*.rb",
						"lib/rdoc/stats/*.rb",
						"lib/rexml/*.rb",
						"lib/rexml/dtd/*.rb",
						"lib/rexml/formatters/*.rb",
						"lib/rexml/light/*.rb",
						"lib/rexml/parsers/*.rb",
						"lib/rexml/validation/*.rb",
						"lib/rinda/*.rb",
						"lib/rss/*.rb",
						"lib/rss/content/*.rb",
						"lib/rss/dublincore/*.rb",
						"lib/rss/maker/*.rb",
						"lib/rubygems/*.rb",
						"lib/rubygems/commands/*.rb",
						"lib/rubygems/core_ext/*.rb",
						"lib/rubygems/ext/*.rb",
						"lib/rubygems/package/*.rb",
						"lib/rubygems/package/tar_reader/*.rb",
						"lib/rubygems/request_set/*.rb",
						"lib/rubygems/resolver/*.rb",
						"lib/rubygems/security/*.rb",
						"lib/rubygems/source/*.rb",
						"lib/rubygems/util/*.rb",
						"lib/shell/*.rb",
						"lib/test/*.rb",
						"lib/test/unit/*.rb",
						"lib/uri/*.rb",
						"lib/webrick/*.rb",
						"lib/webrick/httpauth/*.rb",
						"lib/webrick/httpservlet/*.rb",
						"lib/xmlrpc/*.rb",
						"lib/yaml/*.rb",
						// TODO(sqs): should probably add ext/
						// TODO(sqs): it's annoying that filematch.Glob doesn't support
						// '**', so we have to list out each dir here (not just lib/**/*.rb)
					},
				},
			},
		},
	},
}
