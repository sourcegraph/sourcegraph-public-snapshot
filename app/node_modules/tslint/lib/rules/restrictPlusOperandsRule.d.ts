import * as ts from "typescript";
import * as Lint from "../lint";
export declare class Rule extends Lint.Rules.TypedRule {
    static metadata: Lint.IRuleMetadata;
    static MISMATCHED_TYPES_FAILURE: string;
    static UNSUPPORTED_TYPE_FAILURE_FACTORY: (type: string) => string;
    applyWithProgram(sourceFile: ts.SourceFile, program: ts.Program): Lint.RuleFailure[];
}
