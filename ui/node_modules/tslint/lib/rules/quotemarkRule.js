"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ts = require("typescript");
var Lint = require("../lint");
var QuoteMark;
(function (QuoteMark) {
    QuoteMark[QuoteMark["SINGLE_QUOTES"] = 0] = "SINGLE_QUOTES";
    QuoteMark[QuoteMark["DOUBLE_QUOTES"] = 1] = "DOUBLE_QUOTES";
})(QuoteMark || (QuoteMark = {}));
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.isEnabled = function () {
        if (_super.prototype.isEnabled.call(this)) {
            var quoteMarkString = this.getOptions().ruleArguments[0];
            return (quoteMarkString === "single" || quoteMarkString === "double");
        }
        return false;
    };
    Rule.prototype.apply = function (sourceFile) {
        return this.applyWithWalker(new QuotemarkWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "quotemark",
        description: "Requires single or double quotes for string literals.",
        optionsDescription: (_a = ["\n            Five arguments may be optionally provided:\n\n            * `\"single\"` enforces single quotes.\n            * `\"double\"` enforces double quotes.\n            * `\"jsx-single\"` enforces single quotes for JSX attributes.\n            * `\"jsx-double\"` enforces double quotes for JSX attributes.\n            * `\"avoid-escape\"` allows you to use the \"other\" quotemark in cases where escaping would normally be required.\n            For example, `[true, \"double\", \"avoid-escape\"]` would not report a failure on the string literal `'Hello \"World\"'`."], _a.raw = ["\n            Five arguments may be optionally provided:\n\n            * \\`\"single\"\\` enforces single quotes.\n            * \\`\"double\"\\` enforces double quotes.\n            * \\`\"jsx-single\"\\` enforces single quotes for JSX attributes.\n            * \\`\"jsx-double\"\\` enforces double quotes for JSX attributes.\n            * \\`\"avoid-escape\"\\` allows you to use the \"other\" quotemark in cases where escaping would normally be required.\n            For example, \\`[true, \"double\", \"avoid-escape\"]\\` would not report a failure on the string literal \\`'Hello \"World\"'\\`."], Lint.Utils.dedent(_a)),
        options: {
            type: "array",
            items: {
                type: "string",
                enum: ["single", "double", "jsx-single", "jsx-double", "avoid-escape"],
            },
            minLength: 0,
            maxLength: 5,
        },
        optionExamples: ['[true, "single", "avoid-escape"]', '[true, "single", "jsx-double"]'],
        type: "style",
    };
    Rule.SINGLE_QUOTE_FAILURE = "\" should be '";
    Rule.DOUBLE_QUOTE_FAILURE = "' should be \"";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var QuotemarkWalker = (function (_super) {
    __extends(QuotemarkWalker, _super);
    function QuotemarkWalker(sourceFile, options) {
        _super.call(this, sourceFile, options);
        this.quoteMark = QuoteMark.DOUBLE_QUOTES;
        var ruleArguments = this.getOptions();
        if (ruleArguments.indexOf("single") > -1) {
            this.quoteMark = QuoteMark.SINGLE_QUOTES;
        }
        if (ruleArguments.indexOf("jsx-single") > -1) {
            this.jsxQuoteMark = QuoteMark.SINGLE_QUOTES;
        }
        else if (ruleArguments.indexOf("jsx-double") > -1) {
            this.jsxQuoteMark = QuoteMark.DOUBLE_QUOTES;
        }
        else {
            this.jsxQuoteMark = this.quoteMark;
        }
        this.avoidEscape = ruleArguments.indexOf("avoid-escape") > 0;
    }
    QuotemarkWalker.prototype.visitStringLiteral = function (node) {
        var inJsx = (node.parent.kind === ts.SyntaxKind.JsxAttribute);
        var text = node.getText();
        var width = node.getWidth();
        var position = node.getStart();
        var firstCharacter = text.charAt(0);
        var lastCharacter = text.charAt(text.length - 1);
        var quoteMark = inJsx ? this.jsxQuoteMark : this.quoteMark;
        var expectedQuoteMark = (quoteMark === QuoteMark.SINGLE_QUOTES) ? "'" : "\"";
        if (firstCharacter !== expectedQuoteMark || lastCharacter !== expectedQuoteMark) {
            var includesOtherQuoteMark = text.slice(1, -1).indexOf(expectedQuoteMark) !== -1;
            if (!(this.avoidEscape && includesOtherQuoteMark)) {
                var failureMessage = (quoteMark === QuoteMark.SINGLE_QUOTES)
                    ? Rule.SINGLE_QUOTE_FAILURE
                    : Rule.DOUBLE_QUOTE_FAILURE;
                this.addFailure(this.createFailure(position, width, failureMessage));
            }
        }
        _super.prototype.visitStringLiteral.call(this, node);
    };
    return QuotemarkWalker;
}(Lint.RuleWalker));
