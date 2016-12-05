"use strict";
/* tslint:disable:no-empty */
try {
    // cackward compatibility for node v0.12
    require("es6-promise").polyfill();
}
catch (e) {
}
/* tslint:enable:no-empty */
try {
    require("typescript");
}
catch (e) {
    console.error("typescript is required. please try 'npm install -g typescript'\n");
}
var fs = require("fs");
var commandpost = require("commandpost");
var lib = require("./");
var utils_1 = require("./utils");
var packageJson = JSON.parse(fs.readFileSync(__dirname + "/../package.json").toString());
var root = commandpost
    .create("tsfmt [files...]")
    .version(packageJson.version, "-v, --version")
    .option("-r, --replace", "replace .ts file")
    .option("--verify", "checking file format")
    .option("--baseDir <path>", "config file lookup from <path>")
    .option("--stdin", "get formatting content from stdin")
    .option("--no-tsconfig", "don't read a tsconfig.json")
    .option("--no-tslint", "don't read a tslint.json")
    .option("--no-editorconfig", "don't read a .editorconfig")
    .option("--no-tsfmt", "don't read a tsfmt.json")
    .option("--verbose", "makes output more verbose")
    .action(function (opts, args) {
    var replace = !!opts.replace;
    var verify = !!opts.verify;
    var baseDir = opts.baseDir ? opts.baseDir[0] : void 0;
    var stdin = !!opts.stdin;
    var tsconfig = !!opts.tsconfig;
    var tslint = !!opts.tslint;
    var editorconfig = !!opts.editorconfig;
    var tsfmt = !!opts.tsfmt;
    var verbose = !!opts.verbose;
    var files = args.files;
    var useTsconfig = false;
    if (files.length === 0) {
        var configFileName = utils_1.getConfigFileName(baseDir || process.cwd(), "tsconfig.json");
        if (configFileName) {
            files = utils_1.readFilesFromTsconfig(configFileName);
            if (verbose) {
                console.log("read: " + configFileName);
            }
            useTsconfig = true;
        }
    }
    if (files.length === 0 && !opts.stdin) {
        process.stdout.write(root.helpText() + "\n");
        return;
    }
    if (verbose) {
        console.log("replace:	  " + (replace ? "ON" : "OFF"));
        console.log("verify:	   " + (verify ? "ON" : "OFF"));
        console.log("baseDir:	   " + (baseDir ? baseDir : process.cwd()));
        console.log("stdin:		" + (stdin ? "ON" : "OFF"));
        console.log("files from tsconfig:	 " + (useTsconfig ? "ON" : "OFF"));
        console.log("tsconfig:	 " + (tsconfig ? "ON" : "OFF"));
        console.log("tslint:	   " + (tslint ? "ON" : "OFF"));
        console.log("editorconfig: " + (editorconfig ? "ON" : "OFF"));
        console.log("tsfmt:		" + (tsfmt ? "ON" : "OFF"));
    }
    if (stdin) {
        if (replace) {
            errorHandler("--stdin option can not use with --replace option");
            return;
        }
        lib
            .processStream(files[0] || "temp.ts", process.stdin, {
            replace: replace,
            verify: verify,
            baseDir: baseDir,
            tsconfig: tsconfig,
            tslint: tslint,
            editorconfig: editorconfig,
            tsfmt: tsfmt,
            verbose: verbose,
        })
            .then(function (result) {
            var resultMap = {};
            resultMap[result.fileName] = result;
            return resultMap;
        })
            .then(showResultHandler)
            .catch(errorHandler);
    }
    else {
        lib
            .processFiles(files, {
            replace: replace,
            verify: verify,
            baseDir: baseDir,
            tsconfig: tsconfig,
            tslint: tslint,
            editorconfig: editorconfig,
            tsfmt: tsfmt,
            verbose: verbose,
        })
            .then(showResultHandler)
            .catch(errorHandler);
    }
});
commandpost
    .exec(root, process.argv)
    .catch(errorHandler);
function showResultHandler(resultMap) {
    var hasError = Object.keys(resultMap).filter(function (fileName) { return resultMap[fileName].error; }).length !== 0;
    if (hasError) {
        Object.keys(resultMap)
            .map(function (fileName) { return resultMap[fileName]; })
            .filter(function (result) { return result.error; })
            .forEach(function (result) { return process.stderr.write(result.message); });
        process.exit(1);
    }
    else {
        Object.keys(resultMap)
            .map(function (fileName) { return resultMap[fileName]; })
            .forEach(function (result) {
            if (result.message) {
                process.stdout.write(result.message);
            }
        });
    }
    return Promise.resolve(null);
}
function errorHandler(err) {
    if (err instanceof Error) {
        console.error(err.stack);
    }
    else {
        console.error(err);
    }
    return Promise.resolve(null).then(function () {
        process.exit(1);
        return null;
    });
}
//# sourceMappingURL=cli.js.map