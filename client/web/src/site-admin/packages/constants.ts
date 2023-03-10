import { ExternalServiceKind, PackageRepoReferenceKind } from '../../graphql-operations'

export const ExternalServicePackageMap: Partial<
    Record<
        ExternalServiceKind,
        {
            label: string
            value: PackageRepoReferenceKind
        }
    >
> = {
    [ExternalServiceKind.NPMPACKAGES]: {
        label: 'npm',
        value: PackageRepoReferenceKind.NPMPACKAGES,
    },
    [ExternalServiceKind.GOMODULES]: {
        label: 'Go',
        value: PackageRepoReferenceKind.GOMODULES,
    },
    [ExternalServiceKind.JVMPACKAGES]: {
        label: 'JVM',
        value: PackageRepoReferenceKind.JVMPACKAGES,
    },
    [ExternalServiceKind.RUBYPACKAGES]: {
        label: 'Ruby',
        value: PackageRepoReferenceKind.RUBYPACKAGES,
    },
    [ExternalServiceKind.PYTHONPACKAGES]: {
        label: 'Python',
        value: PackageRepoReferenceKind.PYTHONPACKAGES,
    },
    [ExternalServiceKind.RUSTPACKAGES]: {
        label: 'Rust',
        value: PackageRepoReferenceKind.RUSTPACKAGES,
    },
}

export const PackageExternalServiceMap: Partial<
    Record<
        PackageRepoReferenceKind,
        {
            label: string
            value: ExternalServiceKind
        }
    >
> = {
    [PackageRepoReferenceKind.NPMPACKAGES]: {
        label: 'npm',
        value: ExternalServiceKind.NPMPACKAGES,
    },
    [PackageRepoReferenceKind.GOMODULES]: {
        label: 'Go',
        value: ExternalServiceKind.GOMODULES,
    },
    [PackageRepoReferenceKind.JVMPACKAGES]: {
        label: 'JVM',
        value: ExternalServiceKind.JVMPACKAGES,
    },
    [PackageRepoReferenceKind.RUBYPACKAGES]: {
        label: 'Ruby',
        value: ExternalServiceKind.RUBYPACKAGES,
    },
    [PackageRepoReferenceKind.PYTHONPACKAGES]: {
        label: 'Python',
        value: ExternalServiceKind.PYTHONPACKAGES,
    },
    [PackageRepoReferenceKind.RUSTPACKAGES]: {
        label: 'Rust',
        value: ExternalServiceKind.RUSTPACKAGES,
    },
}
