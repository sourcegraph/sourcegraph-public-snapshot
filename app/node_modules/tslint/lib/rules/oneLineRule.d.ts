import * as ts from "typescript";
import * as Lint from "../lint";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    static BRACE_FAILURE_STRING: string;
    static CATCH_FAILURE_STRING: string;
    static ELSE_FAILURE_STRING: string;
    static FINALLY_FAILURE_STRING: string;
    static WHITESPACE_FAILURE_STRING: string;
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
}
