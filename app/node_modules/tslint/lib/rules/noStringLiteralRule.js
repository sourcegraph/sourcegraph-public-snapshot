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
        return this.applyWithWalker(new NoStringLiteralWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-string-literal",
        description: "Disallows object access via string literals.",
        rationale: "Encourages using strongly-typed property access.",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING = "object access via string literals is disallowed";
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoStringLiteralWalker = (function (_super) {
    __extends(NoStringLiteralWalker, _super);
    function NoStringLiteralWalker() {
        _super.apply(this, arguments);
    }
    NoStringLiteralWalker.prototype.visitElementAccessExpression = function (node) {
        var argument = node.argumentExpression;
        if (argument != null) {
            var accessorText = argument.getText();
            if (argument.kind === ts.SyntaxKind.StringLiteral && accessorText.length > 2) {
                var unquotedAccessorText = accessorText.substring(1, accessorText.length - 1);
                if (isValidIdentifier(unquotedAccessorText)) {
                    this.addFailure(this.createFailure(argument.getStart(), argument.getWidth(), Rule.FAILURE_STRING));
                }
            }
        }
        _super.prototype.visitElementAccessExpression.call(this, node);
    };
    return NoStringLiteralWalker;
}(Lint.RuleWalker));
function isValidIdentifier(token) {
    var scanner = ts.createScanner(ts.ScriptTarget.ES5, false, ts.LanguageVariant.Standard, token);
    scanner.scan();
    return scanner.getTokenText() === token && scanner.isIdentifier();
}
