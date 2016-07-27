"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var Lint = require("../lint");
var ts = require("typescript");
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        return this.applyWithWalker(new NoReturnInFinallyScopeAwareWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-unsafe-finally",
        description: (_a = ["\n            Disallows control flow statements, such as `return`, `continue`,\n            `break` and `throws` in finally blocks."], _a.raw = ["\n            Disallows control flow statements, such as \\`return\\`, \\`continue\\`,\n            \\`break\\` and \\`throws\\` in finally blocks."], Lint.Utils.dedent(_a)),
        descriptionDetails: "",
        rationale: (_b = ["\n            When used inside `finally` blocks, control flow statements,\n            such as `return`, `continue`, `break` and `throws`\n            override any other control flow statements in the same try/catch scope.\n            This is confusing and unexpected behavior."], _b.raw = ["\n            When used inside \\`finally\\` blocks, control flow statements,\n            such as \\`return\\`, \\`continue\\`, \\`break\\` and \\`throws\\`\n            override any other control flow statements in the same try/catch scope.\n            This is confusing and unexpected behavior."], Lint.Utils.dedent(_b)),
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_TYPE_BREAK = "break";
    Rule.FAILURE_TYPE_CONTINUE = "continue";
    Rule.FAILURE_TYPE_RETURN = "return";
    Rule.FAILURE_TYPE_THROW = "throw";
    Rule.FAILURE_STRING_FACTORY = function (name) { return (name + " statements in finally blocks are forbidden."); };
    return Rule;
    var _a, _b;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoReturnInFinallyScopeAwareWalker = (function (_super) {
    __extends(NoReturnInFinallyScopeAwareWalker, _super);
    function NoReturnInFinallyScopeAwareWalker() {
        _super.apply(this, arguments);
    }
    NoReturnInFinallyScopeAwareWalker.prototype.visitBreakStatement = function (node) {
        if (!this.isControlFlowWithinFinallyBlock(isBreakBoundary, node)) {
            _super.prototype.visitBreakStatement.call(this, node);
            return;
        }
        this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING_FACTORY(Rule.FAILURE_TYPE_BREAK)));
    };
    NoReturnInFinallyScopeAwareWalker.prototype.visitContinueStatement = function (node) {
        if (!this.isControlFlowWithinFinallyBlock(isContinueBoundary, node)) {
            _super.prototype.visitContinueStatement.call(this, node);
            return;
        }
        this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING_FACTORY(Rule.FAILURE_TYPE_CONTINUE)));
    };
    NoReturnInFinallyScopeAwareWalker.prototype.visitLabeledStatement = function (node) {
        this.getCurrentScope().labels.push(node.label.text);
        _super.prototype.visitLabeledStatement.call(this, node);
    };
    NoReturnInFinallyScopeAwareWalker.prototype.visitReturnStatement = function (node) {
        if (!this.isControlFlowWithinFinallyBlock(isReturnsOrThrowsBoundary)) {
            _super.prototype.visitReturnStatement.call(this, node);
            return;
        }
        this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING_FACTORY(Rule.FAILURE_TYPE_RETURN)));
    };
    NoReturnInFinallyScopeAwareWalker.prototype.visitThrowStatement = function (node) {
        if (!this.isControlFlowWithinFinallyBlock(isReturnsOrThrowsBoundary)) {
            _super.prototype.visitThrowStatement.call(this, node);
            return;
        }
        this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING_FACTORY(Rule.FAILURE_TYPE_THROW)));
    };
    NoReturnInFinallyScopeAwareWalker.prototype.createScope = function (node) {
        var isScopeBoundary = _super.prototype.isScopeBoundary.call(this, node);
        return {
            isBreakBoundary: isScopeBoundary || isLoopBlock(node) || isCaseBlock(node),
            isContinueBoundary: isScopeBoundary || isLoopBlock(node),
            isFinallyBlock: isFinallyBlock(node),
            isReturnsThrowsBoundary: isScopeBoundary,
            labels: [],
        };
    };
    NoReturnInFinallyScopeAwareWalker.prototype.isScopeBoundary = function (node) {
        return _super.prototype.isScopeBoundary.call(this, node) ||
            isFinallyBlock(node) ||
            isLoopBlock(node) ||
            isCaseBlock(node);
    };
    NoReturnInFinallyScopeAwareWalker.prototype.isControlFlowWithinFinallyBlock = function (isControlFlowBoundary, node) {
        var scopes = this.getAllScopes();
        var currentScope = this.getCurrentScope();
        var depth = this.getCurrentDepth();
        while (currentScope) {
            if (isControlFlowBoundary(currentScope, node)) {
                return false;
            }
            if (currentScope.isFinallyBlock) {
                return true;
            }
            currentScope = scopes[--depth];
        }
    };
    return NoReturnInFinallyScopeAwareWalker;
}(Lint.ScopeAwareRuleWalker));
function isLoopBlock(node) {
    var parent = node.parent;
    return parent &&
        node.kind === ts.SyntaxKind.Block &&
        (parent.kind === ts.SyntaxKind.ForInStatement ||
            parent.kind === ts.SyntaxKind.ForOfStatement ||
            parent.kind === ts.SyntaxKind.ForStatement ||
            parent.kind === ts.SyntaxKind.WhileStatement ||
            parent.kind === ts.SyntaxKind.DoStatement);
}
function isCaseBlock(node) {
    return node.kind === ts.SyntaxKind.CaseBlock;
}
function isFinallyBlock(node) {
    var parent = node.parent;
    return parent &&
        node.kind === ts.SyntaxKind.Block &&
        isTryStatement(parent) &&
        parent.finallyBlock === node;
}
function isTryStatement(node) {
    return node.kind === ts.SyntaxKind.TryStatement;
}
function isReturnsOrThrowsBoundary(scope) {
    return scope.isReturnsThrowsBoundary;
}
function isContinueBoundary(scope, node) {
    return node.label ? scope.labels.indexOf(node.label.text) >= 0 : scope.isContinueBoundary;
}
function isBreakBoundary(scope, node) {
    return node.label ? scope.labels.indexOf(node.label.text) >= 0 : scope.isBreakBoundary;
}
