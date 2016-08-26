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
        return this.applyWithWalker(new NoShadowedVariableWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-shadowed-variable",
        description: "Disallows shadowing variable declarations.",
        rationale: "Shadowing a variable masks access to it and obscures to what value an identifier actually refers.",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING_FACTORY = function (name) { return ("Shadowed variable: '" + name + "'"); };
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoShadowedVariableWalker = (function (_super) {
    __extends(NoShadowedVariableWalker, _super);
    function NoShadowedVariableWalker() {
        _super.apply(this, arguments);
    }
    NoShadowedVariableWalker.prototype.createScope = function () {
        return new ScopeInfo();
    };
    NoShadowedVariableWalker.prototype.createBlockScope = function () {
        return new ScopeInfo();
    };
    NoShadowedVariableWalker.prototype.visitBindingElement = function (node) {
        var isSingleVariable = node.name.kind === ts.SyntaxKind.Identifier;
        var variableDeclaration = Lint.getBindingElementVariableDeclaration(node);
        if (isSingleVariable) {
            var name_1 = node.name;
            if (variableDeclaration) {
                var isBlockScopedVariable = Lint.isBlockScopedVariable(variableDeclaration);
                this.handleSingleVariableIdentifier(name_1, isBlockScopedVariable);
            }
            else {
                this.handleSingleParameterIdentifier(name_1);
            }
        }
        _super.prototype.visitBindingElement.call(this, node);
    };
    NoShadowedVariableWalker.prototype.visitCatchClause = function (node) {
        this.visitBlock(node.block);
    };
    NoShadowedVariableWalker.prototype.visitCallSignature = function (node) {
    };
    NoShadowedVariableWalker.prototype.visitFunctionType = function (node) {
    };
    NoShadowedVariableWalker.prototype.visitConstructorType = function (node) {
    };
    NoShadowedVariableWalker.prototype.visitIndexSignatureDeclaration = function (node) {
    };
    NoShadowedVariableWalker.prototype.visitMethodSignature = function (node) {
    };
    NoShadowedVariableWalker.prototype.visitParameterDeclaration = function (node) {
        var isSingleParameter = node.name.kind === ts.SyntaxKind.Identifier;
        if (isSingleParameter) {
            this.handleSingleParameterIdentifier(node.name);
        }
        _super.prototype.visitParameterDeclaration.call(this, node);
    };
    NoShadowedVariableWalker.prototype.visitTypeLiteral = function (node) {
    };
    NoShadowedVariableWalker.prototype.visitVariableDeclaration = function (node) {
        var isSingleVariable = node.name.kind === ts.SyntaxKind.Identifier;
        if (isSingleVariable) {
            this.handleSingleVariableIdentifier(node.name, Lint.isBlockScopedVariable(node));
        }
        _super.prototype.visitVariableDeclaration.call(this, node);
    };
    NoShadowedVariableWalker.prototype.handleSingleVariableIdentifier = function (variableIdentifier, isBlockScoped) {
        var variableName = variableIdentifier.text;
        var currentScope = this.getCurrentScope();
        var currentBlockScope = this.getCurrentBlockScope();
        if (this.isVarInAnyScope(variableName) && currentBlockScope.varNames.indexOf(variableName) < 0) {
            this.addFailureOnIdentifier(variableIdentifier);
        }
        if (!isBlockScoped
            || this.getCurrentBlockDepth() === 1
            || this.getCurrentBlockDepth() === this.getCurrentDepth()) {
            currentScope.varNames.push(variableName);
        }
        currentBlockScope.varNames.push(variableName);
    };
    NoShadowedVariableWalker.prototype.handleSingleParameterIdentifier = function (variableIdentifier) {
        var variableName = variableIdentifier.text;
        var currentScope = this.getCurrentScope();
        if (this.isVarInAnyScope(variableName)) {
            this.addFailureOnIdentifier(variableIdentifier);
        }
        currentScope.varNames.push(variableName);
    };
    NoShadowedVariableWalker.prototype.isVarInAnyScope = function (varName) {
        return this.getAllScopes().some(function (scopeInfo) { return scopeInfo.varNames.indexOf(varName) >= 0; });
    };
    NoShadowedVariableWalker.prototype.addFailureOnIdentifier = function (ident) {
        var failureString = Rule.FAILURE_STRING_FACTORY(ident.text);
        this.addFailure(this.createFailure(ident.getStart(), ident.getWidth(), failureString));
    };
    return NoShadowedVariableWalker;
}(Lint.BlockScopeAwareRuleWalker));
var ScopeInfo = (function () {
    function ScopeInfo() {
        this.varNames = [];
    }
    return ScopeInfo;
}());
