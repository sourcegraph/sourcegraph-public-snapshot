import * as Lint from "../lint";
import * as ts from "typescript";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    static FAILURE_STRING_FACTORY: (lineCount: number, lineLimit: number) => string;
    isEnabled(): boolean;
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
}
