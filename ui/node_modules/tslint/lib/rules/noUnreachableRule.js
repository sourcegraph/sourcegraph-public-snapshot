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
        return this.applyWithWalker(new NoUnreachableWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-unreachable",
        description: "Disallows unreachable code after `break`, `catch`, `throw`, and `return` statements.",
        rationale: "Unreachable code is often indication of a logic error.",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING = "unreachable code";
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoUnreachableWalker = (function (_super) {
    __extends(NoUnreachableWalker, _super);
    function NoUnreachableWalker(sourceFile, options) {
        _super.call(this, sourceFile, options);
        this.hasReturned = false;
    }
    NoUnreachableWalker.prototype.visitNode = function (node) {
        var previousReturned = this.hasReturned;
        if (node.kind === ts.SyntaxKind.FunctionDeclaration || node.kind === ts.SyntaxKind.TypeAliasDeclaration) {
            this.hasReturned = false;
        }
        if (this.hasReturned) {
            this.hasReturned = false;
            this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING));
        }
        _super.prototype.visitNode.call(this, node);
        if (node.kind === ts.SyntaxKind.FunctionDeclaration || node.kind === ts.SyntaxKind.TypeAliasDeclaration) {
            this.hasReturned = previousReturned;
        }
    };
    NoUnreachableWalker.prototype.visitBlock = function (node) {
        _super.prototype.visitBlock.call(this, node);
        this.hasReturned = false;
    };
    NoUnreachableWalker.prototype.visitCaseClause = function (node) {
        _super.prototype.visitCaseClause.call(this, node);
        this.hasReturned = false;
    };
    NoUnreachableWalker.prototype.visitDefaultClause = function (node) {
        _super.prototype.visitDefaultClause.call(this, node);
        this.hasReturned = false;
    };
    NoUnreachableWalker.prototype.visitIfStatement = function (node) {
        this.visitNode(node.expression);
        this.visitNode(node.thenStatement);
        this.hasReturned = false;
        if (node.elseStatement != null) {
            this.visitNode(node.elseStatement);
            this.hasReturned = false;
        }
    };
    NoUnreachableWalker.prototype.visitBreakStatement = function (node) {
        _super.prototype.visitBreakStatement.call(this, node);
        this.hasReturned = true;
    };
    NoUnreachableWalker.prototype.visitContinueStatement = function (node) {
        _super.prototype.visitContinueStatement.call(this, node);
        this.hasReturned = true;
    };
    NoUnreachableWalker.prototype.visitReturnStatement = function (node) {
        _super.prototype.visitReturnStatement.call(this, node);
        this.hasReturned = true;
    };
    NoUnreachableWalker.prototype.visitThrowStatement = function (node) {
        _super.prototype.visitThrowStatement.call(this, node);
        this.hasReturned = true;
    };
    return NoUnreachableWalker;
}(Lint.RuleWalker));
