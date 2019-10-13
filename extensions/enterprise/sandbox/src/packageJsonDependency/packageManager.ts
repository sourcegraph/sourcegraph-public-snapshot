import { TextDocument, WorkspaceEdit } from 'sourcegraph'
import { PackageJsonDependencyCampaignContext } from './packageJsonDependency'

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
    packagesWithDependencySatisfyingVersionRange(
        dep: PackageJsonDependency,
        queryFilters?: string
    ): Promise<ResolvedDependencyInPackage[]>
    editForDependencyAction(
        dep: ResolvedDependencyInPackage,
        action: PackageJsonDependencyCampaignContext['action']
    ): Promise<WorkspaceEdit>
}
