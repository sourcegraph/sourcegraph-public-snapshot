import { Dependency, DependencyMeta } from '../interfaces';
export declare function parseDependency(raw: string): Dependency;
export declare function resolveDependency(raw: string, filename: string): string;
export interface ParseDependencyOptions {
    name?: string;
    source?: string;
}
export declare function parseDependencyExpression(raw: string, options?: ParseDependencyOptions): {
    name: string;
    location: string;
};
export declare function buildDependencyExpression(type: string, meta: DependencyMeta): string;
export declare function expandRegistry(raw: string, options?: ParseDependencyOptions): string;
