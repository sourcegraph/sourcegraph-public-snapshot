import { mdiCloudQuestion } from '@mdi/js'

import { Icon } from '@sourcegraph/wildcard'

import { defaultExternalServices } from '../../components/externalServices/externalServices'
import { ExternalServiceKind } from '../../graphql-operations'

export type PackageHost = 'npm' | 'go' | 'semanticdb' | 'scip-ruby' | 'python' | 'rust-analyzer'

const PACKAGE_HOST_TO_EXTERNAL_REPO: Record<PackageHost, ExternalServiceKind> = {
    npm: ExternalServiceKind.NPMPACKAGES,
    go: ExternalServiceKind.GOMODULES,
    semanticdb: ExternalServiceKind.JVMPACKAGES,
    'scip-ruby': ExternalServiceKind.RUBYPACKAGES,
    python: ExternalServiceKind.PYTHONPACKAGES,
    'rust-analyzer': ExternalServiceKind.RUSTPACKAGES,
}

interface PackageRepositoryIconProps {
    host: PackageHost
}

export const PackageRepositoryIcon: React.FunctionComponent<PackageRepositoryIconProps> = ({ host }) => {
    const IconComponent = defaultExternalServices[PACKAGE_HOST_TO_EXTERNAL_REPO[host]].icon
    return IconComponent ? (
        <Icon as={IconComponent} aria-label="Package host logo" className="mr-2" />
    ) : (
        <Icon svgPath={mdiCloudQuestion} aria-label="Unknown package host" className="mr-2" />
    )
}
