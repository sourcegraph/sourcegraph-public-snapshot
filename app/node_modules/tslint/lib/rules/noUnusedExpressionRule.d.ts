import * as ts from "typescript";
import * as Lint from "../lint";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    static FAILURE_STRING: string;
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
}
export declare class NoUnusedExpressionWalker extends Lint.RuleWalker {
    protected expressionIsUnused: boolean;
    protected static isDirective(node: ts.Node, checkPreviousSiblings?: boolean): boolean;
    constructor(sourceFile: ts.SourceFile, options: Lint.IOptions);
    visitExpressionStatement(node: ts.ExpressionStatement): void;
    visitBinaryExpression(node: ts.BinaryExpression): void;
    visitPrefixUnaryExpression(node: ts.PrefixUnaryExpression): void;
    visitPostfixUnaryExpression(node: ts.PostfixUnaryExpression): void;
    visitBlock(node: ts.Block): void;
    visitArrowFunction(node: ts.FunctionLikeDeclaration): void;
    visitCallExpression(node: ts.CallExpression): void;
    protected visitNewExpression(node: ts.NewExpression): void;
    visitConditionalExpression(node: ts.ConditionalExpression): void;
    protected checkExpressionUsage(node: ts.ExpressionStatement): void;
}
