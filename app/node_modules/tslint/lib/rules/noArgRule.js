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
        return this.applyWithWalker(new NoArgWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-arg",
        description: "Disallows use of `arguments.callee`.",
        rationale: (_a = ["\n            Using `arguments.callee` makes various performance optimizations impossible.\n            See [MDN](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Functions/arguments/callee)\n            for more details on why to avoid `arguments.callee`."], _a.raw = ["\n            Using \\`arguments.callee\\` makes various performance optimizations impossible.\n            See [MDN](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Functions/arguments/callee)\n            for more details on why to avoid \\`arguments.callee\\`."], Lint.Utils.dedent(_a)),
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING = "Access to arguments.callee is forbidden";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoArgWalker = (function (_super) {
    __extends(NoArgWalker, _super);
    function NoArgWalker() {
        _super.apply(this, arguments);
    }
    NoArgWalker.prototype.visitPropertyAccessExpression = function (node) {
        var expression = node.expression;
        var name = node.name;
        if (expression.kind === ts.SyntaxKind.Identifier && name.text === "callee") {
            var identifierExpression = expression;
            if (identifierExpression.text === "arguments") {
                this.addFailure(this.createFailure(expression.getStart(), expression.getWidth(), Rule.FAILURE_STRING));
            }
        }
        _super.prototype.visitPropertyAccessExpression.call(this, node);
    };
    return NoArgWalker;
}(Lint.RuleWalker));
