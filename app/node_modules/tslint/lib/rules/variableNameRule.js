"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ts = require("typescript");
var Lint = require("../lint");
var BANNED_KEYWORDS = ["any", "Number", "number", "String", "string", "Boolean", "boolean", "Undefined", "undefined"];
var OPTION_LEADING_UNDERSCORE = "allow-leading-underscore";
var OPTION_TRAILING_UNDERSCORE = "allow-trailing-underscore";
var OPTION_BAN_KEYWORDS = "ban-keywords";
var OPTION_CHECK_FORMAT = "check-format";
var OPTION_ALLOW_PASCAL_CASE = "allow-pascal-case";
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        var variableNameWalker = new VariableNameWalker(sourceFile, this.getOptions());
        return this.applyWithWalker(variableNameWalker);
    };
    Rule.metadata = {
        ruleName: "variable-name",
        description: "Checks variable names for various errors.",
        optionsDescription: (_a = ["\n            Five arguments may be optionally provided:\n\n            * `\"", "\"`: allows only camelCased or UPPER_CASED variable names\n              * `\"", "\"` allows underscores at the beginning (only has an effect if \"check-format\" specified)\n              * `\"", "\"` allows underscores at the end. (only has an effect if \"check-format\" specified)\n              * `\"", "\"` allows PascalCase in addtion to camelCase.\n            * `\"", "\"`: disallows the use of certain TypeScript keywords (`any`, `Number`, `number`, `String`,\n            `string`, `Boolean`, `boolean`, `undefined`) as variable or parameter names."], _a.raw = ["\n            Five arguments may be optionally provided:\n\n            * \\`\"", "\"\\`: allows only camelCased or UPPER_CASED variable names\n              * \\`\"", "\"\\` allows underscores at the beginning (only has an effect if \"check-format\" specified)\n              * \\`\"", "\"\\` allows underscores at the end. (only has an effect if \"check-format\" specified)\n              * \\`\"", "\"\\` allows PascalCase in addtion to camelCase.\n            * \\`\"", "\"\\`: disallows the use of certain TypeScript keywords (\\`any\\`, \\`Number\\`, \\`number\\`, \\`String\\`,\n            \\`string\\`, \\`Boolean\\`, \\`boolean\\`, \\`undefined\\`) as variable or parameter names."], Lint.Utils.dedent(_a, OPTION_CHECK_FORMAT, OPTION_LEADING_UNDERSCORE, OPTION_TRAILING_UNDERSCORE, OPTION_ALLOW_PASCAL_CASE, OPTION_BAN_KEYWORDS)),
        options: {
            type: "array",
            items: {
                type: "string",
                enum: [
                    OPTION_CHECK_FORMAT,
                    OPTION_LEADING_UNDERSCORE,
                    OPTION_TRAILING_UNDERSCORE,
                    OPTION_ALLOW_PASCAL_CASE,
                    OPTION_BAN_KEYWORDS,
                ],
            },
            minLength: 0,
            maxLength: 5,
        },
        optionExamples: ['[true, "ban-keywords", "check-format", "allow-leading-underscore"]'],
        type: "style",
    };
    Rule.FORMAT_FAILURE = "variable name must be in camelcase or uppercase";
    Rule.KEYWORD_FAILURE = "variable name clashes with keyword/type";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var VariableNameWalker = (function (_super) {
    __extends(VariableNameWalker, _super);
    function VariableNameWalker(sourceFile, options) {
        _super.call(this, sourceFile, options);
        this.shouldBanKeywords = this.hasOption(OPTION_BAN_KEYWORDS);
        this.shouldCheckFormat = !this.shouldBanKeywords || this.hasOption(OPTION_CHECK_FORMAT);
    }
    VariableNameWalker.prototype.visitBindingElement = function (node) {
        if (node.name.kind === ts.SyntaxKind.Identifier) {
            var identifier = node.name;
            this.handleVariableNameFormat(identifier);
            this.handleVariableNameKeyword(identifier);
        }
        _super.prototype.visitBindingElement.call(this, node);
    };
    VariableNameWalker.prototype.visitParameterDeclaration = function (node) {
        if (node.name.kind === ts.SyntaxKind.Identifier) {
            var identifier = node.name;
            this.handleVariableNameFormat(identifier);
            this.handleVariableNameKeyword(identifier);
        }
        _super.prototype.visitParameterDeclaration.call(this, node);
    };
    VariableNameWalker.prototype.visitPropertyDeclaration = function (node) {
        if (node.name != null && node.name.kind === ts.SyntaxKind.Identifier) {
            var identifier = node.name;
            this.handleVariableNameFormat(identifier);
        }
        _super.prototype.visitPropertyDeclaration.call(this, node);
    };
    VariableNameWalker.prototype.visitVariableDeclaration = function (node) {
        if (node.name.kind === ts.SyntaxKind.Identifier) {
            var identifier = node.name;
            this.handleVariableNameFormat(identifier);
            this.handleVariableNameKeyword(identifier);
        }
        _super.prototype.visitVariableDeclaration.call(this, node);
    };
    VariableNameWalker.prototype.visitVariableStatement = function (node) {
        if (!Lint.hasModifier(node.modifiers, ts.SyntaxKind.DeclareKeyword)) {
            _super.prototype.visitVariableStatement.call(this, node);
        }
    };
    VariableNameWalker.prototype.handleVariableNameFormat = function (name) {
        var variableName = name.text;
        if (this.shouldCheckFormat && !this.isCamelCase(variableName) && !isUpperCase(variableName)) {
            this.addFailure(this.createFailure(name.getStart(), name.getWidth(), Rule.FORMAT_FAILURE));
        }
    };
    VariableNameWalker.prototype.handleVariableNameKeyword = function (name) {
        var variableName = name.text;
        if (this.shouldBanKeywords && BANNED_KEYWORDS.indexOf(variableName) !== -1) {
            this.addFailure(this.createFailure(name.getStart(), name.getWidth(), Rule.KEYWORD_FAILURE));
        }
    };
    VariableNameWalker.prototype.isCamelCase = function (name) {
        var firstCharacter = name.charAt(0);
        var lastCharacter = name.charAt(name.length - 1);
        var middle = name.substr(1, name.length - 2);
        if (name.length <= 0) {
            return true;
        }
        if (!this.hasOption(OPTION_LEADING_UNDERSCORE) && firstCharacter === "_") {
            return false;
        }
        if (!this.hasOption(OPTION_TRAILING_UNDERSCORE) && lastCharacter === "_") {
            return false;
        }
        if (!this.hasOption(OPTION_ALLOW_PASCAL_CASE) && !isLowerCase(firstCharacter)) {
            return false;
        }
        return middle.indexOf("_") === -1;
    };
    return VariableNameWalker;
}(Lint.RuleWalker));
function isLowerCase(name) {
    return name === name.toLowerCase();
}
function isUpperCase(name) {
    return name === name.toUpperCase();
}
