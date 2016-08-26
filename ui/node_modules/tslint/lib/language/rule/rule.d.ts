import * as ts from "typescript";
import { RuleWalker } from "../walker/ruleWalker";
export interface IRuleMetadata {
    ruleName: string;
    type: RuleType;
    description: string;
    descriptionDetails?: string;
    optionsDescription?: string;
    options: any;
    optionExamples?: string[];
    rationale?: string;
    requiresTypeInfo?: boolean;
}
export declare type RuleType = "functionality" | "maintainability" | "style" | "typescript";
export interface IOptions {
    ruleArguments?: any[];
    ruleName: string;
    disabledIntervals: IDisabledInterval[];
}
export interface IDisabledInterval {
    startPosition: number;
    endPosition: number;
}
export interface IRule {
    getOptions(): IOptions;
    isEnabled(): boolean;
    apply(sourceFile: ts.SourceFile): RuleFailure[];
    applyWithWalker(walker: RuleWalker): RuleFailure[];
}
export declare class Replacement {
    private innerStart;
    private innerLength;
    private innerText;
    static applyAll(content: string, replacements: Replacement[]): string;
    constructor(innerStart: number, innerLength: number, innerText: string);
    start: number;
    length: number;
    end: number;
    text: string;
    apply(content: string): string;
}
export declare class Fix {
    private innerRuleName;
    private innerReplacements;
    static applyAll(content: string, fixes: Fix[]): string;
    constructor(innerRuleName: string, innerReplacements: Replacement[]);
    ruleName: string;
    replacements: Replacement[];
    apply(content: string): string;
}
export declare class RuleFailurePosition {
    private position;
    private lineAndCharacter;
    constructor(position: number, lineAndCharacter: ts.LineAndCharacter);
    getPosition(): number;
    getLineAndCharacter(): ts.LineAndCharacter;
    toJson(): {
        character: number;
        line: number;
        position: number;
    };
    equals(ruleFailurePosition: RuleFailurePosition): boolean;
}
export declare class RuleFailure {
    private sourceFile;
    private failure;
    private ruleName;
    private fix;
    private fileName;
    private startPosition;
    private endPosition;
    constructor(sourceFile: ts.SourceFile, start: number, end: number, failure: string, ruleName: string, fix?: Fix);
    getFileName(): string;
    getRuleName(): string;
    getStartPosition(): RuleFailurePosition;
    getEndPosition(): RuleFailurePosition;
    getFailure(): string;
    hasFix(): boolean;
    getFix(): Fix;
    toJson(): any;
    equals(ruleFailure: RuleFailure): boolean;
    private createFailurePosition(position);
}
