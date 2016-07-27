import * as ts from "typescript";
import * as Lint from "../lint";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    static LOWERCASE_FAILURE: string;
    static UPPERCASE_FAILURE: string;
    static LEADING_SPACE_FAILURE: string;
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
}
