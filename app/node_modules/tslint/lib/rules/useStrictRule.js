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
        var useStrictWalker = new UseStrictWalker(sourceFile, this.getOptions());
        return this.applyWithWalker(useStrictWalker);
    };
    Rule.metadata = {
        ruleName: "use-strict",
        description: "Requires using ECMAScript 5's strict mode.",
        optionsDescription: (_a = ["\n            Two arguments may be optionally provided:\n\n            * `check-module` checks that all top-level modules are using strict mode.\n            * `check-function` checks that all top-level functions are using strict mode."], _a.raw = ["\n            Two arguments may be optionally provided:\n\n            * \\`check-module\\` checks that all top-level modules are using strict mode.\n            * \\`check-function\\` checks that all top-level functions are using strict mode."], Lint.Utils.dedent(_a)),
        options: {
            type: "array",
            items: {
                type: "string",
                enum: ["check-module", "check-function"],
            },
            minLength: 0,
            maxLength: 2,
        },
        optionExamples: ['[true, "check-module"]'],
        type: "functionality",
    };
    Rule.FAILURE_STRING = "missing 'use strict'";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var UseStrictWalker = (function (_super) {
    __extends(UseStrictWalker, _super);
    function UseStrictWalker() {
        _super.apply(this, arguments);
    }
    UseStrictWalker.prototype.createScope = function () {
        return {};
    };
    UseStrictWalker.prototype.visitModuleDeclaration = function (node) {
        if (!Lint.hasModifier(node.modifiers, ts.SyntaxKind.DeclareKeyword)
            && this.hasOption(UseStrictWalker.OPTION_CHECK_MODULE)
            && node.body != null
            && node.body.kind === ts.SyntaxKind.ModuleBlock) {
            var firstModuleDeclaration = getFirstInModuleDeclarationsChain(node);
            var hasOnlyModuleDeclarationParents = firstModuleDeclaration.parent.kind === ts.SyntaxKind.SourceFile;
            if (hasOnlyModuleDeclarationParents) {
                this.handleBlock(firstModuleDeclaration, node.body);
            }
        }
        _super.prototype.visitModuleDeclaration.call(this, node);
    };
    UseStrictWalker.prototype.visitFunctionDeclaration = function (node) {
        if (this.getCurrentDepth() === 2 &&
            this.hasOption(UseStrictWalker.OPTION_CHECK_FUNCTION) &&
            node.body != null) {
            this.handleBlock(node, node.body);
        }
        _super.prototype.visitFunctionDeclaration.call(this, node);
    };
    UseStrictWalker.prototype.handleBlock = function (node, block) {
        var isFailure = true;
        if (block.statements != null && block.statements.length > 0) {
            var firstStatement = block.statements[0];
            if (firstStatement.kind === ts.SyntaxKind.ExpressionStatement) {
                var firstChild = firstStatement.getChildAt(0);
                if (firstChild.kind === ts.SyntaxKind.StringLiteral
                    && firstChild.text === UseStrictWalker.USE_STRICT_STRING) {
                    isFailure = false;
                }
            }
        }
        if (isFailure) {
            this.addFailure(this.createFailure(node.getStart(), node.getFirstToken().getWidth(), Rule.FAILURE_STRING));
        }
    };
    UseStrictWalker.OPTION_CHECK_FUNCTION = "check-function";
    UseStrictWalker.OPTION_CHECK_MODULE = "check-module";
    UseStrictWalker.USE_STRICT_STRING = "use strict";
    return UseStrictWalker;
}(Lint.ScopeAwareRuleWalker));
function getFirstInModuleDeclarationsChain(node) {
    var current = node;
    while (current.parent.kind === ts.SyntaxKind.ModuleDeclaration) {
        current = current.parent;
    }
    return current;
}
