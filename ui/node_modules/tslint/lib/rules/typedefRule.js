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
        return this.applyWithWalker(new TypedefWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "typedef",
        description: "Requires type definitions to exist.",
        optionsDescription: (_a = ["\n            Seven arguments may be optionally provided:\n\n            * `\"call-signature\"` checks return type of functions.\n            * `\"arrow-call-signature\"` checks return type of arrow functions.\n            * `\"parameter\"` checks type specifier of function parameters for non-arrow functions.\n            * `\"arrow-parameter\"` checks type specifier of function parameters for arrow functions.\n            * `\"property-declaration\"` checks return types of interface properties.\n            * `\"variable-declaration\"` checks variable declarations.\n            * `\"member-variable-declaration\"` checks member variable declarations."], _a.raw = ["\n            Seven arguments may be optionally provided:\n\n            * \\`\"call-signature\"\\` checks return type of functions.\n            * \\`\"arrow-call-signature\"\\` checks return type of arrow functions.\n            * \\`\"parameter\"\\` checks type specifier of function parameters for non-arrow functions.\n            * \\`\"arrow-parameter\"\\` checks type specifier of function parameters for arrow functions.\n            * \\`\"property-declaration\"\\` checks return types of interface properties.\n            * \\`\"variable-declaration\"\\` checks variable declarations.\n            * \\`\"member-variable-declaration\"\\` checks member variable declarations."], Lint.Utils.dedent(_a)),
        options: {
            type: "array",
            items: {
                type: "string",
                enum: [
                    "call-signature",
                    "arrow-call-signature",
                    "parameter",
                    "arrow-parameter",
                    "property-declaration",
                    "variable-declaration",
                    "member-variable-declaration",
                ],
            },
            minLength: 0,
            maxLength: 7,
        },
        optionExamples: ['[true, "call-signature", "parameter", "member-variable-declaration"]'],
        type: "typescript",
    };
    Rule.FAILURE_STRING = "missing type declaration";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var TypedefWalker = (function (_super) {
    __extends(TypedefWalker, _super);
    function TypedefWalker() {
        _super.apply(this, arguments);
    }
    TypedefWalker.prototype.visitFunctionDeclaration = function (node) {
        this.handleCallSignature(node);
        _super.prototype.visitFunctionDeclaration.call(this, node);
    };
    TypedefWalker.prototype.visitFunctionExpression = function (node) {
        this.handleCallSignature(node);
        _super.prototype.visitFunctionExpression.call(this, node);
    };
    TypedefWalker.prototype.visitArrowFunction = function (node) {
        var location = (node.parameters != null) ? node.parameters.end : null;
        if (node.parent.kind !== ts.SyntaxKind.CallExpression && !isTypedPropertyDeclaration(node.parent)) {
            this.checkTypeAnnotation("arrow-call-signature", location, node.type, node.name);
        }
        _super.prototype.visitArrowFunction.call(this, node);
    };
    TypedefWalker.prototype.visitGetAccessor = function (node) {
        this.handleCallSignature(node);
        _super.prototype.visitGetAccessor.call(this, node);
    };
    TypedefWalker.prototype.visitMethodDeclaration = function (node) {
        this.handleCallSignature(node);
        _super.prototype.visitMethodDeclaration.call(this, node);
    };
    TypedefWalker.prototype.visitMethodSignature = function (node) {
        this.handleCallSignature(node);
        _super.prototype.visitMethodSignature.call(this, node);
    };
    TypedefWalker.prototype.visitObjectLiteralExpression = function (node) {
        for (var _i = 0, _a = node.properties; _i < _a.length; _i++) {
            var property = _a[_i];
            switch (property.kind) {
                case ts.SyntaxKind.PropertyAssignment:
                    this.visitPropertyAssignment(property);
                    break;
                case ts.SyntaxKind.MethodDeclaration:
                    this.visitMethodDeclaration(property);
                    break;
                case ts.SyntaxKind.GetAccessor:
                    this.visitGetAccessor(property);
                    break;
                case ts.SyntaxKind.SetAccessor:
                    this.visitSetAccessor(property);
                    break;
                default:
                    break;
            }
        }
    };
    TypedefWalker.prototype.visitParameterDeclaration = function (node) {
        if (node.type == null || node.type.kind !== ts.SyntaxKind.StringLiteral) {
            var isArrowFunction = node.parent.kind === ts.SyntaxKind.ArrowFunction;
            var performCheck = true;
            var optionName = void 0;
            if (isArrowFunction && isTypedPropertyDeclaration(node.parent.parent)) {
                performCheck = false;
            }
            else if (isArrowFunction && isPropertyDeclaration(node.parent.parent)) {
                optionName = "member-variable-declaration";
            }
            else if (isArrowFunction) {
                optionName = "arrow-parameter";
            }
            else {
                optionName = "parameter";
            }
            if (performCheck) {
                this.checkTypeAnnotation(optionName, node.getEnd(), node.type, node.name);
            }
        }
        _super.prototype.visitParameterDeclaration.call(this, node);
    };
    TypedefWalker.prototype.visitPropertyAssignment = function (node) {
        switch (node.initializer.kind) {
            case ts.SyntaxKind.ArrowFunction:
            case ts.SyntaxKind.FunctionExpression:
                this.handleCallSignature(node.initializer);
                break;
            default:
                break;
        }
        _super.prototype.visitPropertyAssignment.call(this, node);
    };
    TypedefWalker.prototype.visitPropertyDeclaration = function (node) {
        var optionName = "member-variable-declaration";
        var performCheck = !(node.initializer != null && node.initializer.kind === ts.SyntaxKind.ArrowFunction && node.type == null);
        if (performCheck) {
            this.checkTypeAnnotation(optionName, node.name.getEnd(), node.type, node.name);
        }
        _super.prototype.visitPropertyDeclaration.call(this, node);
    };
    TypedefWalker.prototype.visitPropertySignature = function (node) {
        var optionName = "property-declaration";
        this.checkTypeAnnotation(optionName, node.name.getEnd(), node.type, node.name);
        _super.prototype.visitPropertySignature.call(this, node);
    };
    TypedefWalker.prototype.visitSetAccessor = function (node) {
        this.handleCallSignature(node);
        _super.prototype.visitSetAccessor.call(this, node);
    };
    TypedefWalker.prototype.visitVariableDeclaration = function (node) {
        if (node.parent != null && node.parent.parent != null
            && node.parent.kind !== ts.SyntaxKind.CatchClause
            && node.parent.parent.kind !== ts.SyntaxKind.ForInStatement
            && node.parent.parent.kind !== ts.SyntaxKind.ForOfStatement) {
            this.checkTypeAnnotation("variable-declaration", node.name.getEnd(), node.type, node.name);
        }
        _super.prototype.visitVariableDeclaration.call(this, node);
    };
    TypedefWalker.prototype.handleCallSignature = function (node) {
        var location = (node.parameters != null) ? node.parameters.end : null;
        if (node.kind !== ts.SyntaxKind.SetAccessor && node.kind !== ts.SyntaxKind.ArrowFunction) {
            this.checkTypeAnnotation("call-signature", location, node.type, node.name);
        }
    };
    TypedefWalker.prototype.checkTypeAnnotation = function (option, location, typeAnnotation, name) {
        if (this.hasOption(option) && typeAnnotation == null) {
            var ns = "";
            if (name != null && name.kind === ts.SyntaxKind.Identifier) {
                ns = ": '" + name.text + "'";
            }
            var failure = this.createFailure(location, 1, "expected " + option + ns + " to have a typedef");
            this.addFailure(failure);
        }
    };
    return TypedefWalker;
}(Lint.RuleWalker));
function isPropertyDeclaration(node) {
    return node.kind === ts.SyntaxKind.PropertyDeclaration;
}
function isTypedPropertyDeclaration(node) {
    return isPropertyDeclaration(node) && node.type != null;
}
