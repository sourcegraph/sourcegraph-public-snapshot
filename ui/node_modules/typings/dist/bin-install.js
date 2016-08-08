"use strict";
var listify = require('listify');
var typings_core_1 = require('typings-core');
var cli_1 = require('./support/cli');
function help() {
    return "\ntypings install (with no arguments, in package directory)\ntypings install [<name>=]<location>\n\n  <name>      Module name of the installed definition\n  <location>  The location to read from (described below)\n\nValid Locations:\n  [<source>~]<pkg>[@<version>][#<tag>]\n  file:<path>\n  github:<org>/<repo>[/<path>][#<commitish>]\n  bitbucket:<org>/<repo>[/<path>][#<commitish>]\n  npm:<pkg>[/<path>]\n  bower:<pkg>[/<path>]\n  http(s)://<host>/<path>\n\n  <source>    The registry mirror: \"npm\", \"bower\", \"env\", \"global\", \"lib\" or \"dt\"\n              When not specified, `defaultSource` in `.typingsrc` will be used.\n  <path>      Path to a `.d.ts` file or `typings.json`\n  <host>      A domain name (with optional port)\n  <version>   A semver range (E.g. \">=4.0\")\n  <tag>       The specific tag of a registry entry\n  <commitish> A git commit, tag or branch\n\nOptions:\n  [--save|-S]       Persist to \"dependencies\"\n  [--save-dev|-D]   Persist to \"devDependencies\"\n  [--save-peer|-P]  Persist to \"peerDependencies\"\n  [--global|-G]     Install and persist as an global definition\n    [-SG]           Persist to \"globalDependencies\"\n    [-DG]           Persist to \"globalDevDependencies\"\n  [--production]    Install only production dependencies (omits dev dependencies)\n\nAliases: i, in\n";
}
exports.help = help;
function exec(args, options) {
    var emitter = options.emitter;
    if (typeof options.ambient !== 'undefined') {
        cli_1.logError('The "ambient" flag is deprecated. Please use "global" instead', 'deprecated');
        return;
    }
    if (args.length === 0) {
        return typings_core_1.install(options)
            .then(function (result) {
            console.log(cli_1.archifyDependencyTree(result));
        });
    }
    emitter.on('reference', function (_a) {
        var reference = _a.reference, resolution = _a.resolution, name = _a.name;
        cli_1.logInfo("Stripped reference \"" + reference + "\" during installation from \"" + name + "\" (" + resolution + ")", 'reference');
    });
    emitter.on('globaldependencies', function (_a) {
        var name = _a.name, dependencies = _a.dependencies;
        var deps = Object.keys(dependencies).map(function (x) { return JSON.stringify(x); });
        if (deps.length) {
            cli_1.logInfo("\"" + name + "\" lists global dependencies on " + listify(deps) + " and should be installed", 'globaldependencies');
        }
    });
    return typings_core_1.installDependenciesRaw(args, options)
        .then(function (results) {
        for (var _i = 0, results_1 = results; _i < results_1.length; _i++) {
            var result = results_1[_i];
            console.log(cli_1.archifyDependencyTree(result));
        }
    });
}
exports.exec = exec;
//# sourceMappingURL=bin-install.js.map