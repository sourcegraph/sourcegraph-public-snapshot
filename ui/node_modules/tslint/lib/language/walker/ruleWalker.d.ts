import * as ts from "typescript";
import { IOptions } from "../../lint";
import { Fix, Replacement, RuleFailure } from "../rule/rule";
import { SyntaxWalker } from "./syntaxWalker";
export declare class RuleWalker extends SyntaxWalker {
    private sourceFile;
    private limit;
    private position;
    private options;
    private failures;
    private disabledIntervals;
    private ruleName;
    constructor(sourceFile: ts.SourceFile, options: IOptions);
    getSourceFile(): ts.SourceFile;
    getFailures(): RuleFailure[];
    getLimit(): number;
    getOptions(): any;
    hasOption(option: string): boolean;
    skip(node: ts.Node): void;
    createFailure(start: number, width: number, failure: string, fix?: Fix): RuleFailure;
    addFailure(failure: RuleFailure): void;
    createReplacement(start: number, length: number, text: string): Replacement;
    appendText(start: number, text: string): Replacement;
    deleteText(start: number, length: number): Replacement;
    private existsFailure(failure);
}
