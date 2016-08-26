import * as configuration from "./configuration";
import * as formatters from "./formatters";
import { RuleFailure } from "./language/rule/rule";
import * as rules from "./rules";
import * as test from "./test";
import * as linter from "./tslint";
import * as utils from "./utils";
export * from "./language/rule/rule";
export * from "./enableDisableRules";
export * from "./formatterLoader";
export * from "./ruleLoader";
export * from "./language/utils";
export * from "./language/languageServiceHost";
export * from "./language/walker";
export * from "./language/formatter/formatter";
export declare var Configuration: typeof configuration;
export declare var Formatters: typeof formatters;
export declare var Linter: typeof linter;
export declare var Rules: typeof rules;
export declare var Test: typeof test;
export declare var Utils: typeof utils;
export interface LintResult {
    failureCount: number;
    failures: RuleFailure[];
    format: string | Function;
    output: string;
}
export interface ILinterOptionsRaw {
    configuration?: any;
    formatter?: string | Function;
    formattersDirectory?: string;
    rulesDirectory?: string | string[];
}
export interface ILinterOptions extends ILinterOptionsRaw {
    configuration: any;
    formatter: string | Function;
    rulesDirectory: string | string[];
}
