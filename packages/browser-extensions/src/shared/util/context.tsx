import * as path from 'path'
import * as runtime from '../../browser/runtime'
import storage from '../../browser/storage'
import { isPhabricator } from '../context'
import { EventLogger } from '../tracking/EventLogger'

export const DEFAULT_SOURCEGRAPH_URL = 'https://sourcegraph.com'

export let eventLogger = new EventLogger()

export let sourcegraphUrl =
    window.localStorage.getItem('SOURCEGRAPH_URL') || window.SOURCEGRAPH_URL || DEFAULT_SOURCEGRAPH_URL

export let renderMermaidGraphsEnabled = false

export let inlineSymbolSearchEnabled = false

export let useExtensions = false

interface UrlCache {
    [key: string]: string
}

export const repoUrlCache: UrlCache = {}

if (window.SG_ENV === 'EXTENSION') {
    storage.getSync(items => {
        sourcegraphUrl = items.sourcegraphURL

        renderMermaidGraphsEnabled = items.renderMermaidGraphsEnabled

        inlineSymbolSearchEnabled = items.inlineSymbolSearchEnabled
        useExtensions = items.useExtensions
    })
}

export function setSourcegraphUrl(url: string): void {
    sourcegraphUrl = url
}

export function isBrowserExtension(): boolean {
    return window.SOURCEGRAPH_PHABRICATOR_EXTENSION || false
}

export function isSourcegraphDotCom(url: string = sourcegraphUrl): boolean {
    return url === DEFAULT_SOURCEGRAPH_URL
}

export function checkIsOnlySourcegraphDotCom(handler: (res: boolean) => void): void {
    if (window.SG_ENV === 'EXTENSION') {
        storage.getSync(items => handler(isSourcegraphDotCom(items.sourcegraphURL)))
    } else {
        handler(false)
    }
}

export function setRenderMermaidGraphsEnabled(enabled: boolean): void {
    renderMermaidGraphsEnabled = enabled
}

export function setInlineSymbolSearchEnabled(enabled: boolean): void {
    inlineSymbolSearchEnabled = enabled
}

export function setUseExtensions(value: boolean): void {
    useExtensions = value
}

/**
 * modeToHighlightJsName gets the highlight.js name of the language given a
 * mode.
 */
export function modeToHighlightJsName(mode: string): string {
    switch (mode) {
        case 'html':
            return 'xml'
        default:
            return mode
    }
}

/**
 * getModeFromPath returns the LSP mode for the provided file path.
 */
export function getModeFromPath(filePath: string): string | undefined {
    const fileName = path.basename(filePath)
    const ext = getPathExtension(filePath)

    return getModeFromExactFilename(fileName) || getModeFromExtension(ext)
}

/**
 * getModeFromExactFilename returns the LSP mode for the
 * provided file name (e.g. "dockerfile")
 *
 * Cherry picked from https://github.com/github/linguist/blob/master/lib/linguist/languages.yml
 */
function getModeFromExactFilename(fileName: string): string | undefined {
    switch (fileName.toLowerCase()) {
        case 'dockerfile':
            return 'dockerfile'

        default:
            return undefined
    }
}

/**
 * getModeFromExtension returns the LSP mode for the
 * provided file extension (e.g. "jsx")
 *
 * Cherry picked from https://github.com/isagalaev/highlight.js/tree/master/src/languages
 * and https://github.com/github/linguist/blob/master/lib/linguist/languages.yml.
 */
