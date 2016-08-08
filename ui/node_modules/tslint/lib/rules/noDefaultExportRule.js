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
        return this.applyWithWalker(new NoDefaultExportWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-default-export",
        description: "Disallows default exports in ES6-style modules.",
        descriptionDetails: "Use named exports instead.",
        rationale: (_a = ["\n            Named imports/exports [promote clarity](https://github.com/palantir/tslint/issues/1182#issue-151780453).\n            In addition, current tooling differs on the correct way to handle default imports/exports.\n            Avoiding them all together can help avoid tooling bugs and conflicts."], _a.raw = ["\n            Named imports/exports [promote clarity](https://github.com/palantir/tslint/issues/1182#issue-151780453).\n            In addition, current tooling differs on the correct way to handle default imports/exports.\n            Avoiding them all together can help avoid tooling bugs and conflicts."], Lint.Utils.dedent(_a)),
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "maintainability",
    };
    Rule.FAILURE_STRING = "Use of default exports is forbidden";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoDefaultExportWalker = (function (_super) {
    __extends(NoDefaultExportWalker, _super);
    function NoDefaultExportWalker() {
        _super.apply(this, arguments);
    }
    NoDefaultExportWalker.prototype.visitExportAssignment = function (node) {
        var exportMember = node.getChildAt(1);
        if (exportMember != null && exportMember.kind === ts.SyntaxKind.DefaultKeyword) {
            this.addFailure(this.createFailure(exportMember.getStart(), exportMember.getWidth(), Rule.FAILURE_STRING));
        }
        _super.prototype.visitExportAssignment.call(this, node);
    };
    NoDefaultExportWalker.prototype.visitNode = function (node) {
        if (node.kind === ts.SyntaxKind.DefaultKeyword && node.parent != null) {
            var nodes = node.parent.modifiers;
            if (nodes != null &&
                nodes.length === 2 &&
                nodes[0].kind === ts.SyntaxKind.ExportKeyword &&
                nodes[1].kind === ts.SyntaxKind.DefaultKeyword) {
                this.addFailure(this.createFailure(nodes[1].getStart(), nodes[1].getWidth(), Rule.FAILURE_STRING));
            }
        }
        _super.prototype.visitNode.call(this, node);
    };
    return NoDefaultExportWalker;
}(Lint.RuleWalker));
