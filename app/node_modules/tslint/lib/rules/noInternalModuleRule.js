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
        return this.applyWithWalker(new NoInternalModuleWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-internal-module",
        description: "Disallows internal `module`",
        rationale: "Using `module` leads to a confusion of concepts with external modules. Use the newer `namespace` keyword instead.",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "typescript",
    };
    Rule.FAILURE_STRING = "The internal 'module' syntax is deprecated, use the 'namespace' keyword instead.";
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoInternalModuleWalker = (function (_super) {
    __extends(NoInternalModuleWalker, _super);
    function NoInternalModuleWalker() {
        _super.apply(this, arguments);
    }
    NoInternalModuleWalker.prototype.visitModuleDeclaration = function (node) {
        if (this.isInternalModuleDeclaration(node)) {
            this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING));
        }
        _super.prototype.visitModuleDeclaration.call(this, node);
    };
    NoInternalModuleWalker.prototype.isInternalModuleDeclaration = function (node) {
        return !Lint.isNodeFlagSet(node, ts.NodeFlags.Namespace)
            && !Lint.isNestedModuleDeclaration(node)
            && node.name.kind === ts.SyntaxKind.Identifier
            && !isGlobalAugmentation(node);
    };
    return NoInternalModuleWalker;
}(Lint.RuleWalker));
function isGlobalAugmentation(node) {
    return node.name.kind === ts.SyntaxKind.Identifier && node.name.text === "global";
}
