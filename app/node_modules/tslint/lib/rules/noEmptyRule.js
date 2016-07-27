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
        return this.applyWithWalker(new BlockWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-empty",
        description: "Disallows empty blocks.",
        descriptionDetails: "Blocks with a comment inside are not considered empty.",
        rationale: "Empty blocks are often indicators of missing code.",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING = "block is empty";
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var BlockWalker = (function (_super) {
    __extends(BlockWalker, _super);
    function BlockWalker() {
        _super.apply(this, arguments);
        this.ignoredBlocks = [];
    }
    BlockWalker.prototype.visitBlock = function (node) {
        var openBrace = node.getChildAt(0);
        var closeBrace = node.getChildAt(node.getChildCount() - 1);
        var sourceFileText = node.getSourceFile().text;
        var hasCommentAfter = ts.getTrailingCommentRanges(sourceFileText, openBrace.getEnd()) != null;
        var hasCommentBefore = ts.getLeadingCommentRanges(sourceFileText, closeBrace.getFullStart()) != null;
        var isSkipped = this.ignoredBlocks.indexOf(node) !== -1;
        if (node.statements.length <= 0 && !hasCommentAfter && !hasCommentBefore && !isSkipped) {
            this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING));
        }
        _super.prototype.visitBlock.call(this, node);
    };
    BlockWalker.prototype.visitConstructorDeclaration = function (node) {
        var parameters = node.parameters;
        var isSkipped = false;
        for (var _i = 0, parameters_1 = parameters; _i < parameters_1.length; _i++) {
            var param = parameters_1[_i];
            var hasPropertyAccessModifier = Lint.hasModifier(param.modifiers, ts.SyntaxKind.PrivateKeyword, ts.SyntaxKind.ProtectedKeyword, ts.SyntaxKind.PublicKeyword);
            if (hasPropertyAccessModifier) {
                isSkipped = true;
                this.ignoredBlocks.push(node.body);
                break;
            }
            if (isSkipped) {
                break;
            }
        }
        _super.prototype.visitConstructorDeclaration.call(this, node);
    };
    return BlockWalker;
}(Lint.RuleWalker));
