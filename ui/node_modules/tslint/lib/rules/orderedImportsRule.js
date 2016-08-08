"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var Lint = require("../lint");
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        var orderedImportsWalker = new OrderedImportsWalker(sourceFile, this.getOptions());
        return this.applyWithWalker(orderedImportsWalker);
    };
    Rule.metadata = {
        ruleName: "ordered-imports",
        description: "Requires that import statements be alphabetized.",
        descriptionDetails: (_a = ["\n            Enforce a consistent ordering for ES6 imports:\n            - Named imports must be alphabetized (i.e. \"import {A, B, C} from \"foo\";\")\n                - The exact ordering can be controled by the named-imports-order option.\n                - \"longName as name\" imports are ordered by \"longName\".\n            - Import sources must be alphabetized within groups, i.e.:\n                    import * as foo from \"a\";\n                    import * as bar from \"b\";\n            - Groups of imports are delineated by blank lines. You can use these to group imports\n                however you like, e.g. by first- vs. third-party or thematically."], _a.raw = ["\n            Enforce a consistent ordering for ES6 imports:\n            - Named imports must be alphabetized (i.e. \"import {A, B, C} from \"foo\";\")\n                - The exact ordering can be controled by the named-imports-order option.\n                - \"longName as name\" imports are ordered by \"longName\".\n            - Import sources must be alphabetized within groups, i.e.:\n                    import * as foo from \"a\";\n                    import * as bar from \"b\";\n            - Groups of imports are delineated by blank lines. You can use these to group imports\n                however you like, e.g. by first- vs. third-party or thematically."], Lint.Utils.dedent(_a)),
        optionsDescription: (_b = ["\n            You may set the `\"named-imports-order\"` option to control the ordering of named\n            imports (the `{A, B, C}` in 'import {A, B, C} from \"foo\"`.)\n\n            Possible values for `\"named-imports-order\"` are:\n\n            * `\"case-insensitive'`: Correct order is `{A, b, C}`. (This is the default.)\n            * `\"lowercase-first\"`: Correct order is `{b, A, C}`.\n            * `\"lowercase-last\"`: Correct order is `{A, C, b}`.\n        "], _b.raw = ["\n            You may set the \\`\"named-imports-order\"\\` option to control the ordering of named\n            imports (the \\`{A, B, C}\\` in \\'import {A, B, C} from \"foo\"\\`.)\n\n            Possible values for \\`\"named-imports-order\"\\` are:\n\n            * \\`\"case-insensitive'\\`: Correct order is \\`{A, b, C}\\`. (This is the default.)\n            * \\`\"lowercase-first\"\\`: Correct order is \\`{b, A, C}\\`.\n            * \\`\"lowercase-last\"\\`: Correct order is \\`{A, C, b}\\`.\n        "], Lint.Utils.dedent(_b)),
        options: {
            type: "object",
            properties: {
                "named-imports-order": {
                    type: "string",
                    enum: ["case-insensitive", "lowercase-first", "lowercase-last"],
                },
            },
            additionalProperties: false,
        },
        optionExamples: ["true", '[true, {"named-imports-order": "lowercase-first"}]'],
        type: "style",
    };
    Rule.NAMED_IMPORTS_UNORDERED = "Named imports must be alphabetized.";
    Rule.IMPORT_SOURCES_UNORDERED = "Import sources within a group must be alphabetized.";
    return Rule;
    var _a, _b;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
function flipCase(x) {
    return x.split("").map(function (char) {
        if (char >= "a" && char <= "z") {
            return char.toUpperCase();
        }
        else if (char >= "A" && char <= "Z") {
            return char.toLowerCase();
        }
        return char;
    }).join("");
}
function findUnsortedPair(xs, transform) {
    for (var i = 1; i < xs.length; i++) {
        if (transform(xs[i].getText()) < transform(xs[i - 1].getText())) {
            return [xs[i - 1], xs[i]];
        }
    }
    return null;
}
var TRANFORMS = {
    "case-insensitive": function (x) { return x.toLowerCase(); },
    "lowercase-first": flipCase,
    "lowercase-last": function (x) { return x; },
};
var OrderedImportsWalker = (function (_super) {
    __extends(OrderedImportsWalker, _super);
    function OrderedImportsWalker(sourceFile, options) {
        _super.call(this, sourceFile, options);
        this.lastImportSource = null;
        this.namedImportsOrder = null;
        var optionSet = this.getOptions()[0] || {};
        this.namedImportsOrder = optionSet["named-imports-order"] || "case-insensitive";
    }
    OrderedImportsWalker.prototype.visitImportDeclaration = function (node) {
        var source = node.moduleSpecifier.getText();
        if (this.lastImportSource && source < this.lastImportSource) {
            this.addFailure(this.createFailure(node.getStart(), node.getWidth(), Rule.IMPORT_SOURCES_UNORDERED));
        }
        this.lastImportSource = source;
        _super.prototype.visitImportDeclaration.call(this, node);
    };
    OrderedImportsWalker.prototype.visitNamedImports = function (node) {
        var imports = node.elements;
        var pair = findUnsortedPair(imports, TRANFORMS[this.namedImportsOrder]);
        if (pair !== null) {
            var a = pair[0], b = pair[1];
            this.addFailure(this.createFailure(a.getStart(), b.getEnd() - a.getStart(), Rule.NAMED_IMPORTS_UNORDERED));
        }
        _super.prototype.visitNamedImports.call(this, node);
    };
    OrderedImportsWalker.prototype.visitNode = function (node) {
        var prefixLength = node.getStart() - node.getFullStart();
        var prefix = node.getFullText().slice(0, prefixLength);
        if (prefix.indexOf("\n\n") >= 0 ||
            prefix.indexOf("\r\n\r\n") >= 0) {
            this.lastImportSource = null;
        }
        _super.prototype.visitNode.call(this, node);
    };
    return OrderedImportsWalker;
}(Lint.RuleWalker));
