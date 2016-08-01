"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var Lint = require("../lint");
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        return this.applyWithWalker(new NoAngleBracketTypeAssertionWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-angle-bracket-type-assertion",
        description: "Requires the use of `as Type` for type assertions instead of `<Type>`.",
        rationale: (_a = ["\n            Both formats of type assertions have the same effect, but only `as` type assertions\n            work in `.tsx` files. This rule ensures that you have a consistent type assertion style\n            across your codebase."], _a.raw = ["\n            Both formats of type assertions have the same effect, but only \\`as\\` type assertions\n            work in \\`.tsx\\` files. This rule ensures that you have a consistent type assertion style\n            across your codebase."], Lint.Utils.dedent(_a)),
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "style",
    };
    Rule.FAILURE_STRING = "Type assertion using the '<>' syntax is forbidden. Use the 'as' syntax instead.";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoAngleBracketTypeAssertionWalker = (function (_super) {
    __extends(NoAngleBracketTypeAssertionWalker, _super);
    function NoAngleBracketTypeAssertionWalker() {
        _super.apply(this, arguments);
    }
    NoAngleBracketTypeAssertionWalker.prototype.visitTypeAssertionExpression = function (node) {
        this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING));
        _super.prototype.visitTypeAssertionExpression.call(this, node);
    };
    return NoAngleBracketTypeAssertionWalker;
}(Lint.RuleWalker));
