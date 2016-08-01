import * as ts from "typescript";
import { IOptions } from "../../lint";
import { RuleWalker } from "./ruleWalker";
export declare class SkippableTokenAwareRuleWalker extends RuleWalker {
    protected tokensToSkipStartEndMap: {
        [start: number]: number;
    };
    constructor(sourceFile: ts.SourceFile, options: IOptions);
    protected visitRegularExpressionLiteral(node: ts.Node): void;
    protected visitIdentifier(node: ts.Identifier): void;
    protected visitTemplateExpression(node: ts.TemplateExpression): void;
    protected addTokenToSkipFromNode(node: ts.Node): void;
}
