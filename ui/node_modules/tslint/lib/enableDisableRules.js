"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ts = require("typescript");
var utils_1 = require("./language/utils");
var skippableTokenAwareRuleWalker_1 = require("./language/walker/skippableTokenAwareRuleWalker");
var EnableDisableRulesWalker = (function (_super) {
    __extends(EnableDisableRulesWalker, _super);
    function EnableDisableRulesWalker() {
        _super.apply(this, arguments);
        this.enableDisableRuleMap = {};
    }
    EnableDisableRulesWalker.prototype.visitSourceFile = function (node) {
        var _this = this;
        _super.prototype.visitSourceFile.call(this, node);
        var scan = ts.createScanner(ts.ScriptTarget.ES5, false, ts.LanguageVariant.Standard, node.text);
        utils_1.scanAllTokens(scan, function (scanner) {
            var startPos = scanner.getStartPos();
            if (_this.tokensToSkipStartEndMap[startPos] != null) {
                scanner.setTextPos(_this.tokensToSkipStartEndMap[startPos]);
                return;
            }
            if (scanner.getToken() === ts.SyntaxKind.MultiLineCommentTrivia ||
                scanner.getToken() === ts.SyntaxKind.SingleLineCommentTrivia) {
                var commentText = scanner.getTokenText();
                var startPosition = scanner.getTokenPos();
                _this.handlePossibleTslintSwitch(commentText, startPosition, node, scanner);
            }
        });
    };
    EnableDisableRulesWalker.prototype.getStartOfLinePosition = function (node, position, lineOffset) {
        if (lineOffset === void 0) { lineOffset = 0; }
        return node.getPositionOfLineAndCharacter(node.getLineAndCharacterOfPosition(position).line + lineOffset, 0);
    };
    EnableDisableRulesWalker.prototype.handlePossibleTslintSwitch = function (commentText, startingPosition, node, scanner) {
        if (commentText.match(/^(\/\*|\/\/)\s*tslint:/)) {
            var commentTextParts = commentText.split(":");
            var enableOrDisableMatch = commentTextParts[1].match(/^(enable|disable)(-(line|next-line))?(\s|$)/);
            if (enableOrDisableMatch != null) {
                var isEnabled = enableOrDisableMatch[1] === "enable";
                var isCurrentLine = enableOrDisableMatch[3] === "line";
                var isNextLine = enableOrDisableMatch[3] === "next-line";
                var rulesList = ["all"];
                if (commentTextParts.length === 2) {
                    rulesList = commentTextParts[1].split(/\s+/).slice(1);
                    rulesList = rulesList.filter(function (item) { return !!item && item.indexOf("*/") === -1; });
                    rulesList = rulesList.length > 0 ? rulesList : ["all"];
                }
                else if (commentTextParts.length > 2) {
                    rulesList = commentTextParts[2].split(/\s+/);
                }
                for (var _i = 0, rulesList_1 = rulesList; _i < rulesList_1.length; _i++) {
                    var ruleToAdd = rulesList_1[_i];
                    if (!(ruleToAdd in this.enableDisableRuleMap)) {
                        this.enableDisableRuleMap[ruleToAdd] = [];
                    }
                    if (isCurrentLine) {
                        this.enableDisableRuleMap[ruleToAdd].push({
                            isEnabled: isEnabled,
                            position: this.getStartOfLinePosition(node, startingPosition),
                        });
                        this.enableDisableRuleMap[ruleToAdd].push({
                            isEnabled: !isEnabled,
                            position: scanner.getTextPos() + 1,
                        });
                    }
                    else {
                        this.enableDisableRuleMap[ruleToAdd].push({
                            isEnabled: isEnabled,
                            position: startingPosition,
                        });
                        if (isNextLine) {
                            this.enableDisableRuleMap[ruleToAdd].push({
                                isEnabled: !isEnabled,
                                position: this.getStartOfLinePosition(node, startingPosition, 2),
                            });
                        }
                    }
                }
            }
        }
    };
    return EnableDisableRulesWalker;
}(skippableTokenAwareRuleWalker_1.SkippableTokenAwareRuleWalker));
exports.EnableDisableRulesWalker = EnableDisableRulesWalker;
