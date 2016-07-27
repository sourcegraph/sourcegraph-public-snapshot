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
        var alignWalker = new AlignWalker(sourceFile, this.getOptions());
        return this.applyWithWalker(alignWalker);
    };
    Rule.metadata = {
        ruleName: "align",
        description: "Enforces vertical alignment.",
        rationale: "Helps maintain a readable, consistent style in your codebase.",
        optionsDescription: (_a = ["\n            Three arguments may be optionally provided:\n\n            * `\"parameters\"` checks alignment of function parameters.\n            * `\"arguments\"` checks alignment of function call arguments.\n            * `\"statements\"` checks alignment of statements."], _a.raw = ["\n            Three arguments may be optionally provided:\n\n            * \\`\"parameters\"\\` checks alignment of function parameters.\n            * \\`\"arguments\"\\` checks alignment of function call arguments.\n            * \\`\"statements\"\\` checks alignment of statements."], Lint.Utils.dedent(_a)),
        options: {
            type: "array",
            items: {
                type: "string",
                enum: ["arguments", "parameters", "statements"],
            },
            minLength: 1,
            maxLength: 3,
        },
        optionExamples: ['[true, "parameters", "statements"]'],
        type: "style",
    };
    Rule.PARAMETERS_OPTION = "parameters";
    Rule.ARGUMENTS_OPTION = "arguments";
    Rule.STATEMENTS_OPTION = "statements";
    Rule.FAILURE_STRING_SUFFIX = " are not aligned";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var AlignWalker = (function (_super) {
    __extends(AlignWalker, _super);
    function AlignWalker() {
        _super.apply(this, arguments);
    }
    AlignWalker.prototype.visitConstructorDeclaration = function (node) {
        this.checkAlignment(Rule.PARAMETERS_OPTION, node.parameters);
        _super.prototype.visitConstructorDeclaration.call(this, node);
    };
    AlignWalker.prototype.visitFunctionDeclaration = function (node) {
        this.checkAlignment(Rule.PARAMETERS_OPTION, node.parameters);
        _super.prototype.visitFunctionDeclaration.call(this, node);
    };
    AlignWalker.prototype.visitFunctionExpression = function (node) {
        this.checkAlignment(Rule.PARAMETERS_OPTION, node.parameters);
        _super.prototype.visitFunctionExpression.call(this, node);
    };
    AlignWalker.prototype.visitMethodDeclaration = function (node) {
        this.checkAlignment(Rule.PARAMETERS_OPTION, node.parameters);
        _super.prototype.visitMethodDeclaration.call(this, node);
    };
    AlignWalker.prototype.visitCallExpression = function (node) {
        this.checkAlignment(Rule.ARGUMENTS_OPTION, node.arguments);
        _super.prototype.visitCallExpression.call(this, node);
    };
    AlignWalker.prototype.visitNewExpression = function (node) {
        this.checkAlignment(Rule.ARGUMENTS_OPTION, node.arguments);
        _super.prototype.visitNewExpression.call(this, node);
    };
    AlignWalker.prototype.visitBlock = function (node) {
        this.checkAlignment(Rule.STATEMENTS_OPTION, node.statements);
        _super.prototype.visitBlock.call(this, node);
    };
    AlignWalker.prototype.checkAlignment = function (kind, nodes) {
        if (nodes == null || nodes.length === 0 || !this.hasOption(kind)) {
            return;
        }
        var prevPos = getPosition(nodes[0]);
        var alignToColumn = prevPos.character;
        for (var _i = 0, _a = nodes.slice(1); _i < _a.length; _i++) {
            var node = _a[_i];
            var curPos = getPosition(node);
            if (curPos.line !== prevPos.line && curPos.character !== alignToColumn) {
                this.addFailure(this.createFailure(node.getStart(), node.getWidth(), kind + Rule.FAILURE_STRING_SUFFIX));
                break;
            }
            prevPos = curPos;
        }
    };
    return AlignWalker;
}(Lint.RuleWalker));
function getPosition(node) {
    return node.getSourceFile().getLineAndCharacterOfPosition(node.getStart());
}
