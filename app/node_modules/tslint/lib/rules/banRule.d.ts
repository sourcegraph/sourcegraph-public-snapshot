import * as ts from "typescript";
import * as Lint from "../lint";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    static FAILURE_STRING_FACTORY: (expression: string, messageAddition?: string) => string;
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
}
export declare class BanFunctionWalker extends Lint.RuleWalker {
    private bannedFunctions;
    addBannedFunction(bannedFunction: string[]): void;
    visitCallExpression(node: ts.CallExpression): void;
}
