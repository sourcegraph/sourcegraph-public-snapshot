import * as ts from "typescript";
import * as Lint from "../lint";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    static IMPORT_SOURCES_UNORDERED: string;
    static NAMED_IMPORTS_UNORDERED: string;
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
}
