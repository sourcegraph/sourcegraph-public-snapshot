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
        return this.applyWithWalker(new NoNamespaceWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-namespace",
        description: "Disallows use of internal \`module\`s and \`namespace\`s.",
        descriptionDetails: "This rule still allows the use of `declare module ... {}`",
        rationale: (_a = ["\n            ES6-style external modules are the standard way to modularize code.\n            Using `module {}` and `namespace {}` are outdated ways to organize TypeScript code."], _a.raw = ["\n            ES6-style external modules are the standard way to modularize code.\n            Using \\`module {}\\` and \\`namespace {}\\` are outdated ways to organize TypeScript code."], Lint.Utils.dedent(_a)),
        optionsDescription: (_b = ["\n            One argument may be optionally provided:\n\n            * `allow-declarations` allows `declare namespace ... {}` to describe external APIs."], _b.raw = ["\n            One argument may be optionally provided:\n\n            * \\`allow-declarations\\` allows \\`declare namespace ... {}\\` to describe external APIs."], Lint.Utils.dedent(_b)),
        options: {
            type: "array",
            items: {
                type: "string",
                enum: ["allow-declarations"],
            },
            minLength: 0,
            maxLength: 1,
        },
        optionExamples: ["true", '[true, "allow-declarations"]'],
        type: "typescript",
    };
    Rule.FAILURE_STRING = "'namespace' and 'module' are disallowed";
    return Rule;
    var _a, _b;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoNamespaceWalker = (function (_super) {
    __extends(NoNamespaceWalker, _super);
    function NoNamespaceWalker() {
        _super.apply(this, arguments);
    }
    NoNamespaceWalker.prototype.visitModuleDeclaration = function (decl) {
        _super.prototype.visitModuleDeclaration.call(this, decl);
        if (decl.name.kind === ts.SyntaxKind.StringLiteral) {
            return;
        }
        if (Lint.isNodeFlagSet(decl, ts.NodeFlags.Ambient) && this.hasOption("allow-declarations")) {
            return;
        }
        if (Lint.isNestedModuleDeclaration(decl)) {
            return;
        }
        this.addFailure(this.createFailure(decl.getStart(), decl.getWidth(), Rule.FAILURE_STRING));
    };
    return NoNamespaceWalker;
}(Lint.RuleWalker));
