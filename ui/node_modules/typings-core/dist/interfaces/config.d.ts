export interface ConfigJson {
    main?: string;
    browser?: Browser;
    version?: string;
    homepage?: string;
    resolution?: string | ResolutionMap;
    files?: string[];
    global?: boolean;
    postmessage?: string;
    name?: string;
    dependencies?: Dependencies;
    devDependencies?: Dependencies;
    peerDependencies?: Dependencies;
    globalDependencies?: Dependencies;
    globalDevDependencies?: Dependencies;
}
export declare type DependencyString = string;
export declare type Browser = string | Overrides;
export interface Overrides {
    [dependency: string]: string;
}
export interface Dependencies {
    [name: string]: DependencyString;
}
export interface ResolutionMap {
    main?: string;
    browser?: string;
    [resolution: string]: string;
}