function getModeFromExtension(ext: string): string | undefined {
    switch (ext.toLowerCase()) {
        // Ada
        case 'adb':
        case 'ada':
        case 'ads':
            return 'ada'

        // Actionscript
        case 'as':
            return 'actionscript'

        // Apache
        case 'apacheconf':
            return 'apache'

        // Applescript
        case 'applescript':
        case 'scpt':
            return 'applescript'

        // Bash
        case 'sh':
        case 'bash':
        case 'zsh':
            return 'bash'

        // Clojure
        case 'clj':
        case 'cljs':
        case 'cljx':
            return 'clojure'

        // CSS
        case 'css':
            return 'css'

        // CMake
        case 'cmake':
        case 'cmake.in':
        case 'in': // TODO(john): hack b/c we don't properly parse extensions w/ '.' in them
            return 'cmake'

        // Coffeescript
        case 'coffee':
        case 'cake':
        case 'cson':
        case 'cjsx':
        case 'iced':
            return 'coffescript'

        // C#
        case 'cs':
        case 'csx':
            return 'cs'

        // C++
        case 'c':
        case 'cc':
        case 'cpp':
        case 'c++':
        case 'h++':
        case 'hh':
        case 'h':
            return 'cpp'

        // Dart
        case 'dart':
            return 'dart'

        // Diff
        case 'diff':
        case 'patch':
            return 'diff'

        // Django
        case 'jinja':
            return 'django'

        // DOS
        case 'bat':
        case 'cmd':
            return 'dos'

        // Elixir
        case 'ex':
        case 'exs':
            return 'elixir'

        // Elm
        case 'elm':
            return 'elm'

        // Erlang
        case 'erl':
            return 'erlang'

        // Fortran
        case 'f':
        case 'for':
        case 'frt':
        case 'fr':
        case 'fs':
        case 'forth':
        case '4th':
        case 'fth':
            return 'fortran'

        // F#
        case 'fs':
            return 'fsharp'

        // Go
        case 'go':
            return 'go'

        // HAML
        case 'haml':
            return 'haml'

        // Handlebars
        case 'hbs':
        case 'handlebars':
            return 'handlebars'

        // Haskell
        case 'hs':
        case 'hsc':
            return 'haskell'

        // HTML
        case 'htm':
        case 'html':
        case 'xhtml':
            return 'html'

        // INI
        case 'ini':
        case 'cfg':
        case 'prefs':
        case 'pro':
        case 'properties':
            return 'ini'

        // Java
        case 'java':
            return 'java'

        // JavaScript
        case 'js':
        case 'jsx':
        case 'es':
        case 'es6':
        case 'jss':
        case 'jsm':
        case 'mjs':
            return 'javascript'

        // JSON
        case 'json':
        case 'sublime_metrics':
        case 'sublime_session':
        case 'sublime-keymap':
        case 'sublime-mousemap':
        case 'sublime-project':
        case 'sublime-settings':
        case 'sublime-workspace':
            return 'json'

        // Julia
        case 'jl':
            return 'julia'

        // Kotlin
        case 'kt':
        case 'ktm':
        case 'kts':
            return 'kotlin'

        // Less
        case 'less':
            return 'less'

        // Lisp
        case 'lisp':
        case 'asd':
        case 'cl':
        case 'lsp':
        case 'l':
        case 'ny':
        case 'podsl':
        case 'sexp':
        case 'el':
            return 'lisp'

        // Lua
        case 'lua':
        case 'fcgi':
        case 'nse':
        case 'pd_lua':
        case 'rbxs':
        case 'wlua':
            return 'lua'

        // Makefile
        case 'mk':
        case 'mak':
            return 'makefile'

        // Markdown
        case 'md':
        case 'mkdown':
        case 'mkd':
            return 'markdown'

        // nginx
        case 'nginxconf':
            return 'nginx'

        // Objective-C
        case 'm':
        case 'mm':
            return 'objectivec'

        // OCaml
        case 'ml':
        case 'eliom':
        case 'eliomi':
        case 'ml4':
        case 'mli':
        case 'mll':
        case 'mly':
            return 'ocaml'

        // Perl
        case 'pl':
        case 'al':
        case 'cgi':
        case 'fcgi':
        case 'perl':
        case 'ph':
        case 'plx':
        case 'pm':
        case 'pod':
        case 'psgi':
        case 't':
            return 'perl'

        // PHP
        case 'php':
        case 'phtml':
        case 'php3':
        case 'php4':
        case 'php5':
        case 'php6':
        case 'php7':
        case 'phps':
            return 'php'

        // Powershell
        case 'ps1':
        case 'psd1':
        case 'psm1':
            return 'powershell'

        // Proto
        case 'proto':
            return 'protobuf'

        // Python
        case 'py':
        case 'pyc':
        case 'pyd':
        case 'pyo':
        case 'pyw':
        case 'pyz':
            return 'python'

        // R
        case 'r':
        case 'rd':
        case 'rsx':
            return 'r'

        // Ruby
        case 'rb':
        case 'builder':
        case 'eye':
        case 'fcgi':
        case 'gemspec':
        case 'god':
        case 'jbuilder':
        case 'mspec':
        case 'pluginspec':
        case 'podspec':
        case 'rabl':
        case 'rake':
        case 'rbuild':
        case 'rbw':
        case 'rbx':
        case 'ru':
        case 'ruby':
        case 'spec':
        case 'thor':
        case 'watchr':
            return 'ruby'

        // Rust
        case 'rs':
        case 'rs.in':
            return 'rust'

        // SASS
        case 'sass':
        case 'scss':
            return 'scss'

        // Scala
        case 'sbt':
        case 'sc':
        case 'scala':
            return 'scala'

        // Scheme
        case 'scm':
        case 'sch':
        case 'sls':
        case 'sps':
        case 'ss':
            return 'scheme'

        // Smalltalk
        case 'st':
            return 'smalltalk'

        // SQL
        case 'sql':
            return 'sql'

        // Stylus
        case 'styl':
            return 'stylus'

        // Swift
        case 'swift':
            return 'swift'

        // Thrift
        case 'thrift':
            return 'thrift'

        // TypeScript
        case 'ts':
        case 'tsx':
            return 'typescript'

        // Twig
        case 'twig':
            return 'twig'

        // Visual Basic
        case 'vb':
            return 'vbnet'
        case 'vbs':
            return 'vbscrip'

        // Verilog
        case 'v':
        case 'veo':
            return 'verilog'

        // VIM
        case 'vim':
            return 'vim'

        // XML
        case 'xml':
        case 'adml':
        case 'admx':
        case 'ant':
        case 'axml':
        case 'builds':
        case 'ccxml':
        case 'clixml':
        case 'cproject':
        case 'csl':
        case 'csproj':
        case 'ct':
        case 'dita':
        case 'ditamap':
        case 'ditaval':
        case 'dll.config':
        case 'dotsettings':
        case 'filters':
        case 'fsproj':
        case 'fxml':
        case 'glade':
        case 'gml':
        case 'grxml':
        case 'iml':
        case 'ivy':
        case 'jelly':
        case 'jsproj':
        case 'kml':
        case 'launch':
        case 'mdpolicy':
        case 'mjml':
        case 'mod':
        case 'mxml':
        case 'nproj':
        case 'nuspec':
        case 'odd':
        case 'osm':
        case 'pkgproj':
        case 'plist':
        case 'pluginspec':
        case 'props':
        case 'ps1xml':
        case 'psc1':
        case 'pt':
        case 'rdf':
        case 'resx':
        case 'rss':
        case 'sch':
        case 'scxml':
        case 'sfproj':
        case 'srdf':
        case 'storyboard':
        case 'stTheme':
        case 'sublime-snippet':
        case 'targets':
        case 'tmCommand':
        case 'tml':
        case 'tmLanguage':
        case 'tmPreferences':
        case 'tmSnippet':
        case 'tmTheme':
        case 'ts':
        case 'tsx':
        case 'ui':
        case 'urdf':
        case 'ux':
        case 'vbproj':
        case 'vcxproj':
        case 'vsixmanifest':
        case 'vssettings':
        case 'vstemplate':
        case 'vxml':
        case 'wixproj':
        case 'wsdl':
        case 'wsf':
        case 'wxi':
        case 'wxl':
        case 'wxs':
        case 'x3d':
        case 'xacro':
        case 'xaml':
        case 'xib':
        case 'xlf':
        case 'xliff':
        case 'xmi':
        case 'xml.dist':
        case 'xproj':
        case 'xsd':
        case 'xspec':
        case 'xul':
        case 'zcml':
            return 'xml'

        // YAML
        case 'yml':
        case 'yaml':
            return 'yaml'

        default:
            return undefined
    }
}

