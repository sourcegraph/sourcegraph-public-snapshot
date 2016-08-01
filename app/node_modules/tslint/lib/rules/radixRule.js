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
        var radixWalker = new RadixWalker(sourceFile, this.getOptions());
        return this.applyWithWalker(radixWalker);
    };
    Rule.metadata = {
        ruleName: "radix",
        description: "Requires the radix parameter to be specified when calling `parseInt`.",
        rationale: (_a = ["\n            From [MDN](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/parseInt): \n            > Always specify this parameter to eliminate reader confusion and to guarantee predictable behavior.\n            > Different implementations produce different results when a radix is not specified, usually defaulting the value to 10."], _a.raw = ["\n            From [MDN](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/parseInt): \n            > Always specify this parameter to eliminate reader confusion and to guarantee predictable behavior.\n            > Different implementations produce different results when a radix is not specified, usually defaulting the value to 10."], Lint.Utils.dedent(_a)),
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING = "Missing radix parameter";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var RadixWalker = (function (_super) {
    __extends(RadixWalker, _super);
    function RadixWalker() {
        _super.apply(this, arguments);
    }
    RadixWalker.prototype.visitCallExpression = function (node) {
        var expression = node.expression;
        if (expression.kind === ts.SyntaxKind.Identifier
            && node.getFirstToken().getText() === "parseInt"
            && node.arguments.length < 2) {
            this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING));
        }
        _super.prototype.visitCallExpression.call(this, node);
    };
    return RadixWalker;
}(Lint.RuleWalker));
