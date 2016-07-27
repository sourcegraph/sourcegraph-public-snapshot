export declare class Line {
}
export declare class CodeLine extends Line {
    contents: string;
    constructor(contents: string);
}
export declare class MessageSubstitutionLine extends Line {
    key: string;
    message: string;
    constructor(key: string, message: string);
}
export declare class ErrorLine extends Line {
    startCol: number;
    constructor(startCol: number);
}
export declare class MultilineErrorLine extends ErrorLine {
    constructor(startCol: number);
}
export declare class EndErrorLine extends ErrorLine {
    endCol: number;
    message: string;
    constructor(startCol: number, endCol: number, message: string);
}
export declare const ZERO_LENGTH_ERROR: string;
export declare function parseLine(text: string): Line;
export declare function printLine(line: Line, code?: string): string;
