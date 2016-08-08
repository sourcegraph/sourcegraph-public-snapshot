import * as ts from "typescript";
import { AbstractRule } from "./abstractRule";
import { RuleFailure } from "./rule";
export declare abstract class TypedRule extends AbstractRule {
    apply(sourceFile: ts.SourceFile): RuleFailure[];
    abstract applyWithProgram(sourceFile: ts.SourceFile, program: ts.Program): RuleFailure[];
}
