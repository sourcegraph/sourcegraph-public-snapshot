import Promise = require('any-promise');
export declare function help(): string;
export interface Options {
    versions: boolean;
}
export declare function exec(args: string[], options: Options): Promise<void>;
