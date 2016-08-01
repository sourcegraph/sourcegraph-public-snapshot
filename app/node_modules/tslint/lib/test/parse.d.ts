import { LintError } from "./lintError";
export declare function removeErrorMarkup(text: string): string;
export declare function parseErrorsFromMarkup(text: string): LintError[];
export declare function createMarkupFromErrors(code: string, lintErrors: LintError[]): string;
