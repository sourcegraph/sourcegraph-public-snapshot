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
        return this.applyWithWalker(new NoDuplicateVariableWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-duplicate-variable",
        description: "Disallows duplicate variable declarations in the same block scope.",
        descriptionDetails: (_a = ["\n            This rule is only useful when using the `var` keyword -\n            the compiler will detect redeclarations of `let` and `const` variables."], _a.raw = ["\n            This rule is only useful when using the \\`var\\` keyword -\n            the compiler will detect redeclarations of \\`let\\` and \\`const\\` variables."], Lint.Utils.dedent(_a)),
        rationale: (_b = ["\n            A variable can be reassigned if necessary -\n            there's no good reason to have a duplicate variable declaration."], _b.raw = ["\n            A variable can be reassigned if necessary -\n            there's no good reason to have a duplicate variable declaration."], Lint.Utils.dedent(_b)),
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING_FACTORY = function (name) { return ("Duplicate variable: '" + name + "'"); };
    return Rule;
    var _a, _b;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoDuplicateVariableWalker = (function (_super) {
    __extends(NoDuplicateVariableWalker, _super);
    function NoDuplicateVariableWalker() {
        _super.apply(this, arguments);
    }
    NoDuplicateVariableWalker.prototype.createScope = function () {
        return null;
    };
    NoDuplicateVariableWalker.prototype.createBlockScope = function () {
        return new ScopeInfo();
    };
    NoDuplicateVariableWalker.prototype.visitBindingElement = function (node) {
        var isSingleVariable = node.name.kind === ts.SyntaxKind.Identifier;
        var isBlockScoped = Lint.isBlockScopedBindingElement(node);
        if (isSingleVariable && !isBlockScoped) {
            this.handleSingleVariableIdentifier(node.name);
        }
        _super.prototype.visitBindingElement.call(this, node);
    };
    NoDuplicateVariableWalker.prototype.visitCatchClause = function (node) {
        this.visitBlock(node.block);
    };
    NoDuplicateVariableWalker.prototype.visitMethodSignature = function (node) {
    };
    NoDuplicateVariableWalker.prototype.visitTypeLiteral = function (node) {
    };
    NoDuplicateVariableWalker.prototype.visitVariableDeclaration = function (node) {
        var isSingleVariable = node.name.kind === ts.SyntaxKind.Identifier;
        if (isSingleVariable && !Lint.isBlockScopedVariable(node)) {
            this.handleSingleVariableIdentifier(node.name);
        }
        _super.prototype.visitVariableDeclaration.call(this, node);
    };
    NoDuplicateVariableWalker.prototype.handleSingleVariableIdentifier = function (variableIdentifier) {
        var variableName = variableIdentifier.text;
        var currentBlockScope = this.getCurrentBlockScope();
        if (currentBlockScope.varNames.indexOf(variableName) >= 0) {
            this.addFailureOnIdentifier(variableIdentifier);
        }
        else {
            currentBlockScope.varNames.push(variableName);
        }
    };
    NoDuplicateVariableWalker.prototype.addFailureOnIdentifier = function (ident) {
        var failureString = Rule.FAILURE_STRING_FACTORY(ident.text);
        this.addFailure(this.createFailure(ident.getStart(), ident.getWidth(), failureString));
    };
    return NoDuplicateVariableWalker;
}(Lint.BlockScopeAwareRuleWalker));
var ScopeInfo = (function () {
    function ScopeInfo() {
        this.varNames = [];
    }
    return ScopeInfo;
}());
