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
        return this.applyWithWalker(new NoRequireImportsWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-require-imports",
        description: "Disallows invocation of `require()`.",
        rationale: "Prefer the newer ES6-style imports over `require()`.",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "maintainability",
    };
    Rule.FAILURE_STRING = "require() style import is forbidden";
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoRequireImportsWalker = (function (_super) {
    __extends(NoRequireImportsWalker, _super);
    function NoRequireImportsWalker() {
        _super.apply(this, arguments);
    }
    NoRequireImportsWalker.prototype.visitCallExpression = function (node) {
        if (node.arguments != null && node.expression != null) {
            var callExpressionText = node.expression.getText(this.getSourceFile());
            if (callExpressionText === "require") {
                this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING));
            }
        }
        _super.prototype.visitCallExpression.call(this, node);
    };
    NoRequireImportsWalker.prototype.visitImportEqualsDeclaration = function (node) {
        var moduleReference = node.moduleReference;
        if (moduleReference.kind === ts.SyntaxKind.ExternalModuleReference) {
            this.addFailure(this.createFailure(moduleReference.getStart(), moduleReference.getWidth(), Rule.FAILURE_STRING));
        }
        _super.prototype.visitImportEqualsDeclaration.call(this, node);
    };
    return NoRequireImportsWalker;
}(Lint.RuleWalker));
