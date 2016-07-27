"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ts = require("typescript");
var Lint = require("../lint");
var OPTION_VARIABLES_BEFORE_FUNCTIONS = "variables-before-functions";
var OPTION_STATIC_BEFORE_INSTANCE = "static-before-instance";
var OPTION_PUBLIC_BEFORE_PRIVATE = "public-before-private";
var OPTION_ORDER = "order";
var PRESET_ORDERS = {
    "fields-first": [
        "public-static-field",
        "protected-static-field",
        "private-static-field",
        "public-instance-field",
        "protected-instance-field",
        "private-instance-field",
        "constructor",
        "public-static-method",
        "protected-static-method",
        "private-static-method",
        "public-instance-method",
        "protected-instance-method",
        "private-instance-method",
    ],
    "statics-first": [
        "public-static-field",
        "public-static-method",
        "protected-static-field",
        "protected-static-method",
        "private-static-field",
        "private-static-method",
        "public-instance-field",
        "protected-instance-field",
        "private-instance-field",
        "constructor",
        "public-instance-method",
        "protected-instance-method",
        "private-instance-method",
    ],
    "instance-sandwich": [
        "public-static-field",
        "protected-static-field",
        "private-static-field",
        "public-instance-field",
        "protected-instance-field",
        "private-instance-field",
        "constructor",
        "public-instance-method",
        "protected-instance-method",
        "private-instance-method",
        "public-static-method",
        "protected-static-method",
        "private-static-method",
    ],
};
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        return this.applyWithWalker(new MemberOrderingWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "member-ordering",
        description: "Enforces member ordering.",
        rationale: "A consistent ordering for class members can make classes easier to read, navigate, and edit.",
        optionsDescription: (_a = ["\n            One argument, which is an object, must be provided. It should contain an `order` property.\n            The `order` property should have a value of one of the following strings:\n\n            * `fields-first`\n            * `statics-first`\n            * `instance-sandwich`\n\n            Alternatively, the value for `order` maybe be an array consisting of the following strings:\n\n            * `public-static-field`\n            * `protected-static-field`\n            * `private-static-field`\n            * `public-instance-field`\n            * `protected-instance-field`\n            * `private-instance-field`\n            * `constructor`\n            * `public-static-method`\n            * `protected-static-method`\n            * `private-static-method`\n            * `public-instance-method`\n            * `protected-instance-method`\n            * `private-instance-method`\n\n            This is useful if one of the preset orders does not meet your needs."], _a.raw = ["\n            One argument, which is an object, must be provided. It should contain an \\`order\\` property.\n            The \\`order\\` property should have a value of one of the following strings:\n\n            * \\`fields-first\\`\n            * \\`statics-first\\`\n            * \\`instance-sandwich\\`\n\n            Alternatively, the value for \\`order\\` maybe be an array consisting of the following strings:\n\n            * \\`public-static-field\\`\n            * \\`protected-static-field\\`\n            * \\`private-static-field\\`\n            * \\`public-instance-field\\`\n            * \\`protected-instance-field\\`\n            * \\`private-instance-field\\`\n            * \\`constructor\\`\n            * \\`public-static-method\\`\n            * \\`protected-static-method\\`\n            * \\`private-static-method\\`\n            * \\`public-instance-method\\`\n            * \\`protected-instance-method\\`\n            * \\`private-instance-method\\`\n\n            This is useful if one of the preset orders does not meet your needs."], Lint.Utils.dedent(_a)),
        options: {
            type: "object",
            properties: {
                order: {
                    oneOf: [{
                            type: "string",
                            enum: ["fields-first", "statics-first", "instance-sandwich"],
                        }, {
                            type: "array",
                            items: {
                                type: "string",
                                enum: PRESET_ORDERS["statics-first"],
                            },
                            maxLength: 13,
                        }],
                },
            },
            additionalProperties: false,
        },
        optionExamples: ['[true, { "order": "fields-first" }]'],
        type: "typescript",
    };
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
function getModifiers(isMethod, modifiers) {
    return {
        isInstance: !Lint.hasModifier(modifiers, ts.SyntaxKind.StaticKeyword),
        isMethod: isMethod,
        isPrivate: Lint.hasModifier(modifiers, ts.SyntaxKind.PrivateKeyword),
    };
}
function toString(modifiers) {
    return [
        modifiers.isPrivate ? "private" : "public",
        modifiers.isInstance ? "instance" : "static",
        "member",
        modifiers.isMethod ? "function" : "variable",
    ].join(" ");
}
var AccessLevel;
(function (AccessLevel) {
    AccessLevel[AccessLevel["PRIVATE"] = 0] = "PRIVATE";
    AccessLevel[AccessLevel["PROTECTED"] = 1] = "PROTECTED";
    AccessLevel[AccessLevel["PUBLIC"] = 2] = "PUBLIC";
})(AccessLevel || (AccessLevel = {}));
var Membership;
(function (Membership) {
    Membership[Membership["INSTANCE"] = 0] = "INSTANCE";
    Membership[Membership["STATIC"] = 1] = "STATIC";
})(Membership || (Membership = {}));
var Kind;
(function (Kind) {
    Kind[Kind["FIELD"] = 0] = "FIELD";
    Kind[Kind["METHOD"] = 1] = "METHOD";
})(Kind || (Kind = {}));
function getNodeAndModifiers(node, isMethod, isConstructor) {
    if (isConstructor === void 0) { isConstructor = false; }
    var modifiers = node.modifiers;
    var accessLevel = Lint.hasModifier(modifiers, ts.SyntaxKind.PrivateKeyword) ? AccessLevel.PRIVATE
        : Lint.hasModifier(modifiers, ts.SyntaxKind.ProtectedKeyword) ? AccessLevel.PROTECTED
            : AccessLevel.PUBLIC;
    var kind = isMethod ? Kind.METHOD : Kind.FIELD;
    var membership = Lint.hasModifier(modifiers, ts.SyntaxKind.StaticKeyword) ? Membership.STATIC : Membership.INSTANCE;
    return {
        accessLevel: accessLevel,
        isConstructor: isConstructor,
        kind: kind,
        membership: membership,
        node: node,
    };
}
function getNodeOption(_a) {
    var accessLevel = _a.accessLevel, isConstructor = _a.isConstructor, kind = _a.kind, membership = _a.membership;
    if (isConstructor) {
        return "constructor";
    }
    return [
        AccessLevel[accessLevel].toLowerCase(),
        Membership[membership].toLowerCase(),
        Kind[kind].toLowerCase(),
    ].join("-");
}
var MemberOrderingWalker = (function (_super) {
    __extends(MemberOrderingWalker, _super);
    function MemberOrderingWalker() {
        _super.apply(this, arguments);
        this.memberStack = [];
        this.hasOrderOption = this.getHasOrderOption();
    }
    MemberOrderingWalker.prototype.visitClassDeclaration = function (node) {
        this.resetPreviousModifiers();
        this.newMemberList();
        _super.prototype.visitClassDeclaration.call(this, node);
        this.checkMemberOrder();
    };
    MemberOrderingWalker.prototype.visitClassExpression = function (node) {
        this.resetPreviousModifiers();
        this.newMemberList();
        _super.prototype.visitClassExpression.call(this, node);
        this.checkMemberOrder();
    };
    MemberOrderingWalker.prototype.visitInterfaceDeclaration = function (node) {
        this.resetPreviousModifiers();
        this.newMemberList();
        _super.prototype.visitInterfaceDeclaration.call(this, node);
        this.checkMemberOrder();
    };
    MemberOrderingWalker.prototype.visitMethodDeclaration = function (node) {
        this.checkModifiersAndSetPrevious(node, getModifiers(true, node.modifiers));
        this.pushMember(getNodeAndModifiers(node, true));
        _super.prototype.visitMethodDeclaration.call(this, node);
    };
    MemberOrderingWalker.prototype.visitMethodSignature = function (node) {
        this.checkModifiersAndSetPrevious(node, getModifiers(true, node.modifiers));
        this.pushMember(getNodeAndModifiers(node, true));
        _super.prototype.visitMethodSignature.call(this, node);
    };
    MemberOrderingWalker.prototype.visitConstructorDeclaration = function (node) {
        this.checkModifiersAndSetPrevious(node, getModifiers(true, node.modifiers));
        this.pushMember(getNodeAndModifiers(node, true, true));
        _super.prototype.visitConstructorDeclaration.call(this, node);
    };
    MemberOrderingWalker.prototype.visitPropertyDeclaration = function (node) {
        var initializer = node.initializer;
        var isFunction = initializer != null
            && (initializer.kind === ts.SyntaxKind.ArrowFunction || initializer.kind === ts.SyntaxKind.FunctionExpression);
        this.checkModifiersAndSetPrevious(node, getModifiers(isFunction, node.modifiers));
        this.pushMember(getNodeAndModifiers(node, isFunction));
        _super.prototype.visitPropertyDeclaration.call(this, node);
    };
    MemberOrderingWalker.prototype.visitPropertySignature = function (node) {
        this.checkModifiersAndSetPrevious(node, getModifiers(false, node.modifiers));
        this.pushMember(getNodeAndModifiers(node, false));
        _super.prototype.visitPropertySignature.call(this, node);
    };
    MemberOrderingWalker.prototype.visitTypeLiteral = function (node) {
    };
    MemberOrderingWalker.prototype.visitObjectLiteralExpression = function (node) {
    };
    MemberOrderingWalker.prototype.resetPreviousModifiers = function () {
        this.previousMember = {
            isInstance: false,
            isMethod: false,
            isPrivate: false,
        };
    };
    MemberOrderingWalker.prototype.checkModifiersAndSetPrevious = function (node, currentMember) {
        if (!this.canAppearAfter(this.previousMember, currentMember)) {
            var failure = this.createFailure(node.getStart(), node.getWidth(), "Declaration of " + toString(currentMember) + " not allowed to appear after declaration of " + toString(this.previousMember));
            this.addFailure(failure);
        }
        this.previousMember = currentMember;
    };
    MemberOrderingWalker.prototype.canAppearAfter = function (previousMember, currentMember) {
        if (previousMember == null || currentMember == null) {
            return true;
        }
        if (this.hasOption(OPTION_VARIABLES_BEFORE_FUNCTIONS) && previousMember.isMethod !== currentMember.isMethod) {
            return Number(previousMember.isMethod) < Number(currentMember.isMethod);
        }
        if (this.hasOption(OPTION_STATIC_BEFORE_INSTANCE) && previousMember.isInstance !== currentMember.isInstance) {
            return Number(previousMember.isInstance) < Number(currentMember.isInstance);
        }
        if (this.hasOption(OPTION_PUBLIC_BEFORE_PRIVATE) && previousMember.isPrivate !== currentMember.isPrivate) {
            return Number(previousMember.isPrivate) < Number(currentMember.isPrivate);
        }
        return true;
    };
    MemberOrderingWalker.prototype.newMemberList = function () {
        if (this.hasOrderOption) {
            this.memberStack.push([]);
        }
    };
    MemberOrderingWalker.prototype.pushMember = function (node) {
        if (this.hasOrderOption) {
            this.memberStack[this.memberStack.length - 1].push(node);
        }
    };
    MemberOrderingWalker.prototype.checkMemberOrder = function () {
        var _this = this;
        if (this.hasOrderOption) {
            var memberList_1 = this.memberStack.pop();
            var order_1 = this.getOrder();
            var memberRank_1 = memberList_1.map(function (n) { return order_1.indexOf(getNodeOption(n)); });
            var prevRank_1 = -1;
            memberRank_1.forEach(function (rank, i) {
                if (rank === -1) {
                    return;
                }
                if (rank < prevRank_1) {
                    var node = memberList_1[i].node;
                    var nodeType = order_1[rank].split("-").join(" ");
                    var prevNodeType = order_1[prevRank_1].split("-").join(" ");
                    var lowerRanks = memberRank_1.filter(function (r) { return r < rank && r !== -1; }).sort();
                    var locationHint = lowerRanks.length > 0
                        ? "after " + order_1[lowerRanks[lowerRanks.length - 1]].split("-").join(" ") + "s"
                        : "at the beginning of the class/interface";
                    var errorLine1 = ("Declaration of " + nodeType + " not allowed after declaration of " + prevNodeType + ". ") +
                        ("Instead, this should come " + locationHint + ".");
                    _this.addFailure(_this.createFailure(node.getStart(), node.getWidth(), errorLine1));
                }
                else {
                    prevRank_1 = rank;
                }
            });
        }
    };
    MemberOrderingWalker.prototype.getHasOrderOption = function () {
        var allOptions = this.getOptions();
        if (allOptions == null || allOptions.length === 0) {
            return false;
        }
        var firstOption = allOptions[0];
        return firstOption != null && typeof firstOption === "object" && firstOption[OPTION_ORDER] != null;
    };
    MemberOrderingWalker.prototype.getOrder = function () {
        var orderOption = this.getOptions()[0][OPTION_ORDER];
        if (Array.isArray(orderOption)) {
            return orderOption;
        }
        else if (typeof orderOption === "string") {
            return PRESET_ORDERS[orderOption] || PRESET_ORDERS["default"];
        }
    };
    return MemberOrderingWalker;
}(Lint.RuleWalker));
exports.MemberOrderingWalker = MemberOrderingWalker;
