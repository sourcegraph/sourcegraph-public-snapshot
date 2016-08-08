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
        var options = this.getOptions();
        var banFunctionWalker = new BanFunctionWalker(sourceFile, options);
        var functionsToBan = options.ruleArguments;
        functionsToBan.forEach(function (f) { return banFunctionWalker.addBannedFunction(f); });
        return this.applyWithWalker(banFunctionWalker);
    };
    Rule.metadata = {
        ruleName: "ban",
        description: "Bans the use of specific functions.",
        descriptionDetails: "At this time, there is no way to disable global methods with this rule.",
        optionsDescription: "A list of `['object', 'method', 'optional explanation here']` which ban `object.method()`.",
        options: {
            type: "list",
            listType: {
                type: "array",
                arrayMembers: [
                    { type: "string" },
                    { type: "string" },
                    { type: "string" },
                ],
            },
        },
        optionExamples: ["[true, [\"someObject\", \"someFunction\"], [\"someObject\", \"otherFunction\", \"Optional explanation\"]]"],
        type: "functionality",
    };
    Rule.FAILURE_STRING_FACTORY = function (expression, messageAddition) {
        return "Calls to '" + expression + "' are not allowed." + (messageAddition ? " " + messageAddition : "");
    };
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var BanFunctionWalker = (function (_super) {
    __extends(BanFunctionWalker, _super);
    function BanFunctionWalker() {
        _super.apply(this, arguments);
        this.bannedFunctions = [];
    }
    BanFunctionWalker.prototype.addBannedFunction = function (bannedFunction) {
        this.bannedFunctions.push(bannedFunction);
    };
    BanFunctionWalker.prototype.visitCallExpression = function (node) {
        var expression = node.expression;
        if (expression.kind === ts.SyntaxKind.PropertyAccessExpression
            && expression.getChildCount() >= 3) {
            var firstToken = expression.getFirstToken();
            var firstChild = expression.getChildAt(0);
            var secondChild = expression.getChildAt(1);
            var thirdChild = expression.getChildAt(2);
            var rightSideExpression = thirdChild.getFullText();
            var leftSideExpression = void 0;
            if (firstChild.getChildCount() > 0) {
                leftSideExpression = firstChild.getLastToken().getText();
            }
            else {
                leftSideExpression = firstToken.getText();
            }
            if (secondChild.kind === ts.SyntaxKind.DotToken) {
                for (var _i = 0, _a = this.bannedFunctions; _i < _a.length; _i++) {
                    var bannedFunction = _a[_i];
                    if (leftSideExpression === bannedFunction[0] && rightSideExpression === bannedFunction[1]) {
                        var failure = this.createFailure(expression.getStart(), expression.getWidth(), Rule.FAILURE_STRING_FACTORY(leftSideExpression + "." + rightSideExpression, bannedFunction[2]));
                        this.addFailure(failure);
                    }
                }
            }
        }
        _super.prototype.visitCallExpression.call(this, node);
    };
    return BanFunctionWalker;
}(Lint.RuleWalker));
exports.BanFunctionWalker = BanFunctionWalker;
