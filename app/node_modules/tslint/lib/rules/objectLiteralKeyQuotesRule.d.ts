import * as Lint from "../lint";
import * as ts from "typescript";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    static UNNEEDED_QUOTES: (name: string) => string;
    static UNQUOTED_PROPERTY: (name: string) => string;
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
}
