import Promise = require('any-promise');
import { parseDependencyExpression, buildDependencyExpression } from './utils/parse';
import { DependencyTree, Emitter, ResolutionMap } from './interfaces';
export { parseDependencyExpression, buildDependencyExpression };
export interface InstallDependencyOptions {
    save?: boolean;
    saveDev?: boolean;
    savePeer?: boolean;
    global?: boolean;
    cwd: string;
    name?: string;
    source?: string;
    emitter?: Emitter;
}
export interface InstallOptions {
    cwd: string;
    production?: boolean;
    emitter?: Emitter;
}
export interface InstallResult {
    tree: DependencyTree;
    name?: string;
}
export interface InstallDependencyNestedOptions extends InstallDependencyOptions {
    resolutions: ResolutionMap;
}
export declare function install(options: InstallOptions): Promise<InstallResult>;
export interface InstallExpression {
    name: string;
    location: string;
}
export declare function installDependencyRaw(raw: string, options: InstallDependencyOptions): Promise<InstallResult>;
export declare function installDependenciesRaw(raw: string[], options: InstallDependencyOptions): Promise<InstallResult[]>;
export declare function installDependency(expression: InstallExpression, options: InstallDependencyOptions): Promise<InstallResult>;
export declare function installDependencies(expressions: InstallExpression[], options: InstallDependencyOptions): Promise<InstallResult[]>;
