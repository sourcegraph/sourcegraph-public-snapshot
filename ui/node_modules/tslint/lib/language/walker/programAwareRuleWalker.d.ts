import * as ts from "typescript";
import { IOptions } from "../../lint";
import { RuleWalker } from "./ruleWalker";
export declare class ProgramAwareRuleWalker extends RuleWalker {
    private program;
    private typeChecker;
    constructor(sourceFile: ts.SourceFile, options: IOptions, program: ts.Program);
    getProgram(): ts.Program;
    getTypeChecker(): ts.TypeChecker;
}
