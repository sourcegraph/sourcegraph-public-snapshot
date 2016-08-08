import * as ts from "typescript";
import { findConfiguration, findConfigurationPath, getRulesDirectories, loadConfigurationFromPath } from "./configuration";
import { ILinterOptionsRaw, LintResult } from "./lint";
declare class Linter {
    private fileName;
    private source;
    private program;
    static VERSION: string;
    static findConfiguration: typeof findConfiguration;
    static findConfigurationPath: typeof findConfigurationPath;
    static getRulesDirectories: typeof getRulesDirectories;
    static loadConfigurationFromPath: typeof loadConfigurationFromPath;
    private options;
    static createProgram(configFile: string, projectDirectory?: string): ts.Program;
    static getFileNames(program: ts.Program): string[];
    constructor(fileName: string, source: string, options: ILinterOptionsRaw, program?: ts.Program);
    lint(): LintResult;
    private containsRule(rules, rule);
    private computeFullOptions(options?);
}
declare namespace Linter {
}
export = Linter;
