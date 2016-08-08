"use strict";
var __extends = (this && this.__extends) || function (d, b) {
    for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p];
    function __() { this.constructor = d; }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
};
var ts = require("typescript");
var Lint = require("../lint");
var OPTION_BRACE = "check-open-brace";
var OPTION_CATCH = "check-catch";
var OPTION_ELSE = "check-else";
var OPTION_FINALLY = "check-finally";
var OPTION_WHITESPACE = "check-whitespace";
var Rule = (function (_super) {
    __extends(Rule, _super);
    function Rule() {
        _super.apply(this, arguments);
    }
    Rule.prototype.apply = function (sourceFile) {
        var oneLineWalker = new OneLineWalker(sourceFile, this.getOptions());
        return this.applyWithWalker(oneLineWalker);
    };
    Rule.metadata = {
        ruleName: "one-line",
        description: "Requires the specified tokens to be on the same line as the expression preceding them.",
        optionsDescription: (_a = ["\n            Five arguments may be optionally provided:\n\n            * `\"check-catch\"` checks that `catch` is on the same line as the closing brace for `try`.\n            * `\"check-finally\"` checks that `finally` is on the same line as the closing brace for `catch`.\n            * `\"check-else\"` checks that `else` is on the same line as the closing brace for `if`.\n            * `\"check-open-brace\"` checks that an open brace falls on the same line as its preceding expression.\n            * `\"check-whitespace\"` checks preceding whitespace for the specified tokens."], _a.raw = ["\n            Five arguments may be optionally provided:\n\n            * \\`\"check-catch\"\\` checks that \\`catch\\` is on the same line as the closing brace for \\`try\\`.\n            * \\`\"check-finally\"\\` checks that \\`finally\\` is on the same line as the closing brace for \\`catch\\`.\n            * \\`\"check-else\"\\` checks that \\`else\\` is on the same line as the closing brace for \\`if\\`.\n            * \\`\"check-open-brace\"\\` checks that an open brace falls on the same line as its preceding expression.\n            * \\`\"check-whitespace\"\\` checks preceding whitespace for the specified tokens."], Lint.Utils.dedent(_a)),
        options: {
            type: "array",
            items: {
                type: "string",
                enum: ["check-catch", "check-finally", "check-else", "check-open-brace", "check-whitespace"],
            },
            minLength: 0,
            maxLength: 5,
        },
        optionExamples: ['[true, "check-catch", "check-finally", "check-else"]'],
        type: "style",
    };
    Rule.BRACE_FAILURE_STRING = "misplaced opening brace";
    Rule.CATCH_FAILURE_STRING = "misplaced 'catch'";
    Rule.ELSE_FAILURE_STRING = "misplaced 'else'";
    Rule.FINALLY_FAILURE_STRING = "misplaced 'finally'";
    Rule.WHITESPACE_FAILURE_STRING = "missing whitespace";
    return Rule;
    var _a;
}(Lint.Rules.AbstractRule));
exports.Rule = Rule;
var OneLineWalker = (function (_super) {
    __extends(OneLineWalker, _super);
    function OneLineWalker() {
        _super.apply(this, arguments);
    }
    OneLineWalker.prototype.visitIfStatement = function (node) {
        var sourceFile = node.getSourceFile();
        var thenStatement = node.thenStatement;
        if (thenStatement.kind === ts.SyntaxKind.Block) {
            var expressionCloseParen = node.getChildAt(3);
            var thenOpeningBrace = thenStatement.getChildAt(0);
            this.handleOpeningBrace(expressionCloseParen, thenOpeningBrace);
        }
        var elseStatement = node.elseStatement;
        if (elseStatement != null) {
            var elseKeyword = getFirstChildOfKind(node, ts.SyntaxKind.ElseKeyword);
            if (elseStatement.kind === ts.SyntaxKind.Block) {
                var elseOpeningBrace = elseStatement.getChildAt(0);
                this.handleOpeningBrace(elseKeyword, elseOpeningBrace);
            }
            if (this.hasOption(OPTION_ELSE)) {
                var thenStatementEndLine = sourceFile.getLineAndCharacterOfPosition(thenStatement.getEnd()).line;
                var elseKeywordLine = sourceFile.getLineAndCharacterOfPosition(elseKeyword.getStart()).line;
                if (thenStatementEndLine !== elseKeywordLine) {
                    var failure = this.createFailure(elseKeyword.getStart(), elseKeyword.getWidth(), Rule.ELSE_FAILURE_STRING);
                    this.addFailure(failure);
                }
            }
        }
        _super.prototype.visitIfStatement.call(this, node);
    };
    OneLineWalker.prototype.visitCatchClause = function (node) {
        var catchClosingParen = node.getChildren().filter(function (n) { return n.kind === ts.SyntaxKind.CloseParenToken; })[0];
        var catchOpeningBrace = node.block.getChildAt(0);
        this.handleOpeningBrace(catchClosingParen, catchOpeningBrace);
        _super.prototype.visitCatchClause.call(this, node);
    };
    OneLineWalker.prototype.visitTryStatement = function (node) {
        var sourceFile = node.getSourceFile();
        var catchClause = node.catchClause;
        var finallyBlock = node.finallyBlock;
        var finallyKeyword = node.getChildren().filter(function (n) { return n.kind === ts.SyntaxKind.FinallyKeyword; })[0];
        var tryKeyword = node.getChildAt(0);
        var tryBlock = node.tryBlock;
        var tryOpeningBrace = tryBlock.getChildAt(0);
        this.handleOpeningBrace(tryKeyword, tryOpeningBrace);
        if (this.hasOption(OPTION_CATCH) && catchClause != null) {
            var tryClosingBrace = node.tryBlock.getChildAt(node.tryBlock.getChildCount() - 1);
            var catchKeyword = catchClause.getChildAt(0);
            var tryClosingBraceLine = sourceFile.getLineAndCharacterOfPosition(tryClosingBrace.getEnd()).line;
            var catchKeywordLine = sourceFile.getLineAndCharacterOfPosition(catchKeyword.getStart()).line;
            if (tryClosingBraceLine !== catchKeywordLine) {
                var failure = this.createFailure(catchKeyword.getStart(), catchKeyword.getWidth(), Rule.CATCH_FAILURE_STRING);
                this.addFailure(failure);
            }
        }
        if (finallyBlock != null && finallyKeyword != null) {
            var finallyOpeningBrace = finallyBlock.getChildAt(0);
            this.handleOpeningBrace(finallyKeyword, finallyOpeningBrace);
            if (this.hasOption(OPTION_FINALLY)) {
                var previousBlock = catchClause != null ? catchClause.block : node.tryBlock;
                var closingBrace = previousBlock.getChildAt(previousBlock.getChildCount() - 1);
                var closingBraceLine = sourceFile.getLineAndCharacterOfPosition(closingBrace.getEnd()).line;
                var finallyKeywordLine = sourceFile.getLineAndCharacterOfPosition(finallyKeyword.getStart()).line;
                if (closingBraceLine !== finallyKeywordLine) {
                    var failure = this.createFailure(finallyKeyword.getStart(), finallyKeyword.getWidth(), Rule.FINALLY_FAILURE_STRING);
                    this.addFailure(failure);
                }
            }
        }
        _super.prototype.visitTryStatement.call(this, node);
    };
    OneLineWalker.prototype.visitForStatement = function (node) {
        this.handleIterationStatement(node);
        _super.prototype.visitForStatement.call(this, node);
    };
    OneLineWalker.prototype.visitForInStatement = function (node) {
        this.handleIterationStatement(node);
        _super.prototype.visitForInStatement.call(this, node);
    };
    OneLineWalker.prototype.visitWhileStatement = function (node) {
        this.handleIterationStatement(node);
        _super.prototype.visitWhileStatement.call(this, node);
    };
    OneLineWalker.prototype.visitBinaryExpression = function (node) {
        var rightkind = node.right.kind;
        var opkind = node.operatorToken.kind;
        if (opkind === ts.SyntaxKind.EqualsToken && rightkind === ts.SyntaxKind.ObjectLiteralExpression) {
            var equalsToken = node.getChildAt(1);
            var openBraceToken = node.right.getChildAt(0);
            this.handleOpeningBrace(equalsToken, openBraceToken);
        }
        _super.prototype.visitBinaryExpression.call(this, node);
    };
    OneLineWalker.prototype.visitVariableDeclaration = function (node) {
        var initializer = node.initializer;
        if (initializer != null && initializer.kind === ts.SyntaxKind.ObjectLiteralExpression) {
            var equalsToken = node.getChildren().filter(function (n) { return n.kind === ts.SyntaxKind.EqualsToken; })[0];
            var openBraceToken = initializer.getChildAt(0);
            this.handleOpeningBrace(equalsToken, openBraceToken);
        }
        _super.prototype.visitVariableDeclaration.call(this, node);
    };
    OneLineWalker.prototype.visitDoStatement = function (node) {
        var doKeyword = node.getChildAt(0);
        var statement = node.statement;
        if (statement.kind === ts.SyntaxKind.Block) {
            var openBraceToken = statement.getChildAt(0);
            this.handleOpeningBrace(doKeyword, openBraceToken);
        }
        _super.prototype.visitDoStatement.call(this, node);
    };
    OneLineWalker.prototype.visitModuleDeclaration = function (node) {
        var nameNode = node.name;
        var body = node.body;
        if (body.kind === ts.SyntaxKind.ModuleBlock) {
            var openBraceToken = body.getChildAt(0);
            this.handleOpeningBrace(nameNode, openBraceToken);
        }
        _super.prototype.visitModuleDeclaration.call(this, node);
    };
    OneLineWalker.prototype.visitEnumDeclaration = function (node) {
        var nameNode = node.name;
        var openBraceToken = getFirstChildOfKind(node, ts.SyntaxKind.OpenBraceToken);
        this.handleOpeningBrace(nameNode, openBraceToken);
        _super.prototype.visitEnumDeclaration.call(this, node);
    };
    OneLineWalker.prototype.visitSwitchStatement = function (node) {
        var closeParenToken = node.getChildAt(3);
        var openBraceToken = node.caseBlock.getChildAt(0);
        this.handleOpeningBrace(closeParenToken, openBraceToken);
        _super.prototype.visitSwitchStatement.call(this, node);
    };
    OneLineWalker.prototype.visitInterfaceDeclaration = function (node) {
        this.handleClassLikeDeclaration(node);
        _super.prototype.visitInterfaceDeclaration.call(this, node);
    };
    OneLineWalker.prototype.visitClassDeclaration = function (node) {
        this.handleClassLikeDeclaration(node);
        _super.prototype.visitClassDeclaration.call(this, node);
    };
    OneLineWalker.prototype.visitFunctionDeclaration = function (node) {
        this.handleFunctionLikeDeclaration(node);
        _super.prototype.visitFunctionDeclaration.call(this, node);
    };
    OneLineWalker.prototype.visitMethodDeclaration = function (node) {
        this.handleFunctionLikeDeclaration(node);
        _super.prototype.visitMethodDeclaration.call(this, node);
    };
    OneLineWalker.prototype.visitConstructorDeclaration = function (node) {
        this.handleFunctionLikeDeclaration(node);
        _super.prototype.visitConstructorDeclaration.call(this, node);
    };
    OneLineWalker.prototype.visitArrowFunction = function (node) {
        var body = node.body;
        if (body != null && body.kind === ts.SyntaxKind.Block) {
            var arrowToken = getFirstChildOfKind(node, ts.SyntaxKind.EqualsGreaterThanToken);
            var openBraceToken = node.body.getChildAt(0);
            this.handleOpeningBrace(arrowToken, openBraceToken);
        }
        _super.prototype.visitArrowFunction.call(this, node);
    };
    OneLineWalker.prototype.handleFunctionLikeDeclaration = function (node) {
        var body = node.body;
        if (body != null && body.kind === ts.SyntaxKind.Block) {
            var openBraceToken = node.body.getChildAt(0);
            if (node.type != null) {
                this.handleOpeningBrace(node.type, openBraceToken);
            }
            else {
                var closeParenToken = getFirstChildOfKind(node, ts.SyntaxKind.CloseParenToken);
                this.handleOpeningBrace(closeParenToken, openBraceToken);
            }
        }
    };
    OneLineWalker.prototype.handleClassLikeDeclaration = function (node) {
        var lastNodeOfDeclaration = node.name;
        var openBraceToken = getFirstChildOfKind(node, ts.SyntaxKind.OpenBraceToken);
        if (node.heritageClauses != null) {
            lastNodeOfDeclaration = node.heritageClauses[node.heritageClauses.length - 1];
        }
        else if (node.typeParameters != null) {
            lastNodeOfDeclaration = node.typeParameters[node.typeParameters.length - 1];
        }
        this.handleOpeningBrace(lastNodeOfDeclaration, openBraceToken);
    };
    OneLineWalker.prototype.handleIterationStatement = function (node) {
        var closeParenToken = node.getChildAt(node.getChildCount() - 2);
        var statement = node.statement;
        if (statement.kind === ts.SyntaxKind.Block) {
            var openBraceToken = statement.getChildAt(0);
            this.handleOpeningBrace(closeParenToken, openBraceToken);
        }
    };
    OneLineWalker.prototype.handleOpeningBrace = function (previousNode, openBraceToken) {
        if (previousNode == null || openBraceToken == null) {
            return;
        }
        var sourceFile = previousNode.getSourceFile();
        var previousNodeLine = sourceFile.getLineAndCharacterOfPosition(previousNode.getEnd()).line;
        var openBraceLine = sourceFile.getLineAndCharacterOfPosition(openBraceToken.getStart()).line;
        var failure;
        if (this.hasOption(OPTION_BRACE) && previousNodeLine !== openBraceLine) {
            failure = this.createFailure(openBraceToken.getStart(), openBraceToken.getWidth(), Rule.BRACE_FAILURE_STRING);
        }
        else if (this.hasOption(OPTION_WHITESPACE) && previousNode.getEnd() === openBraceToken.getStart()) {
            failure = this.createFailure(openBraceToken.getStart(), openBraceToken.getWidth(), Rule.WHITESPACE_FAILURE_STRING);
        }
        if (failure) {
            this.addFailure(failure);
        }
    };
    return OneLineWalker;
}(Lint.RuleWalker));
function getFirstChildOfKind(node, kind) {
    return node.getChildren().filter(function (child) { return child.kind === kind; })[0];
}
