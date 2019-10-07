import { TextDocument, WorkspaceEdit } from 'sourcegraph'

export interface PackageJsonPackage {
    packageJson: TextDocument
    lockfile: TextDocument
}

export interface PackageJsonDependency {
    name: string
    version: string
}

export interface PackageJsonPackageManager {
    packagesWithUnsatisfiedDependencyVersionRange(dep: PackageJsonDependency): Promise<PackageJsonPackage[]>
    editForDependencyUpgrade(pkg: PackageJsonPackage, dep: PackageJsonDependency): Promise<WorkspaceEdit>
}
