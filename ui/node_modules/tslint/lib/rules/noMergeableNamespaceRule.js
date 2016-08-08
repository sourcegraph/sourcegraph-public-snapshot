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
    Rule.failureStringFactory = function (identifier, locationToMerge) {
        return "Mergeable namespace " + identifier + " found. Merge its contents with the namespace on line " + locationToMerge.line + ".";
    };
    Rule.prototype.apply = function (sourceFile) {
        var languageService = Lint.createLanguageService(sourceFile.fileName, sourceFile.getFullText());
        var noMergeableNamespaceWalker = new NoMergeableNamespaceWalker(sourceFile, this.getOptions(), languageService);
        return this.applyWithWalker(noMergeableNamespaceWalker);
    };
    Rule.metadata = {
        ruleName: "no-mergeable-namespace",
        description: "Disallows mergeable namespaces in the same file.",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "maintainability",
    };
    return Rule;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoMergeableNamespaceWalker = (function (_super) {
    __extends(NoMergeableNamespaceWalker, _super);
    function NoMergeableNamespaceWalker(sourceFile, options, languageService) {
        _super.call(this, sourceFile, options);
        this.languageService = languageService;
    }
    NoMergeableNamespaceWalker.prototype.visitModuleDeclaration = function (node) {
        if (Lint.isNodeFlagSet(node, ts.NodeFlags.Namespace)
            && node.name.kind === ts.SyntaxKind.Identifier) {
            this.validateReferencesForNamespace(node.name.text, node.name.getStart());
        }
        _super.prototype.visitModuleDeclaration.call(this, node);
    };
    NoMergeableNamespaceWalker.prototype.validateReferencesForNamespace = function (name, position) {
        var fileName = this.getSourceFile().fileName;
        var highlights = this.languageService.getDocumentHighlights(fileName, position, [fileName]);
        if (highlights == null || highlights[0].highlightSpans.length > 1) {
            var failureString = Rule.failureStringFactory(name, this.findLocationToMerge(position, highlights[0].highlightSpans));
            this.addFailure(this.createFailure(position, name.length, failureString));
        }
    };
    NoMergeableNamespaceWalker.prototype.findLocationToMerge = function (currentPosition, highlightSpans) {
        var line = ts.getLineAndCharacterOfPosition(this.getSourceFile(), currentPosition).line;
        for (var _i = 0, highlightSpans_1 = highlightSpans; _i < highlightSpans_1.length; _i++) {
            var span = highlightSpans_1[_i];
            var lineAndCharacter = ts.getLineAndCharacterOfPosition(this.getSourceFile(), span.textSpan.start);
            if (lineAndCharacter.line !== line) {
                return lineAndCharacter;
            }
        }
    };
    return NoMergeableNamespaceWalker;
}(Lint.RuleWalker));
