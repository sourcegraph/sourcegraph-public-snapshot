"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ts = require("typescript");
var Lint = require("../lint");
var OPTION_LINEBREAK_STYLE_CRLF = "CRLF";
var OPTION_LINEBREAK_STYLE_LF = "LF";
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        var failures = [];
        var scanner = ts.createScanner(sourceFile.languageVersion, false, sourceFile.languageVariant, sourceFile.getFullText());
        var linebreakStyle = this.getOptions().ruleArguments[0] || OPTION_LINEBREAK_STYLE_LF;
        var expectLF = linebreakStyle === OPTION_LINEBREAK_STYLE_CRLF;
        var expectedEOL = expectLF ? "\r\n" : "\n";
        var failureString = expectLF ? Rule.FAILURE_STRINGS.CRLF : Rule.FAILURE_STRINGS.LF;
        for (var token = scanner.scan(); token !== ts.SyntaxKind.EndOfFileToken; token = scanner.scan()) {
            if (token === ts.SyntaxKind.NewLineTrivia) {
                var text = scanner.getTokenText();
                if (text !== expectedEOL) {
                    failures.push(this.createFailure(sourceFile, scanner, failureString));
                }
            }
        }
        return failures;
    };
    Rule.prototype.createFailure = function (sourceFile, scanner, failure) {
        var start = sourceFile.getPositionOfLineAndCharacter(sourceFile.getLineAndCharacterOfPosition(scanner.getStartPos()).line, 0);
        var end = scanner.getStartPos();
        return new Lint.RuleFailure(sourceFile, start, end, failure, this.getOptions().ruleName);
    };
    Rule.metadata = {
        ruleName: "linebreak-style",
        description: "Enforces a consistent linebreak style.",
        optionsDescription: (_a = ["\n            One of the following options must be provided:\n\n            * `\"", "\"` requires LF (`\\n`) linebreaks\n            * `\"", "\"` requires CRLF (`\\r\\n`) linebreaks"], _a.raw = ["\n            One of the following options must be provided:\n\n            * \\`\"", "\"\\` requires LF (\\`\\\\n\\`) linebreaks\n            * \\`\"", "\"\\` requires CRLF (\\`\\\\r\\\\n\\`) linebreaks"], Lint.Utils.dedent(_a, OPTION_LINEBREAK_STYLE_LF, OPTION_LINEBREAK_STYLE_CRLF)),
        options: {
            type: "string",
            enum: [OPTION_LINEBREAK_STYLE_LF, OPTION_LINEBREAK_STYLE_CRLF],
        },
        optionExamples: [("[true, \"" + OPTION_LINEBREAK_STYLE_LF + "\"]"), ("[true, \"" + OPTION_LINEBREAK_STYLE_CRLF + "\"]")],
        type: "maintainability",
    };
    Rule.FAILURE_STRINGS = {
        CRLF: "Expected linebreak to be '" + OPTION_LINEBREAK_STYLE_CRLF + "'",
        LF: "Expected linebreak to be '" + OPTION_LINEBREAK_STYLE_LF + "'",
    };
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
