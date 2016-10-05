// This file contains all of the languages ("modes" in vscode
// terminology) that we support in the UI. To add a new mode
// (language), just add another registerLanguage call below.
//
// Adapted from
// https://github.com/Microsoft/monaco-languages/blob/master/src/monaco.contribution.ts.
// We can't use that file directly because it assumes a "monaco"
// global object.

import {TPromise} from "vs/base/common/winjs.base";
import {onLanguage, register, setLanguageConfiguration, setMonarchTokensProvider} from "vs/editor/browser/standalone/standaloneLanguages";
import {LanguageConfiguration} from "vs/editor/common/modes/languageConfigurationRegistry";
import {IMonarchLanguage} from "vs/editor/common/modes/monarch/monarchTypes";
import {ILanguageExtensionPoint} from "vs/editor/common/services/modeService";

interface ILang extends ILanguageExtensionPoint {
	module: string;
}

interface ILangImpl {
	conf: LanguageConfiguration;
	language: IMonarchLanguage;
}

let languageDefinitions: {[languageId: string]: ILang} = {};

function _loadLanguage(languageId: string): Thenable<void> {
	let module = languageDefinitions[languageId].module;
	module = module.slice(2); // remove leading "./" so we can hardcode it in the require call for webpack context
	return new TPromise<void>((c, e, p) => {
		const mod = require(`monaco-languages/release/src/${module}.js`) as ILangImpl;
		setMonarchTokensProvider(languageId, mod.language);
		setLanguageConfiguration(languageId, mod.conf);
		c(void 0);
	});
}

let languagePromises: {[languageId: string]: Thenable<void>} = {};

export function loadLanguage(languageId: string): Thenable<void> {
	if (!languagePromises[languageId]) {
		languagePromises[languageId] = _loadLanguage(languageId);
	}
	return languagePromises[languageId];
}

function registerLanguage(def: ILang): void {
	let languageId = def.id;

	languageDefinitions[languageId] = def;
	register(def);
	onLanguage(languageId, () => {
		loadLanguage(languageId);
	});
}

