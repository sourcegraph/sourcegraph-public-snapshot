"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ts = require("typescript");
var Lint = require("../lint");
var noUnusedExpressionRule_1 = require("./noUnusedExpressionRule");
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        return this.applyWithWalker(new NoUnusedNewWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-unused-new",
        description: "Disallows unused 'new' expression statements.",
        descriptionDetails: (_a = ["\n            Unused 'new' expressions indicate that a constructor is being invoked solely for its side effects."], _a.raw = ["\n            Unused 'new' expressions indicate that a constructor is being invoked solely for its side effects."], Lint.Utils.dedent(_a)),
        rationale: (_b = ["\n            Detects constructs such as `new SomeClass()`, where a constructor is used solely for its side effects, which is considered\n            poor style."], _b.raw = ["\n            Detects constructs such as \\`new SomeClass()\\`, where a constructor is used solely for its side effects, which is considered\n            poor style."], Lint.Utils.dedent(_b)),
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING = "do not use 'new' for side effects";
    return Rule;
    var _a, _b;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoUnusedNewWalker = (function (_super) {
    __extends(NoUnusedNewWalker, _super);
    function NoUnusedNewWalker() {
        _super.apply(this, arguments);
        this.expressionContainsNew = false;
    }
    NoUnusedNewWalker.prototype.visitExpressionStatement = function (node) {
        this.expressionContainsNew = false;
        _super.prototype.visitExpressionStatement.call(this, node);
    };
    NoUnusedNewWalker.prototype.visitNewExpression = function (node) {
        _super.prototype.visitNewExpression.call(this, node);
        this.expressionIsUnused = true;
        this.expressionContainsNew = true;
    };
    NoUnusedNewWalker.prototype.checkExpressionUsage = function (node) {
        if (this.expressionIsUnused && this.expressionContainsNew) {
            var expression = node.expression;
            var kind = expression.kind;
            var isValidStandaloneExpression = kind === ts.SyntaxKind.DeleteExpression
                || kind === ts.SyntaxKind.YieldExpression
                || kind === ts.SyntaxKind.AwaitExpression;
            if (!isValidStandaloneExpression && !noUnusedExpressionRule_1.NoUnusedExpressionWalker.isDirective(node)) {
                this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING));
            }
        }
    };
    return NoUnusedNewWalker;
}(noUnusedExpressionRule_1.NoUnusedExpressionWalker));
