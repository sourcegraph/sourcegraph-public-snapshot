import { ResolutionMap } from '../interfaces';
export declare const EOL: string;
export declare function isHttp(url: string): boolean;
export declare function isDefinition(path: string): boolean;
export declare function isModuleName(value: string): boolean;
export declare function normalizeSlashes(path: string): string;
export declare function joinUrl(from: string, to: string): string;
export declare function resolveFrom(from: string, to: string): string;
export declare function relativeTo(from: string, to: string): string;
export declare function toDefinition(path: string): string;
export declare function pathFromDefinition(path: string): string;
export declare function normalizeToDefinition(path: string): string;
export declare function getDefinitionPath(path: string): string;
export interface LocationOptions {
    name: string;
    path: string;
    global: boolean;
}
export interface DependencyLocationResult {
    definition: string;
    directory: string;
    config: string;
}
export declare function getDependencyPath(options: LocationOptions): DependencyLocationResult;
export declare function getInfoFromDependencyLocation(location: string, bundle: string): {
    location: string;
    global: boolean;
    name: string;
};
export declare function detectEOL(contents: string): string;
export declare function normalizeEOL(contents: string, eol: string): string;
export declare function normalizeResolutions(resolutions: string | ResolutionMap, options: {
    cwd: string;
}): ResolutionMap;