export function getPathExtension(path: string): string {
    const pathSplit = path.split('.')
    if (pathSplit.length === 1) {
        return ''
    }
    if (pathSplit.length === 2 && pathSplit[0] === '') {
        return '' // e.g. .gitignore
    }
    return pathSplit[pathSplit.length - 1].toLowerCase()
}

export function getPlatformName():
    | 'phabricator-integration'
    | 'safari-extension'
    | 'firefox-extension'
    | 'chrome-extension' {
    if (window.SOURCEGRAPH_PHABRICATOR_EXTENSION) {
        return 'phabricator-integration'
    }

    if (typeof window.safari !== 'undefined') {
        return 'safari-extension'
    }

    return isFirefoxExtension() ? 'firefox-extension' : 'chrome-extension'
}

export function getExtensionVersionSync(): string {
    return runtime.getExtensionVersionSync()
}

export function getExtensionVersion(): Promise<string> {
    return runtime.getExtensionVersion()
}

export function isFirefoxExtension(): boolean {
    return window.navigator.userAgent.indexOf('Firefox') !== -1
}

export function isE2ETest(): boolean {
    return process.env.NODE_ENV === 'test'
}

/**
 * This method created a unique username based on the platform and domain the user is visiting.
 * Examples: sg_dev_phabricator:matt
 */
export function getDomainUsername(domain: string, username: string): string {
    return `${domain}:${username}`
}

/**
 * Check the DOM to see if we can determine if a repository is private or public.
 */
export function isPrivateRepository(): boolean {
    if (isPhabricator) {
        return true
    }
    const header = document.querySelector('.repohead-details-container')
    if (!header) {
        return false
    }
    return !!header.querySelector('.private')
}

export function canFetchForURL(url: string): boolean {
    if (url === DEFAULT_SOURCEGRAPH_URL && isPrivateRepository()) {
        return false
    }
    return true
}
