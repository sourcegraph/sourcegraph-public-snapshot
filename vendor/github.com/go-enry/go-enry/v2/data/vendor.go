// Code generated by github.com/go-enry/go-enry/v2/internal/code-generator DO NOT EDIT.
// Extracted from github/linguist commit: 4ac734c15a96f9e16fd12330d0cb8de82274f700

package data

import "github.com/go-enry/go-enry/v2/regex"

var VendorMatchers = []regex.EnryRegexp{
	regex.MustCompile(`(^|/)cache/`),
	regex.MustCompile(`^[Dd]ependencies/`),
	regex.MustCompile(`(^|/)dist/`),
	regex.MustCompile(`^deps/`),
	regex.MustCompile(`(^|/)configure$`),
	regex.MustCompile(`(^|/)config\.guess$`),
	regex.MustCompile(`(^|/)config\.sub$`),
	regex.MustCompile(`(^|/)aclocal\.m4`),
	regex.MustCompile(`(^|/)libtool\.m4`),
	regex.MustCompile(`(^|/)ltoptions\.m4`),
	regex.MustCompile(`(^|/)ltsugar\.m4`),
	regex.MustCompile(`(^|/)ltversion\.m4`),
	regex.MustCompile(`(^|/)lt~obsolete\.m4`),
	regex.MustCompile(`(^|/)dotnet-install\.(ps1|sh)$`),
	regex.MustCompile(`(^|/)cpplint\.py`),
	regex.MustCompile(`(^|/)node_modules/`),
	regex.MustCompile(`(^|/)\.yarn/releases/`),
	regex.MustCompile(`(^|/)\.yarn/plugins/`),
	regex.MustCompile(`(^|/)\.yarn/sdks/`),
	regex.MustCompile(`(^|/)\.yarn/versions/`),
	regex.MustCompile(`(^|/)\.yarn/unplugged/`),
	regex.MustCompile(`(^|/)_esy$`),
	regex.MustCompile(`(^|/)bower_components/`),
	regex.MustCompile(`^rebar$`),
	regex.MustCompile(`(^|/)erlang\.mk`),
	regex.MustCompile(`(^|/)Godeps/_workspace/`),
	regex.MustCompile(`(^|/)testdata/`),
	regex.MustCompile(`(^|/)\.indent\.pro`),
	regex.MustCompile(`(\.|-)min\.(js|css)$`),
	regex.MustCompile(`([^\s]*)import\.(css|less|scss|styl)$`),
	regex.MustCompile(`(^|/)bootstrap([^/.]*)(\..*)?\.(js|css|less|scss|styl)$`),
	regex.MustCompile(`(^|/)custom\.bootstrap([^\s]*)(js|css|less|scss|styl)$`),
	regex.MustCompile(`(^|/)font-?awesome\.(css|less|scss|styl)$`),
	regex.MustCompile(`(^|/)font-?awesome/.*\.(css|less|scss|styl)$`),
	regex.MustCompile(`(^|/)foundation\.(css|less|scss|styl)$`),
	regex.MustCompile(`(^|/)normalize\.(css|less|scss|styl)$`),
	regex.MustCompile(`(^|/)skeleton\.(css|less|scss|styl)$`),
	regex.MustCompile(`(^|/)[Bb]ourbon/.*\.(css|less|scss|styl)$`),
	regex.MustCompile(`(^|/)animate\.(css|less|scss|styl)$`),
	regex.MustCompile(`(^|/)materialize\.(css|less|scss|styl|js)$`),
	regex.MustCompile(`(^|/)select2/.*\.(css|scss|js)$`),
	regex.MustCompile(`(^|/)bulma\.(css|sass|scss)$`),
	regex.MustCompile(`(3rd|[Tt]hird)[-_]?[Pp]arty/`),
	regex.MustCompile(`(^|/)vendors?/`),
	regex.MustCompile(`(^|/)[Ee]xtern(als?)?/`),
	regex.MustCompile(`(^|/)[Vv]+endor/`),
	regex.MustCompile(`^debian/`),
	regex.MustCompile(`(^|/)run\.n$`),
	regex.MustCompile(`(^|/)bootstrap-datepicker/`),
	regex.MustCompile(`(^|/)jquery([^.]*)\.js$`),
	regex.MustCompile(`(^|/)jquery\-\d\.\d+(\.\d+)?\.js$`),
	regex.MustCompile(`(^|/)jquery\-ui(\-\d\.\d+(\.\d+)?)?(\.\w+)?\.(js|css)$`),
	regex.MustCompile(`(^|/)jquery\.(ui|effects)\.([^.]*)\.(js|css)$`),
	regex.MustCompile(`(^|/)jquery\.fn\.gantt\.js`),
	regex.MustCompile(`(^|/)jquery\.fancybox\.(js|css)`),
	regex.MustCompile(`(^|/)fuelux\.js`),
	regex.MustCompile(`(^|/)jquery\.fileupload(-\w+)?\.js$`),
	regex.MustCompile(`(^|/)jquery\.dataTables\.js`),
	regex.MustCompile(`(^|/)bootbox\.js`),
	regex.MustCompile(`(^|/)pdf\.worker\.js`),
	regex.MustCompile(`(^|/)slick\.\w+.js$`),
	regex.MustCompile(`(^|/)Leaflet\.Coordinates-\d+\.\d+\.\d+\.src\.js$`),
	regex.MustCompile(`(^|/)leaflet\.draw-src\.js`),
	regex.MustCompile(`(^|/)leaflet\.draw\.css`),
	regex.MustCompile(`(^|/)Control\.FullScreen\.css`),
	regex.MustCompile(`(^|/)Control\.FullScreen\.js`),
	regex.MustCompile(`(^|/)leaflet\.spin\.js`),
	regex.MustCompile(`(^|/)wicket-leaflet\.js`),
	regex.MustCompile(`(^|/)\.sublime-project`),
	regex.MustCompile(`(^|/)\.sublime-workspace`),
	regex.MustCompile(`(^|/)\.vscode/`),
	regex.MustCompile(`(^|/)prototype(.*)\.js$`),
	regex.MustCompile(`(^|/)effects\.js$`),
	regex.MustCompile(`(^|/)controls\.js$`),
	regex.MustCompile(`(^|/)dragdrop\.js$`),
	regex.MustCompile(`(.*?)\.d\.ts$`),
	regex.MustCompile(`(^|/)mootools([^.]*)\d+\.\d+.\d+([^.]*)\.js$`),
	regex.MustCompile(`(^|/)dojo\.js$`),
	regex.MustCompile(`(^|/)MochiKit\.js$`),
	regex.MustCompile(`(^|/)yahoo-([^.]*)\.js$`),
	regex.MustCompile(`(^|/)yui([^.]*)\.js$`),
	regex.MustCompile(`(^|/)ckeditor\.js$`),
	regex.MustCompile(`(^|/)tiny_mce([^.]*)\.js$`),
	regex.MustCompile(`(^|/)tiny_mce/(langs|plugins|themes|utils)`),
	regex.MustCompile(`(^|/)ace-builds/`),
	regex.MustCompile(`(^|/)fontello(.*?)\.css$`),
	regex.MustCompile(`(^|/)MathJax/`),
	regex.MustCompile(`(^|/)Chart\.js$`),
	regex.MustCompile(`(^|/)[Cc]ode[Mm]irror/(\d+\.\d+/)?(lib|mode|theme|addon|keymap|demo)`),
	regex.MustCompile(`(^|/)shBrush([^.]*)\.js$`),
	regex.MustCompile(`(^|/)shCore\.js$`),
	regex.MustCompile(`(^|/)shLegacy\.js$`),
	regex.MustCompile(`(^|/)angular([^.]*)\.js$`),
	regex.MustCompile(`(^|\/)d3(\.v\d+)?([^.]*)\.js$`),
	regex.MustCompile(`(^|/)react(-[^.]*)?\.js$`),
	regex.MustCompile(`(^|/)flow-typed/.*\.js$`),
	regex.MustCompile(`(^|/)modernizr\-\d\.\d+(\.\d+)?\.js$`),
	regex.MustCompile(`(^|/)modernizr\.custom\.\d+\.js$`),
	regex.MustCompile(`(^|/)knockout-(\d+\.){3}(debug\.)?js$`),
	regex.MustCompile(`(^|/)docs?/_?(build|themes?|templates?|static)/`),
	regex.MustCompile(`(^|/)admin_media/`),
	regex.MustCompile(`(^|/)env/`),
	regex.MustCompile(`(^|/)fabfile\.py$`),
	regex.MustCompile(`(^|/)waf$`),
	regex.MustCompile(`(^|/)\.osx$`),
	regex.MustCompile(`\.xctemplate/`),
	regex.MustCompile(`\.imageset/`),
	regex.MustCompile(`(^|/)Carthage/`),
	regex.MustCompile(`(^|/)Sparkle/`),
	regex.MustCompile(`(^|/)Crashlytics\.framework/`),
	regex.MustCompile(`(^|/)Fabric\.framework/`),
	regex.MustCompile(`(^|/)BuddyBuildSDK\.framework/`),
	regex.MustCompile(`(^|/)Realm\.framework`),
	regex.MustCompile(`(^|/)RealmSwift\.framework`),
	regex.MustCompile(`(^|/)\.gitattributes$`),
	regex.MustCompile(`(^|/)\.gitignore$`),
	regex.MustCompile(`(^|/)\.gitmodules$`),
	regex.MustCompile(`(^|/)gradlew$`),
	regex.MustCompile(`(^|/)gradlew\.bat$`),
	regex.MustCompile(`(^|/)gradle/wrapper/`),
	regex.MustCompile(`(^|/)mvnw$`),
	regex.MustCompile(`(^|/)mvnw\.cmd$`),
	regex.MustCompile(`(^|/)\.mvn/wrapper/`),
	regex.MustCompile(`-vsdoc\.js$`),
	regex.MustCompile(`\.intellisense\.js$`),
	regex.MustCompile(`(^|/)jquery([^.]*)\.validate(\.unobtrusive)?\.js$`),
	regex.MustCompile(`(^|/)jquery([^.]*)\.unobtrusive\-ajax\.js$`),
	regex.MustCompile(`(^|/)[Mm]icrosoft([Mm]vc)?([Aa]jax|[Vv]alidation)(\.debug)?\.js$`),
	regex.MustCompile(`(^|/)[Pp]ackages\/.+\.\d+\/`),
	regex.MustCompile(`(^|/)extjs/.*?\.js$`),
	regex.MustCompile(`(^|/)extjs/.*?\.xml$`),
	regex.MustCompile(`(^|/)extjs/.*?\.txt$`),
	regex.MustCompile(`(^|/)extjs/.*?\.html$`),
	regex.MustCompile(`(^|/)extjs/.*?\.properties$`),
	regex.MustCompile(`(^|/)extjs/\.sencha/`),
	regex.MustCompile(`(^|/)extjs/docs/`),
	regex.MustCompile(`(^|/)extjs/builds/`),
	regex.MustCompile(`(^|/)extjs/cmd/`),
	regex.MustCompile(`(^|/)extjs/examples/`),
	regex.MustCompile(`(^|/)extjs/locale/`),
	regex.MustCompile(`(^|/)extjs/packages/`),
	regex.MustCompile(`(^|/)extjs/plugins/`),
	regex.MustCompile(`(^|/)extjs/resources/`),
	regex.MustCompile(`(^|/)extjs/src/`),
	regex.MustCompile(`(^|/)extjs/welcome/`),
	regex.MustCompile(`(^|/)html5shiv\.js$`),
	regex.MustCompile(`(^|/)[Tt]ests?/fixtures/`),
	regex.MustCompile(`(^|/)[Ss]pecs?/fixtures/`),
	regex.MustCompile(`(^|/)cordova([^.]*)\.js$`),
	regex.MustCompile(`(^|/)cordova\-\d\.\d(\.\d)?\.js$`),
	regex.MustCompile(`(^|/)foundation(\..*)?\.js$`),
	regex.MustCompile(`(^|/)Vagrantfile$`),
	regex.MustCompile(`(^|/)\.[Dd][Ss]_[Ss]tore$`),
	regex.MustCompile(`(^|/)inst/extdata/`),
	regex.MustCompile(`(^|/)octicons\.css`),
	regex.MustCompile(`(^|/)sprockets-octicons\.scss`),
	regex.MustCompile(`(^|/)activator$`),
	regex.MustCompile(`(^|/)activator\.bat$`),
	regex.MustCompile(`(^|/)proguard\.pro$`),
	regex.MustCompile(`(^|/)proguard-rules\.pro$`),
	regex.MustCompile(`(^|/)puphpet/`),
	regex.MustCompile(`(^|/)\.google_apis/`),
	regex.MustCompile(`(^|/)Jenkinsfile$`),
	regex.MustCompile(`(^|/)\.gitpod\.Dockerfile$`),
	regex.MustCompile(`(^|/)\.github/`),
	regex.MustCompile(`(^|/)\.obsidian/`),
	regex.MustCompile(`(^|/)\.teamcity/`),
}

