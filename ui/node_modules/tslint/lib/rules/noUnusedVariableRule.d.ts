import * as ts from "typescript";
import * as Lint from "../lint";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    static FAILURE_TYPE_FUNC: string;
    static FAILURE_TYPE_IMPORT: string;
    static FAILURE_TYPE_METHOD: string;
    static FAILURE_TYPE_PARAM: string;
    static FAILURE_TYPE_PROP: string;
    static FAILURE_TYPE_VAR: string;
    static FAILURE_STRING_FACTORY: (type: string, name: string) => string;
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
}
