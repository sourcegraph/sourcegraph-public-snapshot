import Promise = require('any-promise');
import { Emitter } from 'typings-core';
export declare function help(): string;
export interface Options {
    verbose: boolean;
    save: boolean;
    saveDev: boolean;
    savePeer: boolean;
    global: boolean;
    emitter: Emitter;
    production: boolean;
    cwd: string;
    name?: string;
    source?: string;
    ambient: boolean;
}
export declare function exec(args: string[], options: Options): Promise<void>;
