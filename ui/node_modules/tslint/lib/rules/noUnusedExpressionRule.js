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
        return this.applyWithWalker(new NoUnusedExpressionWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-unused-expression",
        description: "Disallows unused expression statements.",
        descriptionDetails: (_a = ["\n            Unused expressions are expression statements which are not assignments or function calls\n            (and thus usually no-ops)."], _a.raw = ["\n            Unused expressions are expression statements which are not assignments or function calls\n            (and thus usually no-ops)."], Lint.Utils.dedent(_a)),
        rationale: (_b = ["\n            Detects potential errors where an assignment or function call was intended."], _b.raw = ["\n            Detects potential errors where an assignment or function call was intended."], Lint.Utils.dedent(_b)),
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING = "expected an assignment or function call";
    return Rule;
    var _a, _b;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoUnusedExpressionWalker = (function (_super) {
    __extends(NoUnusedExpressionWalker, _super);
    function NoUnusedExpressionWalker(sourceFile, options) {
        _super.call(this, sourceFile, options);
        this.expressionIsUnused = true;
    }
    NoUnusedExpressionWalker.isDirective = function (node, checkPreviousSiblings) {
        if (checkPreviousSiblings === void 0) { checkPreviousSiblings = true; }
        var parent = node.parent;
        var grandParentKind = parent.parent == null ? null : parent.parent.kind;
        var isStringExpression = node.kind === ts.SyntaxKind.ExpressionStatement
            && node.expression.kind === ts.SyntaxKind.StringLiteral;
        var parentIsSourceFile = parent.kind === ts.SyntaxKind.SourceFile;
        var parentIsNSBody = parent.kind === ts.SyntaxKind.ModuleBlock;
        var parentIsFunctionBody = parent.kind === ts.SyntaxKind.Block && [
            ts.SyntaxKind.ArrowFunction,
            ts.SyntaxKind.FunctionExpression,
            ts.SyntaxKind.FunctionDeclaration,
            ts.SyntaxKind.MethodDeclaration,
            ts.SyntaxKind.Constructor,
            ts.SyntaxKind.GetAccessor,
            ts.SyntaxKind.SetAccessor,
        ].indexOf(grandParentKind) > -1;
        if (!(parentIsSourceFile || parentIsFunctionBody || parentIsNSBody) || !isStringExpression) {
            return false;
        }
        if (checkPreviousSiblings) {
            var siblings_1 = [];
            ts.forEachChild(node.parent, function (child) { siblings_1.push(child); });
            return siblings_1.slice(0, siblings_1.indexOf(node)).every(function (n) { return NoUnusedExpressionWalker.isDirective(n, false); });
        }
        else {
            return true;
        }
    };
    NoUnusedExpressionWalker.prototype.visitExpressionStatement = function (node) {
        this.expressionIsUnused = true;
        _super.prototype.visitExpressionStatement.call(this, node);
        this.checkExpressionUsage(node);
    };
    NoUnusedExpressionWalker.prototype.visitBinaryExpression = function (node) {
        _super.prototype.visitBinaryExpression.call(this, node);
        switch (node.operatorToken.kind) {
            case ts.SyntaxKind.EqualsToken:
            case ts.SyntaxKind.PlusEqualsToken:
            case ts.SyntaxKind.MinusEqualsToken:
            case ts.SyntaxKind.AsteriskEqualsToken:
            case ts.SyntaxKind.SlashEqualsToken:
            case ts.SyntaxKind.PercentEqualsToken:
            case ts.SyntaxKind.AmpersandEqualsToken:
            case ts.SyntaxKind.CaretEqualsToken:
            case ts.SyntaxKind.BarEqualsToken:
            case ts.SyntaxKind.LessThanLessThanEqualsToken:
            case ts.SyntaxKind.GreaterThanGreaterThanEqualsToken:
            case ts.SyntaxKind.GreaterThanGreaterThanGreaterThanEqualsToken:
                this.expressionIsUnused = false;
                break;
            default:
                this.expressionIsUnused = true;
        }
    };
    NoUnusedExpressionWalker.prototype.visitPrefixUnaryExpression = function (node) {
        _super.prototype.visitPrefixUnaryExpression.call(this, node);
        switch (node.operator) {
            case ts.SyntaxKind.PlusPlusToken:
            case ts.SyntaxKind.MinusMinusToken:
                this.expressionIsUnused = false;
                break;
            default:
                this.expressionIsUnused = true;
        }
    };
    NoUnusedExpressionWalker.prototype.visitPostfixUnaryExpression = function (node) {
        _super.prototype.visitPostfixUnaryExpression.call(this, node);
        this.expressionIsUnused = false;
    };
    NoUnusedExpressionWalker.prototype.visitBlock = function (node) {
        _super.prototype.visitBlock.call(this, node);
        this.expressionIsUnused = true;
    };
    NoUnusedExpressionWalker.prototype.visitArrowFunction = function (node) {
        _super.prototype.visitArrowFunction.call(this, node);
        this.expressionIsUnused = true;
    };
    NoUnusedExpressionWalker.prototype.visitCallExpression = function (node) {
        _super.prototype.visitCallExpression.call(this, node);
        this.expressionIsUnused = false;
    };
    NoUnusedExpressionWalker.prototype.visitNewExpression = function (node) {
        _super.prototype.visitNewExpression.call(this, node);
        this.expressionIsUnused = false;
    };
    NoUnusedExpressionWalker.prototype.visitConditionalExpression = function (node) {
        this.visitNode(node.condition);
        this.expressionIsUnused = true;
        this.visitNode(node.whenTrue);
        var firstExpressionIsUnused = this.expressionIsUnused;
        this.expressionIsUnused = true;
        this.visitNode(node.whenFalse);
        var secondExpressionIsUnused = this.expressionIsUnused;
        this.expressionIsUnused = firstExpressionIsUnused || secondExpressionIsUnused;
    };
    NoUnusedExpressionWalker.prototype.checkExpressionUsage = function (node) {
        if (this.expressionIsUnused) {
            var expression = node.expression;
            var kind = expression.kind;
            var isValidStandaloneExpression = kind === ts.SyntaxKind.DeleteExpression
                || kind === ts.SyntaxKind.YieldExpression
                || kind === ts.SyntaxKind.AwaitExpression;
            if (!isValidStandaloneExpression && !NoUnusedExpressionWalker.isDirective(node)) {
                this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING));
            }
        }
    };
    return NoUnusedExpressionWalker;
}(Lint.RuleWalker));
exports.NoUnusedExpressionWalker = NoUnusedExpressionWalker;
