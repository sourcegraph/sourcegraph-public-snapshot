import Promise = require('any-promise');
export declare function help(): string;
export interface Options {
    cwd: string;
    name: string;
    out: string;
    global: boolean;
    verbose: boolean;
}
export declare function exec(args: string[], options: Options): Promise<any>;
