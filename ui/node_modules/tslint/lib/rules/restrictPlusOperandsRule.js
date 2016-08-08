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
    Rule.prototype.applyWithProgram = function (sourceFile, program) {
        return this.applyWithWalker(new RestrictPlusOperandsWalker(sourceFile, this.getOptions(), program));
    };
    Rule.metadata = {
        ruleName: "restrict-plus-operands",
        description: "When adding two variables, operands must both be of type number or of type string.",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
        requiresTypeInfo: true,
    };
    Rule.MISMATCHED_TYPES_FAILURE = "Types of values used in '+' operation must match";
    Rule.UNSUPPORTED_TYPE_FAILURE_FACTORY = function (type) { return ("cannot add type " + type); };
    return Rule;
}(Lint.Rules.TypedRule));
exports.Rule = Rule;
var RestrictPlusOperandsWalker = (function (_super) {
    __extends(RestrictPlusOperandsWalker, _super);
    function RestrictPlusOperandsWalker() {
        _super.apply(this, arguments);
    }
    RestrictPlusOperandsWalker.prototype.visitBinaryExpression = function (node) {
        if (node.operatorToken.kind === ts.SyntaxKind.PlusToken) {
            var tc = this.getTypeChecker();
            var leftType = tc.typeToString(tc.getTypeAtLocation(node.left));
            var rightType = tc.typeToString(tc.getTypeAtLocation(node.right));
            var width = node.getWidth();
            var position = node.getStart();
            if (leftType !== rightType) {
                this.addFailure(this.createFailure(position, width, Rule.MISMATCHED_TYPES_FAILURE));
            }
            else if (leftType !== "number" && leftType !== "string") {
                var failureString = Rule.UNSUPPORTED_TYPE_FAILURE_FACTORY(leftType);
                this.addFailure(this.createFailure(position, width, failureString));
            }
        }
        _super.prototype.visitBinaryExpression.call(this, node);
    };
    return RestrictPlusOperandsWalker;
}(Lint.ProgramAwareRuleWalker));
