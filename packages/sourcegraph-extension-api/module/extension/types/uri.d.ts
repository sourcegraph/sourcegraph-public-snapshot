import * as sourcegraph from 'sourcegraph';
export declare class URI implements sourcegraph.URI {
    private value;
    static parse(uri: string): sourcegraph.URI;
    static file(path: string): sourcegraph.URI;
    static isURI(value: any): value is sourcegraph.URI;
    constructor(value: string);
    toString(): string;
    toJSON(): any;
}
