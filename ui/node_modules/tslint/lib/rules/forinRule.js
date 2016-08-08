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
        return this.applyWithWalker(new ForInWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "forin",
        description: "Requires a `for ... in` statement to be filtered with an `if` statement.",
        rationale: (_a = ["\n            ```ts\n            for (let key in someObject) {\n                if (someObject.hasOwnProperty(key)) {\n                    // code here\n                }\n            }\n            ```\n            Prevents accidental interation over properties inherited from an object's prototype.\n            See [MDN's `for...in`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/for...in)\n            documentation for more information about `for...in` loops."], _a.raw = ["\n            \\`\\`\\`ts\n            for (let key in someObject) {\n                if (someObject.hasOwnProperty(key)) {\n                    // code here\n                }\n            }\n            \\`\\`\\`\n            Prevents accidental interation over properties inherited from an object's prototype.\n            See [MDN's \\`for...in\\`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/for...in)\n            documentation for more information about \\`for...in\\` loops."], Lint.Utils.dedent(_a)),
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING = "for (... in ...) statements must be filtered with an if statement";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var ForInWalker = (function (_super) {
    __extends(ForInWalker, _super);
    function ForInWalker() {
        _super.apply(this, arguments);
    }
    ForInWalker.prototype.visitForInStatement = function (node) {
        this.handleForInStatement(node);
        _super.prototype.visitForInStatement.call(this, node);
    };
    ForInWalker.prototype.handleForInStatement = function (node) {
        var statement = node.statement;
        var statementKind = node.statement.kind;
        if (statementKind === ts.SyntaxKind.IfStatement) {
            return;
        }
        if (statementKind === ts.SyntaxKind.Block) {
            var blockNode = statement;
            var blockStatements = blockNode.statements;
            if (blockStatements.length >= 1) {
                var firstBlockStatement = blockStatements[0];
                if (firstBlockStatement.kind === ts.SyntaxKind.IfStatement) {
                    if (blockStatements.length === 1) {
                        return;
                    }
                    var ifStatement = firstBlockStatement.thenStatement;
                    if (nodeIsContinue(ifStatement)) {
                        return;
                    }
                }
            }
        }
        var failure = this.createFailure(node.getStart(), node.getWidth(), Rule.FAILURE_STRING);
        this.addFailure(failure);
    };
    return ForInWalker;
}(Lint.RuleWalker));
function nodeIsContinue(node) {
    var kind = node.kind;
    if (kind === ts.SyntaxKind.ContinueStatement) {
        return true;
    }
    if (kind === ts.SyntaxKind.Block) {
        var blockStatements = node.statements;
        if (blockStatements.length === 1 && blockStatements[0].kind === ts.SyntaxKind.ContinueStatement) {
            return true;
        }
    }
    return false;
}
