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
        return this.applyWithWalker(new NoSwitchCaseFallThroughWalker(sourceFile, this.getOptions()));
    };
    Rule.metadata = {
        ruleName: "no-switch-case-fall-through",
        description: "Disallows falling through case statements.",
        descriptionDetails: (_a = ["\n            For example, the following is not allowed:\n\n            ```ts\n            switch(foo) {\n                case 1:\n                    someFunc(foo);\n                case 2:\n                    someOtherFunc(foo);\n            }\n            ```\n\n            However, fall through is allowed when case statements are consecutive or\n            a magic `/* falls through */` comment is present. The following is valid:\n\n            ```ts\n            switch(foo) {\n                case 1:\n                    someFunc(foo);\n                    /* falls through */\n                case 2:\n                case 3:\n                    someOtherFunc(foo);\n            }\n            ```"], _a.raw = ["\n            For example, the following is not allowed:\n\n            \\`\\`\\`ts\n            switch(foo) {\n                case 1:\n                    someFunc(foo);\n                case 2:\n                    someOtherFunc(foo);\n            }\n            \\`\\`\\`\n\n            However, fall through is allowed when case statements are consecutive or\n            a magic \\`/* falls through */\\` comment is present. The following is valid:\n\n            \\`\\`\\`ts\n            switch(foo) {\n                case 1:\n                    someFunc(foo);\n                    /* falls through */\n                case 2:\n                case 3:\n                    someOtherFunc(foo);\n            }\n            \\`\\`\\`"], Lint.Utils.dedent(_a)),
        rationale: "Fall though in switch statements is often unintentional and a bug.",
        optionsDescription: "Not configurable.",
        options: null,
        optionExamples: ["true"],
        type: "functionality",
    };
    Rule.FAILURE_STRING_PART = "expected a 'break' before ";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var NoSwitchCaseFallThroughWalker = (function (_super) {
    __extends(NoSwitchCaseFallThroughWalker, _super);
    function NoSwitchCaseFallThroughWalker() {
        _super.apply(this, arguments);
    }
    NoSwitchCaseFallThroughWalker.prototype.visitSwitchStatement = function (node) {
        var _this = this;
        var isFallingThrough = false;
        var switchClauses = node.caseBlock.clauses;
        switchClauses.forEach(function (child, i) {
            var kind = child.kind;
            if (kind === ts.SyntaxKind.CaseClause) {
                var switchClause = child;
                isFallingThrough = fallsThrough(switchClause.statements);
                if (isFallingThrough && switchClause.statements.length > 0 && ((switchClauses.length - 1) > i)) {
                    if (!isFallThroughAllowed(switchClauses[i + 1])) {
                        _this.addFailure(_this.createFailure(switchClauses[i + 1].getStart(), "case".length, Rule.FAILURE_STRING_PART + "'case'"));
                    }
                }
            }
            else {
                if (isFallingThrough && !isFallThroughAllowed(child)) {
                    var failureString = Rule.FAILURE_STRING_PART + "'default'";
                    _this.addFailure(_this.createFailure(switchClauses[i].getStart(), "default".length, failureString));
                }
            }
        });
        _super.prototype.visitSwitchStatement.call(this, node);
    };
    return NoSwitchCaseFallThroughWalker;
}(Lint.RuleWalker));
exports.NoSwitchCaseFallThroughWalker = NoSwitchCaseFallThroughWalker;
function fallsThrough(statements) {
    return !statements.some(function (statement) {
        return statement.kind === ts.SyntaxKind.BreakStatement
            || statement.kind === ts.SyntaxKind.ThrowStatement
            || statement.kind === ts.SyntaxKind.ReturnStatement
            || statement.kind === ts.SyntaxKind.ContinueStatement;
    });
}
function isFallThroughAllowed(nextCaseOrDefaultStatement) {
    var sourceFileText = nextCaseOrDefaultStatement.getSourceFile().text;
    var firstChild = nextCaseOrDefaultStatement.getChildAt(0);
    var commentRanges = ts.getLeadingCommentRanges(sourceFileText, firstChild.getFullStart());
    if (commentRanges != null) {
        for (var _i = 0, commentRanges_1 = commentRanges; _i < commentRanges_1.length; _i++) {
            var commentRange = commentRanges_1[_i];
            var commentText = sourceFileText.substring(commentRange.pos, commentRange.end);
            if (commentText === "/* falls through */") {
                return true;
            }
        }
    }
    return false;
}
