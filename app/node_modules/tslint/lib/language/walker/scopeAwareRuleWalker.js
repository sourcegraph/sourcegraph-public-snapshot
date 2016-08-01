"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ts = require("typescript");
var ruleWalker_1 = require("./ruleWalker");
var ScopeAwareRuleWalker = (function (_super) {
    __extends(ScopeAwareRuleWalker, _super);
    function ScopeAwareRuleWalker(sourceFile, options) {
        _super.call(this, sourceFile, options);
        this.scopeStack = [this.createScope(sourceFile)];
    }
    ScopeAwareRuleWalker.prototype.getCurrentScope = function () {
        return this.scopeStack[this.scopeStack.length - 1];
    };
    ScopeAwareRuleWalker.prototype.getAllScopes = function () {
        return this.scopeStack.slice();
    };
    ScopeAwareRuleWalker.prototype.getCurrentDepth = function () {
        return this.scopeStack.length;
    };
    ScopeAwareRuleWalker.prototype.onScopeStart = function () {
        return;
    };
    ScopeAwareRuleWalker.prototype.onScopeEnd = function () {
        return;
    };
    ScopeAwareRuleWalker.prototype.visitNode = function (node) {
        var isNewScope = this.isScopeBoundary(node);
        if (isNewScope) {
            this.scopeStack.push(this.createScope(node));
        }
        this.onScopeStart();
        _super.prototype.visitNode.call(this, node);
        this.onScopeEnd();
        if (isNewScope) {
            this.scopeStack.pop();
        }
    };
    ScopeAwareRuleWalker.prototype.isScopeBoundary = function (node) {
        return node.kind === ts.SyntaxKind.FunctionDeclaration
            || node.kind === ts.SyntaxKind.FunctionExpression
            || node.kind === ts.SyntaxKind.PropertyAssignment
            || node.kind === ts.SyntaxKind.ShorthandPropertyAssignment
            || node.kind === ts.SyntaxKind.MethodDeclaration
            || node.kind === ts.SyntaxKind.Constructor
            || node.kind === ts.SyntaxKind.ModuleDeclaration
            || node.kind === ts.SyntaxKind.ArrowFunction
            || node.kind === ts.SyntaxKind.ParenthesizedExpression
            || node.kind === ts.SyntaxKind.ClassDeclaration
            || node.kind === ts.SyntaxKind.ClassExpression
            || node.kind === ts.SyntaxKind.InterfaceDeclaration
            || node.kind === ts.SyntaxKind.GetAccessor
            || node.kind === ts.SyntaxKind.SetAccessor;
    };
    return ScopeAwareRuleWalker;
}(ruleWalker_1.RuleWalker));
exports.ScopeAwareRuleWalker = ScopeAwareRuleWalker;
