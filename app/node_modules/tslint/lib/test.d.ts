import { LintError } from "./test/lintError";
export interface TestResult {
    directory: string;
    results: {
        [fileName: string]: {
            errorsFromMarkup: LintError[];
            errorsFromLinter: LintError[];
            markupFromLinter: string;
            markupFromMarkup: string;
        };
    };
}
export declare function runTest(testDirectory: string, rulesDirectory?: string | string[]): TestResult;
export declare function consoleTestResultHandler(testResult: TestResult): boolean;
