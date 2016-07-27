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
        return this.applyWithWalker(new NoEvalWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-eval",
        description: "Disallows `eval` function invocations.",
        rationale: (_a = ["\n            `eval()` is dangerous as it allows arbitrary code execution with full privileges. There are\n            [alternatives](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/eval)\n            for most of the use cases for `eval()`."], _a.raw = ["\n            \\`eval()\\` is dangerous as it allows arbitrary code execution with full privileges. There are\n            [alternatives](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/eval)\n            for most of the use cases for \\`eval()\\`."], Lint.Utils.dedent(_a)),
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING = "forbidden eval";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoEvalWalker = (function (_super) {
    __extends(NoEvalWalker, _super);
    function NoEvalWalker() {
        _super.apply(this, arguments);
    }
    NoEvalWalker.prototype.visitCallExpression = function (node) {
        var expression = node.expression;
        if (expression.kind === ts.SyntaxKind.Identifier) {
            var expressionName = expression.text;
            if (expressionName === "eval") {
                this.addFailure(this.createFailure(expression.getStart(), expression.getWidth(), Rule.FAILURE_STRING));
            }
        }
        _super.prototype.visitCallExpression.call(this, node);
    };
    return NoEvalWalker;
}(Lint.RuleWalker));
