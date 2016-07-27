import Promise = require('any-promise');
export declare function help(): string;
export interface Options {
    cwd: string;
    save: boolean;
    saveDev: boolean;
    savePeer: boolean;
    global: boolean;
    verbose: boolean;
    help: boolean;
}
export declare function exec(args: string[], options: Options): Promise<void>;
