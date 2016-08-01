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
        return this.applyWithWalker(new TrailingCommaWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "trailing-comma",
        description: "Requires or disallows trailing commas in array and object literals, destructuring assignments and named imports.",
        optionsDescription: (_a = ["\n            One argument which is an object with the keys `multiline` and `singleline`.\n            Both should be set to either `\"always\"` or `\"never\"`.\n\n            * `\"multiline\"` checks multi-line object literals.\n            * `\"singleline\"` checks single-line object literals.\n\n            A array is considered \"multiline\" if its closing bracket is on a line\n            after the last array element. The same general logic is followed for\n            object literals and named import statements."], _a.raw = ["\n            One argument which is an object with the keys \\`multiline\\` and \\`singleline\\`.\n            Both should be set to either \\`\"always\"\\` or \\`\"never\"\\`.\n\n            * \\`\"multiline\"\\` checks multi-line object literals.\n            * \\`\"singleline\"\\` checks single-line object literals.\n\n            A array is considered \"multiline\" if its closing bracket is on a line\n            after the last array element. The same general logic is followed for\n            object literals and named import statements."], Lint.Utils.dedent(_a)),
        options: {
            type: "object",
            properties: {
                multiline: {
                    type: "string",
                    enum: ["always", "never"],
                },
                singleline: {
                    type: "string",
                    enum: ["always", "never"],
                },
            },
            additionalProperties: false,
        },
        optionExamples: ['[true, {"multiline": "always", "singleline": "never"}]'],
        type: "maintainability",
    };
    Rule.FAILURE_STRING_NEVER = "Unnecessary trailing comma";
    Rule.FAILURE_STRING_ALWAYS = "Missing trailing comma";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var TrailingCommaWalker = (function (_super) {
    __extends(TrailingCommaWalker, _super);
    function TrailingCommaWalker() {
        _super.apply(this, arguments);
    }
    TrailingCommaWalker.prototype.visitArrayLiteralExpression = function (node) {
        this.lintNode(node);
        _super.prototype.visitArrayLiteralExpression.call(this, node);
    };
    TrailingCommaWalker.prototype.visitBindingPattern = function (node) {
        if (node.kind === ts.SyntaxKind.ArrayBindingPattern || node.kind === ts.SyntaxKind.ObjectBindingPattern) {
            this.lintNode(node);
        }
        _super.prototype.visitBindingPattern.call(this, node);
    };
    TrailingCommaWalker.prototype.visitNamedImports = function (node) {
        this.lintNode(node);
        _super.prototype.visitNamedImports.call(this, node);
    };
    TrailingCommaWalker.prototype.visitObjectLiteralExpression = function (node) {
        this.lintNode(node);
        _super.prototype.visitObjectLiteralExpression.call(this, node);
    };
    TrailingCommaWalker.prototype.lintNode = function (node) {
        var child = node.getChildAt(1);
        if (child != null && child.kind === ts.SyntaxKind.SyntaxList) {
            var grandChildren = child.getChildren();
            if (grandChildren.length > 0) {
                var lastGrandChild = grandChildren[grandChildren.length - 1];
                var hasTrailingComma = lastGrandChild.kind === ts.SyntaxKind.CommaToken;
                var endLineOfNode = this.getSourceFile().getLineAndCharacterOfPosition(node.getEnd()).line;
                var endLineOfLastElement = this.getSourceFile().getLineAndCharacterOfPosition(lastGrandChild.getEnd()).line;
                var isMultiline = endLineOfNode !== endLineOfLastElement;
                var option = this.getOption(isMultiline ? "multiline" : "singleline");
                if (hasTrailingComma && option === "never") {
                    this.addFailure(this.createFailure(lastGrandChild.getStart(), 1, Rule.FAILURE_STRING_NEVER));
                }
                else if (!hasTrailingComma && option === "always") {
                    this.addFailure(this.createFailure(lastGrandChild.getEnd() - 1, 1, Rule.FAILURE_STRING_ALWAYS));
                }
            }
        }
    };
    TrailingCommaWalker.prototype.getOption = function (option) {
        var allOptions = this.getOptions();
        if (allOptions == null || allOptions.length === 0) {
            return null;
        }
        return allOptions[0][option];
    };
    return TrailingCommaWalker;
}(Lint.RuleWalker));
