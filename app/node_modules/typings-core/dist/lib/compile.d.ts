import Promise = require('any-promise');
import { DependencyTree, Emitter } from '../interfaces';
export interface Options {
    cwd: string;
    name: string;
    global: boolean;
    meta: boolean;
    emitter: Emitter;
}
export interface ResolutionResult {
    main?: string;
    browser?: string;
    [name: string]: string;
}
export interface CompileResult {
    cwd: string;
    name: string;
    tree: DependencyTree;
    results: ResolutionResult;
    global: boolean;
}
export declare function compile(tree: DependencyTree, resolutions: string[], options: Options): Promise<CompileResult>;