registerLanguage({
	id: "bat",
	extensions: [ ".bat", ".cmd"],
	aliases: [ "Batch", "bat" ],
	module: "./bat",
});
registerLanguage({
	id: "coffeescript",
	extensions: [ ".coffee" ],
	aliases: [ "CoffeeScript", "coffeescript", "coffee" ],
	mimetypes: ["text/x-coffeescript", "text/coffeescript"],
	module: "./coffee",
});
registerLanguage({
	id: "c",
	extensions: [ ".c", ".h" ],
	aliases: [ "C", "c" ],
	module: "./cpp",
});
registerLanguage({
	id: "cpp",
	extensions: [ ".cpp", ".cc", ".cxx", ".hpp", ".hh", ".hxx" ],
	aliases: [ "C++", "Cpp", "cpp"],
	module: "./cpp",
});
registerLanguage({
	id: "csharp",
	extensions: [ ".cs", ".csx" ],
	aliases: [ "C#", "csharp" ],
	module: "./csharp",
});
registerLanguage({
	id: "dockerfile",
	extensions: [ ".dockerfile" ],
	filenames: [ "Dockerfile" ],
	aliases: [ "Dockerfile" ],
	module: "./dockerfile",
});
registerLanguage({
	id: "fsharp",
	extensions: [ ".fs", ".fsi", ".ml", ".mli", ".fsx", ".fsscript" ],
	aliases: [ "F#", "FSharp", "fsharp" ],
	module: "./fsharp",
});
registerLanguage({
	id: "go",
	extensions: [ ".go" ],
	aliases: [ "Go" ],
	module: "./go",
});
registerLanguage({
	id: "handlebars",
	extensions: [".handlebars", ".hbs"],
	aliases: ["Handlebars", "handlebars"],
	mimetypes: ["text/x-handlebars-template"],
	module: "./handlebars",
});
registerLanguage({
	id: "html",
	extensions: [".html", ".htm", ".shtml", ".xhtml", ".mdoc", ".jsp", ".asp", ".aspx", ".jshtm"],
	aliases: ["HTML", "htm", "html", "xhtml"],
	mimetypes: ["text/html", "text/x-jshtm", "text/template", "text/ng-template"],
	module: "./html",
});
registerLanguage({
	id: "ini",
	extensions: [ ".ini", ".properties", ".gitconfig" ],
	filenames: ["config", ".gitattributes", ".gitconfig", ".editorconfig"],
	aliases: [ "Ini", "ini" ],
	module: "./ini",
});
registerLanguage({
	id: "jade",
	extensions: [ ".jade", ".pug" ],
	aliases: [ "Jade", "jade" ],
	module: "./jade",
});
registerLanguage({
	id: "java",
	extensions: [ ".java", ".jav" ],
	aliases: [ "Java", "java" ],
	mimetypes: ["text/x-java-source", "text/x-java"],
	module: "./java",
});
registerLanguage({
	id: "lua",
	extensions: [ ".lua" ],
	aliases: [ "Lua", "lua" ],
	module: "./lua",
});
registerLanguage({
	id: "markdown",
	extensions: [".md", ".markdown", ".mdown", ".mkdn", ".mkd", ".mdwn", ".mdtxt", ".mdtext"],
	aliases: ["Markdown", "markdown"],
	module: "./markdown",
});
registerLanguage({
	id: "objective-c",
	extensions: [ ".m" ],
	aliases: [ "Objective-C"],
	module: "./objective-c",
});
registerLanguage({
	id: "postiats",
	extensions: [ ".dats", ".sats", ".hats" ],
	aliases: [ "ATS", "ATS/Postiats" ],
	module: "./postiats",
});
registerLanguage({
	id: "php",
	extensions: [".php", ".php4", ".php5", ".phtml", ".ctp"],
	aliases: ["PHP", "php"],
	mimetypes: ["application/x-php"],
	module: "./php",
});
registerLanguage({
	id: "powershell",
	extensions: [ ".ps1", ".psm1", ".psd1" ],
	aliases: [ "PowerShell", "powershell", "ps", "ps1" ],
	module: "./powershell",
});
registerLanguage({
	id: "python",
	extensions: [ ".py", ".rpy", ".pyw", ".cpy", ".gyp", ".gypi" ],
	aliases: [ "Python", "py" ],
	firstLine: "^#!/.*\\bpython[0-9.-]*\\b",
	module: "./python",
});
registerLanguage({
	id: "r",
	extensions: [ ".r", ".rhistory", ".rprofile", ".rt" ],
	aliases: [ "R", "r" ],
	module: "./r",
});
registerLanguage({
	id: "razor",
	extensions: [".cshtml"],
	aliases: ["Razor", "razor"],
	mimetypes: ["text/x-cshtml"],
	module: "./razor",
});
registerLanguage({
	id: "ruby",
	extensions: [ ".rb", ".rbx", ".rjs", ".gemspec", ".pp" ],
	filenames: [ "rakefile" ],
	aliases: [ "Ruby", "rb" ],
	module: "./ruby",
});
registerLanguage({
	id: "swift",
	aliases: ["Swift", "swift"],
	extensions: [".swift"],
	mimetypes: ["text/swift"],
	module: "./swift",
});
registerLanguage({
	id: "sql",
	extensions: [ ".sql" ],
	aliases: [ "SQL" ],
	module: "./sql",
});
registerLanguage({
	id: "vb",
	extensions: [ ".vb" ],
	aliases: [ "Visual Basic", "vb" ],
	module: "./vb",
});
registerLanguage({
	id: "xml",
	extensions: [ ".xml", ".dtd", ".ascx", ".csproj", ".config", ".wxi", ".wxl", ".wxs", ".xaml", ".svg", ".svgz" ],
	firstLine : "(\\<\\?xml.*)|(\\<svg)|(\\<\\!doctype\\s+svg)",
	aliases: [ "XML", "xml" ],
	mimetypes: ["text/xml", "application/xml", "application/xaml+xml", "application/xml-dtd"],
	module: "./xml",
});
registerLanguage({
	id: "less",
	extensions: [".less"],
	aliases: ["Less", "less"],
	mimetypes: ["text/x-less", "text/less"],
	module: "./less",
});
registerLanguage({
	id: "scss",
	extensions: [".scss"],
	aliases: ["Sass", "sass", "scss"],
	mimetypes: ["text/x-scss", "text/scss"],
	module: "./scss",
});
registerLanguage({
	id: "css",
	extensions: [".css"],
	aliases: ["CSS", "css"],
	mimetypes: ["text/css"],
	module: "./css",
});
registerLanguage({
	id: "yaml",
	extensions: [".yaml", ".yml"],
	aliases: ["YAML", "yaml", "YML", "yml"],
	mimetypes: ["application/x-yaml"],
	module: "./yaml",
});
