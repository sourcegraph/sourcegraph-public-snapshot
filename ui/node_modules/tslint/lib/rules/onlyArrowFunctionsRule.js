"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var Lint = require("../lint");
var OPTION_ALLOW_DECLARATIONS = "allow-declarations";
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        return this.applyWithWalker(new OnlyArrowFunctionsWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "only-arrow-functions",
        description: "Disallows traditional (non-arrow) function expressions.",
        rationale: "Traditional functions don't bind lexical scope, which can lead to unexpected behavior when accessing 'this'.",
        optionsDescription: (_a = ["\n            One argument may be optionally provided:\n\n            * `\"", "\"` allows standalone function declarations.\n        "], _a.raw = ["\n            One argument may be optionally provided:\n\n            * \\`\"", "\"\\` allows standalone function declarations.\n        "], Lint.Utils.dedent(_a, OPTION_ALLOW_DECLARATIONS)),
        options: {
            type: "array",
            items: {
                type: "string",
                enum: [OPTION_ALLOW_DECLARATIONS],
            },
            minLength: 0,
            maxLength: 1,
        },
        optionExamples: ["true", ("[true, \"" + OPTION_ALLOW_DECLARATIONS + "\"]")],
        type: "typescript",
    };
    Rule.FAILURE_STRING = "non-arrow functions are forbidden";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var OnlyArrowFunctionsWalker = (function (_super) {
    __extends(OnlyArrowFunctionsWalker, _super);
    function OnlyArrowFunctionsWalker() {
        _super.apply(this, arguments);
    }
    OnlyArrowFunctionsWalker.prototype.visitFunctionDeclaration = function (node) {
        if (!this.hasOption(OPTION_ALLOW_DECLARATIONS)) {
            this.addFailure(this.createFailure(node.getStart(), "function".length, Rule.FAILURE_STRING));
        }
        _super.prototype.visitFunctionDeclaration.call(this, node);
    };
    OnlyArrowFunctionsWalker.prototype.visitFunctionExpression = function (node) {
        this.addFailure(this.createFailure(node.getStart(), "function".length, Rule.FAILURE_STRING));
        _super.prototype.visitFunctionExpression.call(this, node);
    };
    return OnlyArrowFunctionsWalker;
}(Lint.RuleWalker));
