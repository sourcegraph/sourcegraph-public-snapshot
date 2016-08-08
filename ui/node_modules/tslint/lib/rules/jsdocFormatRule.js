"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ts = require("typescript");
var Lint = require("../lint");
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        return this.applyWithWalker(new JsdocWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "jsdoc-format",
        description: "Enforces basic format rules for JSDoc comments.",
        descriptionDetails: (_a = ["\n            The following rules are enforced for JSDoc comments (comments starting with `/**`):\n\n            * each line contains an asterisk and asterisks must be aligned\n            * each asterisk must be followed by either a space or a newline (except for the first and the last)\n            * the only characters before the asterisk on each line must be whitespace characters\n            * one line comments must start with `/** ` and end with `*/`"], _a.raw = ["\n            The following rules are enforced for JSDoc comments (comments starting with \\`/**\\`):\n\n            * each line contains an asterisk and asterisks must be aligned\n            * each asterisk must be followed by either a space or a newline (except for the first and the last)\n            * the only characters before the asterisk on each line must be whitespace characters\n            * one line comments must start with \\`/** \\` and end with \\`*/\\`"], Lint.Utils.dedent(_a)),
        rationale: "Helps maintain a consistent, readable style for JSDoc comments.",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "style",
    };
    Rule.ALIGNMENT_FAILURE_STRING = "asterisks in jsdoc must be aligned";
    Rule.FORMAT_FAILURE_STRING = "jsdoc is not formatted correctly on this line";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var JsdocWalker = (function (_super) {
    __extends(JsdocWalker, _super);
    function JsdocWalker() {
        _super.apply(this, arguments);
    }
    JsdocWalker.prototype.visitSourceFile = function (node) {
        var _this = this;
        _super.prototype.visitSourceFile.call(this, node);
        Lint.scanAllTokens(ts.createScanner(ts.ScriptTarget.ES5, false, ts.LanguageVariant.Standard, node.text), function (scanner) {
            var startPos = scanner.getStartPos();
            if (_this.tokensToSkipStartEndMap[startPos] != null) {
                scanner.setTextPos(_this.tokensToSkipStartEndMap[startPos]);
                return;
            }
            if (scanner.getToken() === ts.SyntaxKind.MultiLineCommentTrivia) {
                var commentText = scanner.getTokenText();
                var startPosition = scanner.getTokenPos();
                _this.findFailuresForJsdocComment(commentText, startPosition, node);
            }
        });
    };
    JsdocWalker.prototype.findFailuresForJsdocComment = function (commentText, startingPosition, sourceFile) {
        var currentPosition = startingPosition;
        var lines = commentText.split(/\r?\n/);
        var firstLine = lines[0];
        var jsdocPosition = currentPosition;
        var isJsdocMatch = firstLine.match(/^\s*\/\*\*/);
        if (isJsdocMatch != null) {
            if (lines.length === 1) {
                var firstLineMatch = firstLine.match(/^\s*\/\*\* (.* )?\*\/$/);
                if (firstLineMatch == null) {
                    this.addFailureAt(jsdocPosition, firstLine.length, Rule.FORMAT_FAILURE_STRING);
                }
                return;
            }
            var indexToMatch = firstLine.indexOf("**") + sourceFile.getLineAndCharacterOfPosition(currentPosition).character;
            var otherLines = lines.splice(1, lines.length - 2);
            jsdocPosition += firstLine.length + 1;
            for (var _i = 0, otherLines_1 = otherLines; _i < otherLines_1.length; _i++) {
                var line = otherLines_1[_i];
                var asteriskMatch = line.match(/^\s*\*( |$)/);
                if (asteriskMatch == null) {
                    this.addFailureAt(jsdocPosition, line.length, Rule.FORMAT_FAILURE_STRING);
                }
                var asteriskIndex = line.indexOf("*");
                if (asteriskIndex !== indexToMatch) {
                    this.addFailureAt(jsdocPosition, line.length, Rule.ALIGNMENT_FAILURE_STRING);
                }
                jsdocPosition += line.length + 1;
            }
            var lastLine = lines[lines.length - 1];
            var endBlockCommentMatch = lastLine.match(/^\s*\*\/$/);
            if (endBlockCommentMatch == null) {
                this.addFailureAt(jsdocPosition, lastLine.length, Rule.FORMAT_FAILURE_STRING);
            }
            var lastAsteriskIndex = lastLine.indexOf("*");
            if (lastAsteriskIndex !== indexToMatch) {
                this.addFailureAt(jsdocPosition, lastLine.length, Rule.ALIGNMENT_FAILURE_STRING);
            }
        }
    };
    JsdocWalker.prototype.addFailureAt = function (currentPosition, width, failureString) {
        var failure = this.createFailure(currentPosition, width, failureString);
        this.addFailure(failure);
    };
    return JsdocWalker;
}(Lint.SkippableTokenAwareRuleWalker));
