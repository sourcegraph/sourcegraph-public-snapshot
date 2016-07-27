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
        return this.applyWithWalker(new MemberAccessWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "member-access",
        description: "Requires explicit visibility declarations for class members.",
        rationale: "Explicit visibility declarations can make code more readable and accessible for those new to TS.",
        optionsDescription: (_a = ["\n            Two arguments may be optionally provided:\n\n            * `\"check-accessor\"` enforces explicit visibility on get/set accessors (can only be public)\n            * `\"check-constructor\"`  enforces explicit visibility on constructors (can only be public)"], _a.raw = ["\n            Two arguments may be optionally provided:\n\n            * \\`\"check-accessor\"\\` enforces explicit visibility on get/set accessors (can only be public)\n            * \\`\"check-constructor\"\\`  enforces explicit visibility on constructors (can only be public)"], Lint.Utils.dedent(_a)),
        options: {
            type: "array",
            items: {
                type: "string",
                enum: ["check-accessor", "check-constructor"],
            },
            minLength: 0,
            maxLength: 2,
        },
        optionExamples: ["true", '[true, "check-accessor"]'],
        type: "typescript",
    };
    Rule.FAILURE_STRING_FACTORY = function (memberType, memberName, publicOnly) {
        memberName = memberName == null ? "" : " '" + memberName + "'";
        if (publicOnly) {
            return "The " + memberType + memberName + " must be marked as 'public'";
        }
        return "The " + memberType + memberName + " must be marked either 'private', 'public', or 'protected'";
    };
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var MemberAccessWalker = (function (_super) {
    __extends(MemberAccessWalker, _super);
    function MemberAccessWalker() {
        _super.apply(this, arguments);
    }
    MemberAccessWalker.prototype.visitConstructorDeclaration = function (node) {
        if (this.hasOption("check-constructor")) {
            this.validateVisibilityModifiers(node);
        }
        _super.prototype.visitConstructorDeclaration.call(this, node);
    };
    MemberAccessWalker.prototype.visitMethodDeclaration = function (node) {
        this.validateVisibilityModifiers(node);
        _super.prototype.visitMethodDeclaration.call(this, node);
    };
    MemberAccessWalker.prototype.visitPropertyDeclaration = function (node) {
        this.validateVisibilityModifiers(node);
        _super.prototype.visitPropertyDeclaration.call(this, node);
    };
    MemberAccessWalker.prototype.visitGetAccessor = function (node) {
        if (this.hasOption("check-accessor")) {
            this.validateVisibilityModifiers(node);
        }
        _super.prototype.visitGetAccessor.call(this, node);
    };
    MemberAccessWalker.prototype.visitSetAccessor = function (node) {
        if (this.hasOption("check-accessor")) {
            this.validateVisibilityModifiers(node);
        }
        _super.prototype.visitSetAccessor.call(this, node);
    };
    MemberAccessWalker.prototype.validateVisibilityModifiers = function (node) {
        if (node.parent.kind === ts.SyntaxKind.ObjectLiteralExpression) {
            return;
        }
        var hasAnyVisibilityModifiers = Lint.hasModifier(node.modifiers, ts.SyntaxKind.PublicKeyword, ts.SyntaxKind.PrivateKeyword, ts.SyntaxKind.ProtectedKeyword);
        if (!hasAnyVisibilityModifiers) {
            var memberType = void 0;
            var publicOnly = false;
            if (node.kind === ts.SyntaxKind.MethodDeclaration) {
                memberType = "class method";
            }
            else if (node.kind === ts.SyntaxKind.PropertyDeclaration) {
                memberType = "class property";
            }
            else if (node.kind === ts.SyntaxKind.Constructor) {
                memberType = "class constructor";
                publicOnly = true;
            }
            else if (node.kind === ts.SyntaxKind.GetAccessor) {
                memberType = "get property accessor";
            }
            else if (node.kind === ts.SyntaxKind.SetAccessor) {
                memberType = "set property accessor";
            }
            var memberName_1;
            node.getChildren().forEach(function (n) {
                if (n.kind === ts.SyntaxKind.Identifier) {
                    memberName_1 = n.getText();
                }
            });
            var failureString = Rule.FAILURE_STRING_FACTORY(memberType, memberName_1, publicOnly);
            this.addFailure(this.createFailure(node.getStart(), node.getWidth(), failureString));
        }
    };
    return MemberAccessWalker;
}(Lint.RuleWalker));
exports.MemberAccessWalker = MemberAccessWalker;
