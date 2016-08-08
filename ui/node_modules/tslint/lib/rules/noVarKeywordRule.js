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
        var noVarKeywordWalker = new NoVarKeywordWalker(sourceFile, this.getOptions());
        return this.applyWithWalker(noVarKeywordWalker);
    };
    Rule.metadata = {
        ruleName: "no-var-keyword",
        description: "Disallows usage of the `var` keyword.",
        descriptionDetails: "Use `let` or `const` instead.",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING = "Forbidden 'var' keyword, use 'let' or 'const' instead";
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoVarKeywordWalker = (function (_super) {
    __extends(NoVarKeywordWalker, _super);
    function NoVarKeywordWalker() {
        _super.apply(this, arguments);
    }
    NoVarKeywordWalker.prototype.visitVariableStatement = function (node) {
        if (!Lint.hasModifier(node.modifiers, ts.SyntaxKind.ExportKeyword, ts.SyntaxKind.DeclareKeyword)
            && !Lint.isBlockScopedVariable(node)) {
            this.addFailure(this.createFailure(node.getStart(), "var".length, Rule.FAILURE_STRING));
        }
        _super.prototype.visitVariableStatement.call(this, node);
    };
    NoVarKeywordWalker.prototype.visitForStatement = function (node) {
        this.handleInitializerNode(node.initializer);
        _super.prototype.visitForStatement.call(this, node);
    };
    NoVarKeywordWalker.prototype.visitForInStatement = function (node) {
        this.handleInitializerNode(node.initializer);
        _super.prototype.visitForInStatement.call(this, node);
    };
    NoVarKeywordWalker.prototype.visitForOfStatement = function (node) {
        this.handleInitializerNode(node.initializer);
        _super.prototype.visitForOfStatement.call(this, node);
    };
    NoVarKeywordWalker.prototype.handleInitializerNode = function (node) {
        if (node && node.kind === ts.SyntaxKind.VariableDeclarationList &&
            !(Lint.isNodeFlagSet(node, ts.NodeFlags.Let) || Lint.isNodeFlagSet(node, ts.NodeFlags.Const))) {
            this.addFailure(this.createFailure(node.getStart(), "var".length, Rule.FAILURE_STRING));
        }
    };
    return NoVarKeywordWalker;
}(Lint.RuleWalker));
