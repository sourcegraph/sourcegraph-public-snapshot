import Promise = require('any-promise');
import { Dependency, DependencyTree, Emitter } from '../interfaces';
export declare const DEFAULT_DEPENDENCY: DependencyTree;
export interface Options {
    cwd: string;
    emitter: Emitter;
    name?: string;
    dev?: boolean;
    peer?: boolean;
    global?: boolean;
    parent?: DependencyTree;
}
export declare function resolveAllDependencies(options: Options): Promise<DependencyTree>;
export declare function resolveDependency(dependency: Dependency, options: Options): Promise<DependencyTree>;
export declare function resolveBowerDependencies(options: Options): Promise<DependencyTree>;
export declare function resolveNpmDependencies(options: Options): Promise<DependencyTree>;
export declare function resolveTypeDependencies(options: Options): Promise<DependencyTree>;
