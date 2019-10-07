import { TextDocument } from 'sourcegraph'

export interface PackageJsonPackage {
    packageJson: TextDocument
    lockfile: TextDocument
}

export interface PackageJsonPackageManager {
    packagesWithUnsatisfiedDependencyVersionRange(name: string, versionRange: string): Promise<PackageJsonPackage[]>
}
