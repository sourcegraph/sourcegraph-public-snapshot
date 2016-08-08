"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ts = require("typescript");
var Lint = require("../lint");
var OPTION_ALLOW_NULL_CHECK = "allow-null-check";
var OPTION_ALLOW_UNDEFINED_CHECK = "allow-undefined-check";
function isUndefinedExpression(expression) {
    return expression.kind === ts.SyntaxKind.Identifier && expression.getText() === "undefined";
}
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        var comparisonWalker = new ComparisonWalker(sourceFile, this.getOptions());
        return this.applyWithWalker(comparisonWalker);
    };
    Rule.metadata = {
        ruleName: "triple-equals",
        description: "Requires `===` and `!==` in place of `==` and `!=`.",
        optionsDescription: (_a = ["\n            Two arguments may be optionally provided:\n\n            * `\"allow-null-check\"` allows `==` and `!=` when comparing to `null`.\n            * `\"allow-undefined-check\"` allows `==` and `!=` when comparing to `undefined`."], _a.raw = ["\n            Two arguments may be optionally provided:\n\n            * \\`\"allow-null-check\"\\` allows \\`==\\` and \\`!=\\` when comparing to \\`null\\`.\n            * \\`\"allow-undefined-check\"\\` allows \\`==\\` and \\`!=\\` when comparing to \\`undefined\\`."], Lint.Utils.dedent(_a)),
        options: {
            type: "array",
            items: {
                type: "string",
                enum: [OPTION_ALLOW_NULL_CHECK, OPTION_ALLOW_UNDEFINED_CHECK],
            },
            minLength: 0,
            maxLength: 2,
        },
        optionExamples: ["true", '[true, "allow-null-check"]', '[true, "allow-undefined-check"]'],
        type: "functionality",
    };
    Rule.EQ_FAILURE_STRING = "== should be ===";
    Rule.NEQ_FAILURE_STRING = "!= should be !==";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var ComparisonWalker = (function (_super) {
    __extends(ComparisonWalker, _super);
    function ComparisonWalker() {
        _super.apply(this, arguments);
    }
    ComparisonWalker.prototype.visitBinaryExpression = function (node) {
        if (!this.isExpressionAllowed(node)) {
            var position = node.getChildAt(1).getStart();
            this.handleOperatorToken(position, node.operatorToken.kind);
        }
        _super.prototype.visitBinaryExpression.call(this, node);
    };
    ComparisonWalker.prototype.handleOperatorToken = function (position, operator) {
        switch (operator) {
            case ts.SyntaxKind.EqualsEqualsToken:
                this.addFailure(this.createFailure(position, ComparisonWalker.COMPARISON_OPERATOR_WIDTH, Rule.EQ_FAILURE_STRING));
                break;
            case ts.SyntaxKind.ExclamationEqualsToken:
                this.addFailure(this.createFailure(position, ComparisonWalker.COMPARISON_OPERATOR_WIDTH, Rule.NEQ_FAILURE_STRING));
                break;
            default:
                break;
        }
    };
    ComparisonWalker.prototype.isExpressionAllowed = function (node) {
        var nullKeyword = ts.SyntaxKind.NullKeyword;
        return (this.hasOption(OPTION_ALLOW_NULL_CHECK) && (node.left.kind === nullKeyword || node.right.kind === nullKeyword)) || (this.hasOption(OPTION_ALLOW_UNDEFINED_CHECK) && (isUndefinedExpression(node.left) || isUndefinedExpression(node.right)));
    };
    ComparisonWalker.COMPARISON_OPERATOR_WIDTH = 2;
    return ComparisonWalker;
}(Lint.RuleWalker));
