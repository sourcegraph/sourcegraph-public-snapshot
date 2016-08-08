import * as ts from "typescript";
import * as Lint from "../lint";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    static PARAMETERS_OPTION: string;
    static ARGUMENTS_OPTION: string;
    static STATEMENTS_OPTION: string;
    static FAILURE_STRING_SUFFIX: string;
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
}
