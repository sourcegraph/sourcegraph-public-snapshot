import * as ts from "typescript";
import * as Lint from "../lint";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    static FAILURE_STRING_NEVER: string;
    static FAILURE_STRING_ALWAYS: string;
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
}
