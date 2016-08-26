"use strict";
var colors = require("colors");
var diff = require("diff");
var fs = require("fs");
var glob = require("glob");
var path = require("path");
var ts = require("typescript");
var rule_1 = require("./language/rule/rule");
var utils_1 = require("./language/utils");
var parse = require("./test/parse");
var Linter = require("./tslint");
var MARKUP_FILE_EXTENSION = ".lint";
var FIXES_FILE_EXTENSION = ".fix";
function runTest(testDirectory, rulesDirectory) {
    var filesToLint = glob.sync(path.join(testDirectory, "**/*" + MARKUP_FILE_EXTENSION));
    var tslintConfig = Linter.findConfiguration(path.join(testDirectory, "tslint.json"), null);
    var results = { directory: testDirectory, results: {} };
    var _loop_1 = function(fileToLint) {
        var fileBasename = path.basename(fileToLint, MARKUP_FILE_EXTENSION);
        var fileCompileName = fileBasename.replace(/\.lint$/, "");
        var fileText = fs.readFileSync(fileToLint, "utf8");
        var fileTextWithoutMarkup = parse.removeErrorMarkup(fileText);
        var errorsFromMarkup = parse.parseErrorsFromMarkup(fileText);
        var program = void 0;
        if (tslintConfig.linterOptions && tslintConfig.linterOptions.typeCheck) {
            var compilerOptions_1 = utils_1.createCompilerOptions();
            var compilerHost = {
                fileExists: function () { return true; },
                getCanonicalFileName: function (filename) { return filename; },
                getCurrentDirectory: function () { return ""; },
                getDefaultLibFileName: function () { return ts.getDefaultLibFileName(compilerOptions_1); },
                getNewLine: function () { return "\n"; },
                getSourceFile: function (filenameToGet) {
                    if (filenameToGet === this.getDefaultLibFileName()) {
                        var fileText_1 = fs.readFileSync(ts.getDefaultLibFilePath(compilerOptions_1)).toString();
                        return ts.createSourceFile(filenameToGet, fileText_1, compilerOptions_1.target);
                    }
                    else if (filenameToGet === fileCompileName) {
                        return ts.createSourceFile(fileBasename, fileTextWithoutMarkup, compilerOptions_1.target, true);
                    }
                },
                readFile: function () { return null; },
                useCaseSensitiveFileNames: function () { return true; },
                writeFile: function () { return null; },
            };
            program = ts.createProgram([fileCompileName], compilerOptions_1, compilerHost);
            ts.getPreEmitDiagnostics(program);
        }
        var lintOptions = {
            configuration: tslintConfig,
            formatter: "prose",
            formattersDirectory: "",
            rulesDirectory: rulesDirectory,
        };
        var linter = new Linter(fileBasename, fileTextWithoutMarkup, lintOptions, program);
        var failures = linter.lint().failures;
        var errorsFromLinter = failures.map(function (failure) {
            var startLineAndCharacter = failure.getStartPosition().getLineAndCharacter();
            var endLineAndCharacter = failure.getEndPosition().getLineAndCharacter();
            return {
                endPos: {
                    col: endLineAndCharacter.character,
                    line: endLineAndCharacter.line,
                },
                message: failure.getFailure(),
                startPos: {
                    col: startLineAndCharacter.character,
                    line: startLineAndCharacter.line,
                },
            };
        });
        var fixedFileText = void 0;
        var newFileText = void 0;
        try {
            var fixedFile = fileToLint.replace(/\.lint$/, FIXES_FILE_EXTENSION);
            var stat = fs.statSync(fixedFile);
            if (stat.isFile()) {
                fixedFileText = fs.readFileSync(fixedFile, "utf8");
                var fixes = failures.filter(function (f) { return f.hasFix(); }).map(function (f) { return f.getFix(); });
                newFileText = rule_1.Fix.applyAll(fileTextWithoutMarkup, fixes);
            }
        }
        catch (e) {
            fixedFileText = "";
            newFileText = "";
        }
        results.results[fileToLint] = {
            errorsFromLinter: errorsFromLinter,
            errorsFromMarkup: errorsFromMarkup,
            fixesFromLinter: newFileText,
            fixesFromMarkup: fixedFileText,
            markupFromLinter: parse.createMarkupFromErrors(fileTextWithoutMarkup, errorsFromMarkup),
            markupFromMarkup: parse.createMarkupFromErrors(fileTextWithoutMarkup, errorsFromLinter),
        };
    };
    for (var _i = 0, filesToLint_1 = filesToLint; _i < filesToLint_1.length; _i++) {
        var fileToLint = filesToLint_1[_i];
        _loop_1(fileToLint);
    }
    return results;
}
exports.runTest = runTest;
function consoleTestResultHandler(testResult) {
    var didAllTestsPass = true;
    for (var _i = 0, _a = Object.keys(testResult.results); _i < _a.length; _i++) {
        var fileName = _a[_i];
        var results = testResult.results[fileName];
        process.stdout.write(fileName + ":");
        var markupDiffResults = diff.diffLines(results.markupFromMarkup, results.markupFromLinter);
        var fixesDiffResults = diff.diffLines(results.fixesFromMarkup, results.fixesFromLinter);
        var didMarkupTestPass = !markupDiffResults.some(function (diff) { return diff.added || diff.removed; });
        var didFixesTestPass = !fixesDiffResults.some(function (diff) { return diff.added || diff.removed; });
        if (didMarkupTestPass && didFixesTestPass) {
            console.log(colors.green(" Passed"));
        }
        else {
            console.log(colors.red(" Failed!"));
            didAllTestsPass = false;
            if (!didMarkupTestPass) {
                displayDiffResults(markupDiffResults, MARKUP_FILE_EXTENSION);
            }
            if (!didFixesTestPass) {
                displayDiffResults(fixesDiffResults, FIXES_FILE_EXTENSION);
            }
        }
    }
    return didAllTestsPass;
}
exports.consoleTestResultHandler = consoleTestResultHandler;
function displayDiffResults(diffResults, extension) {
    console.log(colors.green("Expected (from " + extension + " file)"));
    console.log(colors.red("Actual (from TSLint)"));
    for (var _i = 0, diffResults_1 = diffResults; _i < diffResults_1.length; _i++) {
        var diffResult = diffResults_1[_i];
        var color = colors.grey;
        if (diffResult.added) {
            color = colors.green;
        }
        else if (diffResult.removed) {
            color = colors.red;
        }
        process.stdout.write(color(diffResult.value));
    }
}
