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
        var walker = new NoConditionalAssignmentWalker(sourceFile, this.getOptions());
        return this.applyWithWalker(walker);
    };
    Rule.metadata = {
        ruleName: "no-conditional-assignment",
        description: "Disallows any type of assignment in conditionals.",
        descriptionDetails: "This applies to `do-while`, `for`, `if`, and `while` statements.",
        rationale: (_a = ["\n            Assignments in conditionals are often typos:\n            for example `if (var1 = var2)` instead of `if (var1 == var2)`.\n            They also can be an indicator of overly clever code which decreases maintainability."], _a.raw = ["\n            Assignments in conditionals are often typos:\n            for example \\`if (var1 = var2)\\` instead of \\`if (var1 == var2)\\`.\n            They also can be an indicator of overly clever code which decreases maintainability."], Lint.Utils.dedent(_a)),
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING = "Assignments in conditional expressions are forbidden";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoConditionalAssignmentWalker = (function (_super) {
    __extends(NoConditionalAssignmentWalker, _super);
    function NoConditionalAssignmentWalker() {
        _super.apply(this, arguments);
        this.isInConditional = false;
    }
    NoConditionalAssignmentWalker.prototype.visitIfStatement = function (node) {
        this.validateConditionalExpression(node.expression);
        _super.prototype.visitIfStatement.call(this, node);
    };
    NoConditionalAssignmentWalker.prototype.visitWhileStatement = function (node) {
        this.validateConditionalExpression(node.expression);
        _super.prototype.visitWhileStatement.call(this, node);
    };
    NoConditionalAssignmentWalker.prototype.visitDoStatement = function (node) {
        this.validateConditionalExpression(node.expression);
        _super.prototype.visitWhileStatement.call(this, node);
    };
    NoConditionalAssignmentWalker.prototype.visitForStatement = function (node) {
        if (node.condition != null) {
            this.validateConditionalExpression(node.condition);
        }
        _super.prototype.visitForStatement.call(this, node);
    };
    NoConditionalAssignmentWalker.prototype.visitBinaryExpression = function (expression) {
        if (this.isInConditional) {
            this.checkForAssignment(expression);
        }
        _super.prototype.visitBinaryExpression.call(this, expression);
    };
    NoConditionalAssignmentWalker.prototype.validateConditionalExpression = function (expression) {
        this.isInConditional = true;
        if (expression.kind === ts.SyntaxKind.BinaryExpression) {
            this.checkForAssignment(expression);
        }
        this.walkChildren(expression);
        this.isInConditional = false;
    };
    NoConditionalAssignmentWalker.prototype.checkForAssignment = function (expression) {
        if (isAssignmentToken(expression.operatorToken)) {
            this.addFailure(this.createFailure(expression.getStart(), expression.getWidth(), Rule.FAILURE_STRING));
        }
    };
    return NoConditionalAssignmentWalker;
}(Lint.RuleWalker));
function isAssignmentToken(token) {
    return token.kind >= ts.SyntaxKind.FirstAssignment && token.kind <= ts.SyntaxKind.LastAssignment;
}
