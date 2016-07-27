import { Browser } from './config';
export interface Dependency {
    type: string;
    raw: string;
    location: string;
    meta: DependencyMeta;
}
export interface DependencyMeta {
    name?: string;
    path?: string;
    org?: string;
    repo?: string;
    sha?: string;
    version?: string;
    tag?: string;
    source?: string;
}
export interface DependencyTree {
    name?: string;
    version?: string;
    main?: string;
    browser?: Browser;
    typings?: string;
    browserTypings?: Browser;
    parent?: DependencyTree;
    files?: string[];
    postmessage?: string;
    src: string;
    raw: string;
    global: boolean;
    dependencies: DependencyBranch;
    devDependencies: DependencyBranch;
    peerDependencies: DependencyBranch;
    globalDependencies: DependencyBranch;
    globalDevDependencies: DependencyBranch;
}
export interface DependencyBranch {
    [name: string]: DependencyTree;
}
