"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ts = require("typescript");
var Lint = require("../lint");
var OPTION_IGNORE_PARMS = "ignore-params";
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        return this.applyWithWalker(new NoInferrableTypesWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-inferrable-types",
        description: "Disallows explicit type declarations for variables or parameters initialized to a number, string, or boolean.",
        rationale: "Explicit types where they can be easily infered by the compiler make code more verbose.",
        optionsDescription: (_a = ["\n            One argument may be optionally provided:\n\n            * `", "` allows specifying an inferrable type annotation for function params.\n            This can be useful when combining with the `typedef` rule."], _a.raw = ["\n            One argument may be optionally provided:\n\n            * \\`", "\\` allows specifying an inferrable type annotation for function params.\n            This can be useful when combining with the \\`typedef\\` rule."], Lint.Utils.dedent(_a, OPTION_IGNORE_PARMS)),
        options: {
            type: "array",
            items: {
                type: "string",
                enum: [OPTION_IGNORE_PARMS],
            },
            minLength: 0,
            maxLength: 1,
        },
        optionExamples: ["true", ("[true, \"" + OPTION_IGNORE_PARMS + "\"]")],
        type: "typescript",
    };
    Rule.FAILURE_STRING_FACTORY = function (type) { return ("LHS type (" + type + ") inferred by RHS expression, remove type annotation"); };
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoInferrableTypesWalker = (function (_super) {
    __extends(NoInferrableTypesWalker, _super);
    function NoInferrableTypesWalker() {
        _super.apply(this, arguments);
    }
    NoInferrableTypesWalker.prototype.visitVariableDeclaration = function (node) {
        this.checkDeclaration(node);
        _super.prototype.visitVariableDeclaration.call(this, node);
    };
    NoInferrableTypesWalker.prototype.visitParameterDeclaration = function (node) {
        if (!this.hasOption(OPTION_IGNORE_PARMS)) {
            this.checkDeclaration(node);
        }
        _super.prototype.visitParameterDeclaration.call(this, node);
    };
    NoInferrableTypesWalker.prototype.checkDeclaration = function (node) {
        if (node.type != null && node.initializer != null) {
            var failure = void 0;
            switch (node.type.kind) {
                case ts.SyntaxKind.BooleanKeyword:
                    if (node.initializer.kind === ts.SyntaxKind.TrueKeyword || node.initializer.kind === ts.SyntaxKind.FalseKeyword) {
                        failure = "boolean";
                    }
                    break;
                case ts.SyntaxKind.NumberKeyword:
                    if (node.initializer.kind === ts.SyntaxKind.NumericLiteral) {
                        failure = "number";
                    }
                    break;
                case ts.SyntaxKind.StringKeyword:
                    switch (node.initializer.kind) {
                        case ts.SyntaxKind.StringLiteral:
                        case ts.SyntaxKind.NoSubstitutionTemplateLiteral:
                        case ts.SyntaxKind.TemplateExpression:
                            failure = "string";
                            break;
                        default:
                            break;
                    }
                    break;
                default:
                    break;
            }
            if (failure != null) {
                this.addFailure(this.createFailure(node.type.getStart(), node.type.getWidth(), Rule.FAILURE_STRING_FACTORY(failure)));
            }
        }
    };
    return NoInferrableTypesWalker;
}(Lint.RuleWalker));
