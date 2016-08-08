import * as Lint from "../lint";
import * as ts from "typescript";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    static FAILURE_TYPE_BREAK: string;
    static FAILURE_TYPE_CONTINUE: string;
    static FAILURE_TYPE_RETURN: string;
    static FAILURE_TYPE_THROW: string;
    static FAILURE_STRING_FACTORY: (name: string) => string;
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
}
