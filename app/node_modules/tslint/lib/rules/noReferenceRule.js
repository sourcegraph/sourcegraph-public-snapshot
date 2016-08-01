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
        return this.applyWithWalker(new NoReferenceWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-reference",
        description: "Disallows `/// <reference path=>` imports (use ES6-style imports instead).",
        rationale: (_a = ["\n            Using `/// <reference path=>` comments to load other files is outdated.\n            Use ES6-style imports to reference other files."], _a.raw = ["\n            Using \\`/// <reference path=>\\` comments to load other files is outdated.\n            Use ES6-style imports to reference other files."], Lint.Utils.dedent(_a)),
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "typescript",
    };
    Rule.FAILURE_STRING = "<reference> is not allowed, use imports";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoReferenceWalker = (function (_super) {
    __extends(NoReferenceWalker, _super);
    function NoReferenceWalker() {
        _super.apply(this, arguments);
    }
    NoReferenceWalker.prototype.visitSourceFile = function (node) {
        for (var _i = 0, _a = node.referencedFiles; _i < _a.length; _i++) {
            var ref = _a[_i];
            this.addFailure(this.createFailure(ref.pos, ref.end - ref.pos, Rule.FAILURE_STRING));
        }
    };
    return NoReferenceWalker;
}(Lint.RuleWalker));
