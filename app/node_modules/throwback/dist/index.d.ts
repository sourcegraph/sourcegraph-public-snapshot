import Promise = require('any-promise');
export interface NextFunction<T> {
    (): Promise<T>;
}
export declare function compose<T>(middleware: Array<(...args: any[]) => T | Promise<T>>): (...args: any[]) => Promise<T>;
