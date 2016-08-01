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
        var languageService = Lint.createLanguageService(sourceFile.fileName, sourceFile.getFullText());
        return this.applyWithWalker(new NoUseBeforeDeclareWalker(sourceFile, this.getOptions(), languageService));
    };
    Rule.metadata = {
        ruleName: "no-use-before-declare",
        description: "Disallows usage of variables before their declaration.",
        descriptionDetails: (_a = ["\n            This rule is primarily useful when using the `var` keyword -\n            the compiler will detect if a `let` and `const` variable is used before it is declared."], _a.raw = ["\n            This rule is primarily useful when using the \\`var\\` keyword -\n            the compiler will detect if a \\`let\\` and \\`const\\` variable is used before it is declared."], Lint.Utils.dedent(_a)),
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING_PREFIX = "variable '";
    Rule.FAILURE_STRING_POSTFIX = "' used before declaration";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoUseBeforeDeclareWalker = (function (_super) {
    __extends(NoUseBeforeDeclareWalker, _super);
    function NoUseBeforeDeclareWalker(sourceFile, options, languageService) {
        _super.call(this, sourceFile, options);
        this.languageService = languageService;
    }
    NoUseBeforeDeclareWalker.prototype.createScope = function () {
        return {};
    };
    NoUseBeforeDeclareWalker.prototype.visitBindingElement = function (node) {
        var isSingleVariable = node.name.kind === ts.SyntaxKind.Identifier;
        var isBlockScoped = Lint.isBlockScopedBindingElement(node);
        if (isSingleVariable && !isBlockScoped) {
            var variableName = node.name.text;
            this.validateUsageForVariable(variableName, node.name.getStart());
        }
        _super.prototype.visitBindingElement.call(this, node);
    };
    NoUseBeforeDeclareWalker.prototype.visitImportDeclaration = function (node) {
        var importClause = node.importClause;
        if (importClause != null && importClause.name != null) {
            var variableIdentifier = importClause.name;
            this.validateUsageForVariable(variableIdentifier.text, variableIdentifier.getStart());
        }
        _super.prototype.visitImportDeclaration.call(this, node);
    };
    NoUseBeforeDeclareWalker.prototype.visitImportEqualsDeclaration = function (node) {
        var name = node.name;
        this.validateUsageForVariable(name.text, name.getStart());
        _super.prototype.visitImportEqualsDeclaration.call(this, node);
    };
    NoUseBeforeDeclareWalker.prototype.visitNamedImports = function (node) {
        for (var _i = 0, _a = node.elements; _i < _a.length; _i++) {
            var namedImport = _a[_i];
            this.validateUsageForVariable(namedImport.name.text, namedImport.name.getStart());
        }
        _super.prototype.visitNamedImports.call(this, node);
    };
    NoUseBeforeDeclareWalker.prototype.visitNamespaceImport = function (node) {
        this.validateUsageForVariable(node.name.text, node.name.getStart());
        _super.prototype.visitNamespaceImport.call(this, node);
    };
    NoUseBeforeDeclareWalker.prototype.visitVariableDeclaration = function (node) {
        var isSingleVariable = node.name.kind === ts.SyntaxKind.Identifier;
        var variableName = node.name.text;
        var currentScope = this.getCurrentScope();
        if (isSingleVariable && currentScope[variableName] == null) {
            this.validateUsageForVariable(variableName, node.getStart());
        }
        currentScope[variableName] = true;
        _super.prototype.visitVariableDeclaration.call(this, node);
    };
    NoUseBeforeDeclareWalker.prototype.validateUsageForVariable = function (name, position) {
        var fileName = this.getSourceFile().fileName;
        var highlights = this.languageService.getDocumentHighlights(fileName, position, [fileName]);
        if (highlights != null) {
            for (var _i = 0, highlights_1 = highlights; _i < highlights_1.length; _i++) {
                var highlight = highlights_1[_i];
                for (var _a = 0, _b = highlight.highlightSpans; _a < _b.length; _a++) {
                    var highlightSpan = _b[_a];
                    var referencePosition = highlightSpan.textSpan.start;
                    if (referencePosition < position) {
                        var failureString = Rule.FAILURE_STRING_PREFIX + name + Rule.FAILURE_STRING_POSTFIX;
                        this.addFailure(this.createFailure(referencePosition, name.length, failureString));
                    }
                }
            }
        }
    };
    return NoUseBeforeDeclareWalker;
}(Lint.ScopeAwareRuleWalker));
