"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ts = require("typescript");
var Lint = require("../lint");
var OPTION_REACT = "react";
var OPTION_CHECK_PARAMETERS = "check-parameters";
var REACT_MODULES = ["react", "react/addons"];
var REACT_NAMESPACE_IMPORT_NAME = "React";
var MODULE_SPECIFIER_MATCH = /^["'](.+)['"]$/;
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        var languageService = Lint.createLanguageService(sourceFile.fileName, sourceFile.getFullText());
        return this.applyWithWalker(new NoUnusedVariablesWalker(sourceFile, this.getOptions(), languageService));
    };
    Rule.metadata = {
        ruleName: "no-unused-variable",
        description: "Disallows unused imports, variables, functions and private class members.",
        optionsDescription: (_a = ["\n            Three optional arguments may be optionally provided:\n\n            * `\"check-parameters\"` disallows unused function and constructor parameters.\n                * NOTE: this option is experimental and does not work with classes\n                that use abstract method declarations, among other things.\n            * `\"react\"` relaxes the rule for a namespace import named `React`\n            (from either the module `\"react\"` or `\"react/addons\"`).\n            Any JSX expression in the file will be treated as a usage of `React`\n            (because it expands to `React.createElement `).\n            * `{\"ignore-pattern\": \"pattern\"}` where pattern is a case-sensitive regexp.\n            Variable names that match the pattern will be ignored."], _a.raw = ["\n            Three optional arguments may be optionally provided:\n\n            * \\`\"check-parameters\"\\` disallows unused function and constructor parameters.\n                * NOTE: this option is experimental and does not work with classes\n                that use abstract method declarations, among other things.\n            * \\`\"react\"\\` relaxes the rule for a namespace import named \\`React\\`\n            (from either the module \\`\"react\"\\` or \\`\"react/addons\"\\`).\n            Any JSX expression in the file will be treated as a usage of \\`React\\`\n            (because it expands to \\`React.createElement \\`).\n            * \\`{\"ignore-pattern\": \"pattern\"}\\` where pattern is a case-sensitive regexp.\n            Variable names that match the pattern will be ignored."], Lint.Utils.dedent(_a)),
        options: {
            type: "array",
            items: {
                oneOf: [{
                        type: "string",
                        enum: ["check-parameters", "react"],
                    }, {
                        type: "object",
                        properties: {
                            "ignore-pattern": { type: "string" },
                        },
                        additionalProperties: false,
                    }],
            },
            minLength: 0,
            maxLength: 3,
        },
        optionExamples: ['[true, "react"]', '[true, {"ignore-pattern": "^_"}]'],
        type: "functionality",
    };
    Rule.FAILURE_TYPE_FUNC = "function";
    Rule.FAILURE_TYPE_IMPORT = "import";
    Rule.FAILURE_TYPE_METHOD = "method";
    Rule.FAILURE_TYPE_PARAM = "parameter";
    Rule.FAILURE_TYPE_PROP = "property";
    Rule.FAILURE_TYPE_VAR = "variable";
    Rule.FAILURE_STRING_FACTORY = function (type, name) { return ("Unused " + type + ": '" + name + "'"); };
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoUnusedVariablesWalker = (function (_super) {
    __extends(NoUnusedVariablesWalker, _super);
    function NoUnusedVariablesWalker(sourceFile, options, languageService) {
        _super.call(this, sourceFile, options);
        this.languageService = languageService;
        this.skipVariableDeclaration = false;
        this.skipParameterDeclaration = false;
        this.hasSeenJsxElement = false;
        this.isReactUsed = false;
        var ignorePatternOption = this.getOptions().filter(function (option) {
            return typeof option === "object" && option["ignore-pattern"] != null;
        })[0];
        if (ignorePatternOption != null) {
            this.ignorePattern = new RegExp(ignorePatternOption["ignore-pattern"]);
        }
    }
    NoUnusedVariablesWalker.prototype.visitSourceFile = function (node) {
        _super.prototype.visitSourceFile.call(this, node);
        if (this.hasOption(OPTION_REACT)
            && this.reactImport != null
            && !this.isReactUsed
            && !this.hasSeenJsxElement) {
            var nameText = this.reactImport.name.getText();
            if (!this.isIgnored(nameText)) {
                var start = this.reactImport.name.getStart();
                var msg = Rule.FAILURE_STRING_FACTORY(Rule.FAILURE_TYPE_IMPORT, nameText);
                this.addFailure(this.createFailure(start, nameText.length, msg));
            }
        }
    };
    NoUnusedVariablesWalker.prototype.visitBindingElement = function (node) {
        var isSingleVariable = node.name.kind === ts.SyntaxKind.Identifier;
        if (isSingleVariable && !this.skipBindingElement) {
            var variableIdentifier = node.name;
            this.validateReferencesForVariable(Rule.FAILURE_TYPE_VAR, variableIdentifier.text, variableIdentifier.getStart());
        }
        _super.prototype.visitBindingElement.call(this, node);
    };
    NoUnusedVariablesWalker.prototype.visitCatchClause = function (node) {
        this.visitBlock(node.block);
    };
    NoUnusedVariablesWalker.prototype.visitFunctionDeclaration = function (node) {
        if (!Lint.hasModifier(node.modifiers, ts.SyntaxKind.ExportKeyword, ts.SyntaxKind.DeclareKeyword)) {
            var variableName = node.name.text;
            this.validateReferencesForVariable(Rule.FAILURE_TYPE_FUNC, variableName, node.name.getStart());
        }
        _super.prototype.visitFunctionDeclaration.call(this, node);
    };
    NoUnusedVariablesWalker.prototype.visitFunctionType = function (node) {
        this.skipParameterDeclaration = true;
        _super.prototype.visitFunctionType.call(this, node);
        this.skipParameterDeclaration = false;
    };
    NoUnusedVariablesWalker.prototype.visitImportDeclaration = function (node) {
        if (!Lint.hasModifier(node.modifiers, ts.SyntaxKind.ExportKeyword)) {
            var importClause = node.importClause;
            if (importClause != null && importClause.name != null) {
                var variableIdentifier = importClause.name;
                this.validateReferencesForVariable(Rule.FAILURE_TYPE_IMPORT, variableIdentifier.text, variableIdentifier.getStart());
            }
        }
        _super.prototype.visitImportDeclaration.call(this, node);
    };
    NoUnusedVariablesWalker.prototype.visitImportEqualsDeclaration = function (node) {
        if (!Lint.hasModifier(node.modifiers, ts.SyntaxKind.ExportKeyword)) {
            var name_1 = node.name;
            this.validateReferencesForVariable(Rule.FAILURE_TYPE_IMPORT, name_1.text, name_1.getStart());
        }
        _super.prototype.visitImportEqualsDeclaration.call(this, node);
    };
    NoUnusedVariablesWalker.prototype.visitIndexSignatureDeclaration = function (node) {
        this.skipParameterDeclaration = true;
        _super.prototype.visitIndexSignatureDeclaration.call(this, node);
        this.skipParameterDeclaration = false;
    };
    NoUnusedVariablesWalker.prototype.visitInterfaceDeclaration = function (node) {
        this.skipParameterDeclaration = true;
        _super.prototype.visitInterfaceDeclaration.call(this, node);
        this.skipParameterDeclaration = false;
    };
    NoUnusedVariablesWalker.prototype.visitJsxElement = function (node) {
        this.hasSeenJsxElement = true;
        _super.prototype.visitJsxElement.call(this, node);
    };
    NoUnusedVariablesWalker.prototype.visitJsxSelfClosingElement = function (node) {
        this.hasSeenJsxElement = true;
        _super.prototype.visitJsxSelfClosingElement.call(this, node);
    };
    NoUnusedVariablesWalker.prototype.visitMethodDeclaration = function (node) {
        if (node.name != null && node.name.kind === ts.SyntaxKind.Identifier) {
            var modifiers = node.modifiers;
            var variableName = node.name.text;
            if (Lint.hasModifier(modifiers, ts.SyntaxKind.PrivateKeyword)) {
                this.validateReferencesForVariable(Rule.FAILURE_TYPE_METHOD, variableName, node.name.getStart());
            }
        }
        if (Lint.hasModifier(node.modifiers, ts.SyntaxKind.AbstractKeyword)) {
            this.skipParameterDeclaration = true;
        }
        _super.prototype.visitMethodDeclaration.call(this, node);
        this.skipParameterDeclaration = false;
    };
    NoUnusedVariablesWalker.prototype.visitNamedImports = function (node) {
        for (var _i = 0, _a = node.elements; _i < _a.length; _i++) {
            var namedImport = _a[_i];
            this.validateReferencesForVariable(Rule.FAILURE_TYPE_IMPORT, namedImport.name.text, namedImport.name.getStart());
        }
        _super.prototype.visitNamedImports.call(this, node);
    };
    NoUnusedVariablesWalker.prototype.visitNamespaceImport = function (node) {
        var importDeclaration = node.parent.parent;
        var moduleSpecifier = importDeclaration.moduleSpecifier.getText();
        var moduleNameMatch = moduleSpecifier.match(MODULE_SPECIFIER_MATCH);
        var isReactImport = (moduleNameMatch != null) && (REACT_MODULES.indexOf(moduleNameMatch[1]) !== -1);
        if (this.hasOption(OPTION_REACT) && isReactImport && node.name.text === REACT_NAMESPACE_IMPORT_NAME) {
            this.reactImport = node;
            var fileName = this.getSourceFile().fileName;
            var position = node.name.getStart();
            var highlights = this.languageService.getDocumentHighlights(fileName, position, [fileName]);
            if (highlights != null && highlights[0].highlightSpans.length > 1) {
                this.isReactUsed = true;
            }
        }
        else {
            this.validateReferencesForVariable(Rule.FAILURE_TYPE_IMPORT, node.name.text, node.name.getStart());
        }
        _super.prototype.visitNamespaceImport.call(this, node);
    };
    NoUnusedVariablesWalker.prototype.visitParameterDeclaration = function (node) {
        var isSingleVariable = node.name.kind === ts.SyntaxKind.Identifier;
        var isPropertyParameter = Lint.hasModifier(node.modifiers, ts.SyntaxKind.PublicKeyword, ts.SyntaxKind.PrivateKeyword, ts.SyntaxKind.ProtectedKeyword);
        if (!isSingleVariable && isPropertyParameter) {
            this.skipBindingElement = true;
        }
        if (this.hasOption(OPTION_CHECK_PARAMETERS)
            && isSingleVariable
            && !this.skipParameterDeclaration
            && !Lint.hasModifier(node.modifiers, ts.SyntaxKind.PublicKeyword)) {
            var nameNode = node.name;
            this.validateReferencesForVariable(Rule.FAILURE_TYPE_PARAM, nameNode.text, node.name.getStart());
        }
        _super.prototype.visitParameterDeclaration.call(this, node);
        this.skipBindingElement = false;
    };
    NoUnusedVariablesWalker.prototype.visitPropertyDeclaration = function (node) {
        if (node.name != null && node.name.kind === ts.SyntaxKind.Identifier) {
            var modifiers = node.modifiers;
            var variableName = node.name.text;
            if (Lint.hasModifier(modifiers, ts.SyntaxKind.PrivateKeyword)) {
                this.validateReferencesForVariable(Rule.FAILURE_TYPE_PROP, variableName, node.name.getStart());
            }
        }
        _super.prototype.visitPropertyDeclaration.call(this, node);
    };
    NoUnusedVariablesWalker.prototype.visitVariableDeclaration = function (node) {
        var isSingleVariable = node.name.kind === ts.SyntaxKind.Identifier;
        if (isSingleVariable && !this.skipVariableDeclaration) {
            var variableIdentifier = node.name;
            this.validateReferencesForVariable(Rule.FAILURE_TYPE_VAR, variableIdentifier.text, variableIdentifier.getStart());
        }
        _super.prototype.visitVariableDeclaration.call(this, node);
    };
    NoUnusedVariablesWalker.prototype.visitVariableStatement = function (node) {
        if (Lint.hasModifier(node.modifiers, ts.SyntaxKind.ExportKeyword, ts.SyntaxKind.DeclareKeyword)) {
            this.skipBindingElement = true;
            this.skipVariableDeclaration = true;
        }
        _super.prototype.visitVariableStatement.call(this, node);
        this.skipBindingElement = false;
        this.skipVariableDeclaration = false;
    };
    NoUnusedVariablesWalker.prototype.validateReferencesForVariable = function (type, name, position) {
        var fileName = this.getSourceFile().fileName;
        var highlights = this.languageService.getDocumentHighlights(fileName, position, [fileName]);
        if ((highlights == null || highlights[0].highlightSpans.length <= 1) && !this.isIgnored(name)) {
            this.addFailure(this.createFailure(position, name.length, Rule.FAILURE_STRING_FACTORY(type, name)));
        }
    };
    NoUnusedVariablesWalker.prototype.isIgnored = function (name) {
        return this.ignorePattern != null && this.ignorePattern.test(name);
    };
    return NoUnusedVariablesWalker;
}(Lint.RuleWalker));
