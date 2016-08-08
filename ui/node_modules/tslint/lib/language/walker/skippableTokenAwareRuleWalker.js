"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ruleWalker_1 = require("./ruleWalker");
var SkippableTokenAwareRuleWalker = (function (_super) {
    __extends(SkippableTokenAwareRuleWalker, _super);
    function SkippableTokenAwareRuleWalker(sourceFile, options) {
        _super.call(this, sourceFile, options);
        this.tokensToSkipStartEndMap = {};
    }
    SkippableTokenAwareRuleWalker.prototype.visitRegularExpressionLiteral = function (node) {
        this.addTokenToSkipFromNode(node);
        _super.prototype.visitRegularExpressionLiteral.call(this, node);
    };
    SkippableTokenAwareRuleWalker.prototype.visitIdentifier = function (node) {
        this.addTokenToSkipFromNode(node);
        _super.prototype.visitIdentifier.call(this, node);
    };
    SkippableTokenAwareRuleWalker.prototype.visitTemplateExpression = function (node) {
        this.addTokenToSkipFromNode(node);
        _super.prototype.visitTemplateExpression.call(this, node);
    };
    SkippableTokenAwareRuleWalker.prototype.addTokenToSkipFromNode = function (node) {
        if (node.getStart() < node.getEnd()) {
            this.tokensToSkipStartEndMap[node.getStart()] = node.getEnd();
        }
    };
    return SkippableTokenAwareRuleWalker;
}(ruleWalker_1.RuleWalker));
exports.SkippableTokenAwareRuleWalker = SkippableTokenAwareRuleWalker;
