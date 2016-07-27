import * as ts from "typescript";
import * as Lint from "../lint";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    static FAILURE_STRINGS: {
        CRLF: string;
        LF: string;
    };
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
    createFailure(sourceFile: ts.SourceFile, scanner: ts.Scanner, failure: string): Lint.RuleFailure;
}
