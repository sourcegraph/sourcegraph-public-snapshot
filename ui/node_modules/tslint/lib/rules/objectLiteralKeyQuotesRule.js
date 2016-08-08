"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var Lint = require("../lint");
var ts = require("typescript");
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        var objectLiteralKeyQuotesWalker = new ObjectLiteralKeyQuotesWalker(sourceFile, this.getOptions());
        return this.applyWithWalker(objectLiteralKeyQuotesWalker);
    };
    Rule.metadata = {
        ruleName: "object-literal-key-quotes",
        description: "Enforces consistent object literal property quote style.",
        descriptionDetails: (_a = ["\n            Object literal property names can be defined in two ways: using literals or using strings.\n            For example, these two objects are equivalent:\n\n            var object1 = {\n                property: true\n            };\n\n            var object2 = {\n                \"property\": true\n            };\n\n            In many cases, it doesn\u2019t matter if you choose to use an identifier instead of a string\n            or vice-versa. Even so, you might decide to enforce a consistent style in your code.\n\n            This rules lets you enforce consistent quoting of property names. Either they should always\n            be quoted (default behavior) or quoted only as needed (\"as-needed\")."], _a.raw = ["\n            Object literal property names can be defined in two ways: using literals or using strings.\n            For example, these two objects are equivalent:\n\n            var object1 = {\n                property: true\n            };\n\n            var object2 = {\n                \"property\": true\n            };\n\n            In many cases, it doesn’t matter if you choose to use an identifier instead of a string\n            or vice-versa. Even so, you might decide to enforce a consistent style in your code.\n\n            This rules lets you enforce consistent quoting of property names. Either they should always\n            be quoted (default behavior) or quoted only as needed (\"as-needed\")."], Lint.Utils.dedent(_a)),
        optionsDescription: (_b = ["\n            Possible settings are:\n\n            * `\"always\"`: Property names should always be quoted. (This is the default.)\n            * `\"as-needed\"`: Only property names which require quotes may be quoted (e.g. those with spaces in them).\n\n            For ES6, computed property names (`{[name]: value}`) and methods (`{foo() {}}`) never need\n            to be quoted."], _b.raw = ["\n            Possible settings are:\n\n            * \\`\"always\"\\`: Property names should always be quoted. (This is the default.)\n            * \\`\"as-needed\"\\`: Only property names which require quotes may be quoted (e.g. those with spaces in them).\n\n            For ES6, computed property names (\\`{[name]: value}\\`) and methods (\\`{foo() {}}\\`) never need\n            to be quoted."], Lint.Utils.dedent(_b)),
        options: {
            type: "string",
            enum: ["always", "as-needed"],
        },
        optionExamples: ["[true, \"as-needed\"]", "[true, \"always\"]"],
        type: "style",
    };
    Rule.UNNEEDED_QUOTES = function (name) { return ("Unnecessarily quoted property '" + name + "' found."); };
    Rule.UNQUOTED_PROPERTY = function (name) { return ("Unquoted property '" + name + "' found."); };
    return Rule;
    var _a, _b;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var IDENTIFIER_NAME_REGEX = /^(?:[\$A-Z_a-z])*$/;
var NUMBER_REGEX = /^[0-9]+$/;
var ObjectLiteralKeyQuotesWalker = (function (_super) {
    __extends(ObjectLiteralKeyQuotesWalker, _super);
    function ObjectLiteralKeyQuotesWalker(sourceFile, options) {
        _super.call(this, sourceFile, options);
        this.mode = this.getOptions()[0] || "always";
    }
    ObjectLiteralKeyQuotesWalker.prototype.visitPropertyAssignment = function (node) {
        var name = node.name;
        if (this.mode === "always") {
            if (name.kind !== ts.SyntaxKind.StringLiteral &&
                name.kind !== ts.SyntaxKind.ComputedPropertyName) {
                this.addFailure(this.createFailure(name.getStart(), name.getWidth(), Rule.UNQUOTED_PROPERTY(name.getText())));
            }
        }
        else if (this.mode === "as-needed") {
            if (name.kind === ts.SyntaxKind.StringLiteral) {
                var stringNode = name;
                var property = stringNode.text;
                var isIdentifier = IDENTIFIER_NAME_REGEX.test(property);
                var isNumber = NUMBER_REGEX.test(property);
                if (isIdentifier || (isNumber && Number(property).toString() === property)) {
                    this.addFailure(this.createFailure(stringNode.getStart(), stringNode.getWidth(), Rule.UNNEEDED_QUOTES(property)));
                }
            }
        }
        _super.prototype.visitPropertyAssignment.call(this, node);
    };
    return ObjectLiteralKeyQuotesWalker;
}(Lint.RuleWalker));
