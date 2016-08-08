#!/usr/bin/env node
"use strict";
var minimist = require('minimist');
var wordwrap = require('wordwrap');
var path_1 = require('path');
var chalk = require('chalk');
var updateNotifier = require('update-notifier');
var extend = require('xtend');
var events_1 = require('events');
var cli_1 = require('./support/cli');
var aliases_1 = require('./aliases');
var pkg = require('../package.json');
var argv = minimist(process.argv.slice(2), {
    boolean: ['version', 'save', 'saveDev', 'savePeer', 'global', 'verbose', 'production'],
    string: ['cwd', 'out', 'name', 'source', 'offset', 'limit', 'sort'],
    alias: {
        global: ['G'],
        version: ['v'],
        save: ['S'],
        saveDev: ['save-dev', 'D'],
        savePeer: ['savePeer', 'P'],
        verbose: ['V'],
        out: ['o'],
        help: ['h']
    },
    default: {
        production: process.env.NODE_ENV === 'production'
    }
});
var cwd = argv.cwd ? path_1.resolve(argv.cwd) : process.cwd();
var emitter = new events_1.EventEmitter();
var args = extend(argv, { emitter: emitter, cwd: cwd });
updateNotifier({ pkg: pkg }).notify();
exec(args);
emitter.on('enoent', function (_a) {
    var path = _a.path;
    cli_1.logWarning("Path \"" + path + "\" is missing", 'enoent');
});
emitter.on('hastypings', function (_a) {
    var name = _a.name, typings = _a.typings;
    cli_1.logWarning(("Typings for \"" + name + "\" already exist in \"" + path_1.relative(cwd, typings) + "\". You should ") +
        "let TypeScript resolve the packaged typings and uninstall the copy installed by Typings", 'hastypings');
});
emitter.on('postmessage', function (_a) {
    var message = _a.message, name = _a.name;
    cli_1.logInfo(name + ": " + message, 'postmessage');
});
emitter.on('badlocation', function (_a) {
    var raw = _a.raw;
    cli_1.logWarning("\"" + raw + "\" is mutable and may change, consider specifying a commit hash", 'badlocation');
});
emitter.on('deprecated', function (_a) {
    var date = _a.date, raw = _a.raw, parent = _a.parent;
    if (parent == null || parent.raw == null) {
        cli_1.logWarning(date.toLocaleDateString() + ": \"" + raw + "\" is deprecated (updated, replaced or removed)", 'deprecated');
    }
});
emitter.on('prune', function (_a) {
    var name = _a.name, global = _a.global, resolution = _a.resolution;
    var suffix = chalk.gray((" (" + resolution + ")") + (global ? ' (global)' : ''));
    cli_1.logInfo("" + name + suffix, 'prune');
});
function exec(options) {
    if (options._.length) {
        var command = aliases_1.aliases[options._[0]];
        var args_1 = options._.slice(1);
        if (command != null) {
            if (options.help) {
                return console.log(command.help());
            }
            return cli_1.handle(command.exec(args_1, options), options);
        }
    }
    else if (options.version) {
        console.log(pkg.version);
        return;
    }
    var wrap = wordwrap(4, 80);
    console.log("\nUsage: typings <command>\n\nCommands:\n" + wrap(Object.keys(aliases_1.aliases).sort().join(', ')) + "\n\ntypings <command> -h   Get help for <command>\ntypings <command> -V   Enable verbose logging\n\ntypings --version      Print the CLI version\n\ntypings@" + pkg.version + " " + path_1.join(__dirname, '..') + "\n");
}
//# sourceMappingURL=bin.js.map