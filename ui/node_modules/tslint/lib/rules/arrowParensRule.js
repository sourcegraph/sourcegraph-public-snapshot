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
        var newParensWalker = new ArrowParensWalker(sourceFile, this.getOptions());
        return this.applyWithWalker(newParensWalker);
    };
    Rule.metadata = {
        ruleName: "arrow-parens",
        description: "Requires parentheses around the parameters of arrow function definitions.",
        rationale: "Maintains stylistic consistency with other arrow function definitions.",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "style",
    };
    Rule.FAILURE_STRING = "Parentheses are required around the parameters of an arrow function definition";
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var ArrowParensWalker = (function (_super) {
    __extends(ArrowParensWalker, _super);
    function ArrowParensWalker() {
        _super.apply(this, arguments);
    }
    ArrowParensWalker.prototype.visitArrowFunction = function (node) {
        if (node.parameters.length === 1) {
            var parameter = node.parameters[0];
            var text = parameter.getText();
            var firstToken = node.getFirstToken();
            var lastToken = node.getChildAt(2);
            var width = text.length;
            var position = parameter.getStart();
            var isGenerics = false;
            if (firstToken.kind === ts.SyntaxKind.LessThanToken) {
                isGenerics = true;
            }
            if ((firstToken.kind !== ts.SyntaxKind.OpenParenToken || lastToken.kind !== ts.SyntaxKind.CloseParenToken)
                && !isGenerics && node.flags !== ts.NodeFlags.Async) {
                this.addFailure(this.createFailure(position, width, Rule.FAILURE_STRING));
            }
        }
        _super.prototype.visitArrowFunction.call(this, node);
    };
    return ArrowParensWalker;
}(Lint.RuleWalker));
