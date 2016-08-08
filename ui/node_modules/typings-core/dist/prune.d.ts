import Promise = require('any-promise');
import { Emitter } from './interfaces';
export interface PruneOptions {
    cwd: string;
    production?: boolean;
    emitter?: Emitter;
}
export declare function prune(options: PruneOptions): Promise<void>;
export declare function rmDependency(options: {
    name: string;
    global: boolean;
    path: string;
    emitter: Emitter;
}): Promise<void>;