// FastVendorMatcher is equivalent to matching any of the VendorMatchers.
var FastVendorMatcher = regex.MustCompile(`(?:^(?:(?:[Dd]ependencies/)|(?:debian/)|(?:deps/)|(?:rebar$)))|(?:(?:^|/)(?:(?:BuddyBuildSDK\.framework/)|(?:Carthage/)|(?:Chart\.js$)|(?:Control\.FullScreen\.css)|(?:Control\.FullScreen\.js)|(?:Crashlytics\.framework/)|(?:Fabric\.framework/)|(?:Godeps/_workspace/)|(?:Jenkinsfile$)|(?:Leaflet\.Coordinates-\d+\.\d+\.\d+\.src\.js$)|(?:MathJax/)|(?:MochiKit\.js$)|(?:RealmSwift\.framework)|(?:Realm\.framework)|(?:Sparkle/)|(?:Vagrantfile$)|(?:[Bb]ourbon/.*\.(css|less|scss|styl)$)|(?:[Cc]ode[Mm]irror/(\d+\.\d+/)?(lib|mode|theme|addon|keymap|demo))|(?:[Ee]xtern(als?)?/)|(?:[Mm]icrosoft([Mm]vc)?([Aa]jax|[Vv]alidation)(\.debug)?\.js$)|(?:[Pp]ackages\/.+\.\d+\/)|(?:[Ss]pecs?/fixtures/)|(?:[Tt]ests?/fixtures/)|(?:[Vv]+endor/)|(?:\.[Dd][Ss]_[Ss]tore$)|(?:\.gitattributes$)|(?:\.github/)|(?:\.gitignore$)|(?:\.gitmodules$)|(?:\.gitpod\.Dockerfile$)|(?:\.google_apis/)|(?:\.indent\.pro)|(?:\.mvn/wrapper/)|(?:\.obsidian/)|(?:\.osx$)|(?:\.sublime-project)|(?:\.sublime-workspace)|(?:\.teamcity/)|(?:\.vscode/)|(?:\.yarn/plugins/)|(?:\.yarn/releases/)|(?:\.yarn/sdks/)|(?:\.yarn/unplugged/)|(?:\.yarn/versions/)|(?:_esy$)|(?:ace-builds/)|(?:aclocal\.m4)|(?:activator$)|(?:activator\.bat$)|(?:admin_media/)|(?:angular([^.]*)\.js$)|(?:animate\.(css|less|scss|styl)$)|(?:bootbox\.js)|(?:bootstrap([^/.]*)(\..*)?\.(js|css|less|scss|styl)$)|(?:bootstrap-datepicker/)|(?:bower_components/)|(?:bulma\.(css|sass|scss)$)|(?:cache/)|(?:ckeditor\.js$)|(?:config\.guess$)|(?:config\.sub$)|(?:configure$)|(?:controls\.js$)|(?:cordova([^.]*)\.js$)|(?:cordova\-\d\.\d(\.\d)?\.js$)|(?:cpplint\.py)|(?:custom\.bootstrap([^\s]*)(js|css|less|scss|styl)$)|(?:dist/)|(?:docs?/_?(build|themes?|templates?|static)/)|(?:dojo\.js$)|(?:dotnet-install\.(ps1|sh)$)|(?:dragdrop\.js$)|(?:effects\.js$)|(?:env/)|(?:erlang\.mk)|(?:extjs/.*?\.html$)|(?:extjs/.*?\.js$)|(?:extjs/.*?\.properties$)|(?:extjs/.*?\.txt$)|(?:extjs/.*?\.xml$)|(?:extjs/\.sencha/)|(?:extjs/builds/)|(?:extjs/cmd/)|(?:extjs/docs/)|(?:extjs/examples/)|(?:extjs/locale/)|(?:extjs/packages/)|(?:extjs/plugins/)|(?:extjs/resources/)|(?:extjs/src/)|(?:extjs/welcome/)|(?:fabfile\.py$)|(?:flow-typed/.*\.js$)|(?:font-?awesome/.*\.(css|less|scss|styl)$)|(?:font-?awesome\.(css|less|scss|styl)$)|(?:fontello(.*?)\.css$)|(?:foundation(\..*)?\.js$)|(?:foundation\.(css|less|scss|styl)$)|(?:fuelux\.js)|(?:gradle/wrapper/)|(?:gradlew$)|(?:gradlew\.bat$)|(?:html5shiv\.js$)|(?:inst/extdata/)|(?:jquery([^.]*)\.js$)|(?:jquery([^.]*)\.unobtrusive\-ajax\.js$)|(?:jquery([^.]*)\.validate(\.unobtrusive)?\.js$)|(?:jquery\-\d\.\d+(\.\d+)?\.js$)|(?:jquery\-ui(\-\d\.\d+(\.\d+)?)?(\.\w+)?\.(js|css)$)|(?:jquery\.(ui|effects)\.([^.]*)\.(js|css)$)|(?:jquery\.dataTables\.js)|(?:jquery\.fancybox\.(js|css))|(?:jquery\.fileupload(-\w+)?\.js$)|(?:jquery\.fn\.gantt\.js)|(?:knockout-(\d+\.){3}(debug\.)?js$)|(?:leaflet\.draw-src\.js)|(?:leaflet\.draw\.css)|(?:leaflet\.spin\.js)|(?:libtool\.m4)|(?:ltoptions\.m4)|(?:ltsugar\.m4)|(?:ltversion\.m4)|(?:lt~obsolete\.m4)|(?:materialize\.(css|less|scss|styl|js)$)|(?:modernizr\-\d\.\d+(\.\d+)?\.js$)|(?:modernizr\.custom\.\d+\.js$)|(?:mootools([^.]*)\d+\.\d+.\d+([^.]*)\.js$)|(?:mvnw$)|(?:mvnw\.cmd$)|(?:node_modules/)|(?:normalize\.(css|less|scss|styl)$)|(?:octicons\.css)|(?:pdf\.worker\.js)|(?:proguard-rules\.pro$)|(?:proguard\.pro$)|(?:prototype(.*)\.js$)|(?:puphpet/)|(?:react(-[^.]*)?\.js$)|(?:run\.n$)|(?:select2/.*\.(css|scss|js)$)|(?:shBrush([^.]*)\.js$)|(?:shCore\.js$)|(?:shLegacy\.js$)|(?:skeleton\.(css|less|scss|styl)$)|(?:slick\.\w+.js$)|(?:sprockets-octicons\.scss)|(?:testdata/)|(?:tiny_mce([^.]*)\.js$)|(?:tiny_mce/(langs|plugins|themes|utils))|(?:vendors?/)|(?:waf$)|(?:wicket-leaflet\.js)|(?:yahoo-([^.]*)\.js$)|(?:yui([^.]*)\.js$)))|(?:(.*?)\.d\.ts$)|(?:(3rd|[Tt]hird)[-_]?[Pp]arty/)|(?:([^\s]*)import\.(css|less|scss|styl)$)|(?:(\.|-)min\.(js|css)$)|(?:(^|\/)d3(\.v\d+)?([^.]*)\.js$)|(?:-vsdoc\.js$)|(?:\.imageset/)|(?:\.intellisense\.js$)|(?:\.xctemplate/)`)
