import { TextDocument, WorkspaceEdit } from 'sourcegraph'

export interface ResolvedDependencyInPackage {
    packageJson: TextDocument
    lockfile: TextDocument

    dependency: ResolvedDependency
}

export interface PackageJsonDependency {
    name: string
    version: string
}

export interface ResolvedDependency extends PackageJsonDependency {
    direct: boolean
}

export interface PackageJsonPackageManager {
    packagesWithUnsatisfiedDependencyVersionRange(
        dep: PackageJsonDependency,
        queryFilters?: string
    ): Promise<ResolvedDependencyInPackage[]>
    editForDependencyUpgrade(dep: ResolvedDependencyInPackage): Promise<WorkspaceEdit>
}
