import Promise = require('any-promise');
import { Emitter } from './interfaces';
import { InstallResult } from './install';
export interface BundleOptions {
    name?: string;
    cwd: string;
    global?: boolean;
    resolution?: string;
    out: string;
    emitter?: Emitter;
}
export declare function bundle(options: BundleOptions): Promise<InstallResult>;
