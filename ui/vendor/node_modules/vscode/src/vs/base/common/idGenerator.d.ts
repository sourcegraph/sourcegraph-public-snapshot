export declare class IdGenerator {
    private _prefix;
    private _lastId;
    constructor(prefix: string);
    nextId(): string;
}
export declare const defaultGenerator: IdGenerator;
