import * as ts from "typescript";
import * as Lint from "../lint";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    static DO_FAILURE_STRING: string;
    static ELSE_FAILURE_STRING: string;
    static FOR_FAILURE_STRING: string;
    static IF_FAILURE_STRING: string;
    static WHILE_FAILURE_STRING: string;
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
}
