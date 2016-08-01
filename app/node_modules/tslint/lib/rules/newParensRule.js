"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var Lint = require("../lint");
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        var newParensWalker = new NewParensWalker(sourceFile, this.getOptions());
        return this.applyWithWalker(newParensWalker);
    };
    Rule.metadata = {
        ruleName: "new-parens",
        description: "Requires parentheses when invoking a constructor via the `new` keyword.",
        rationale: "Maintains stylistic consistency with other function calls.",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "style",
    };
    Rule.FAILURE_STRING = "Parentheses are required when invoking a constructor";
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NewParensWalker = (function (_super) {
    __extends(NewParensWalker, _super);
    function NewParensWalker() {
        _super.apply(this, arguments);
    }
    NewParensWalker.prototype.visitNewExpression = function (node) {
        if (node.arguments === undefined) {
            this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING));
        }
        _super.prototype.visitNewExpression.call(this, node);
    };
    return NewParensWalker;
}(Lint.RuleWalker));
