import { mdiCloudQuestion } from '@mdi/js'

import { Icon } from '@sourcegraph/wildcard'

import { externalRepoIcon } from '../../components/externalServices/externalServices'
import { ExternalRepositoryFields } from '../../graphql-operations'

export const ExternalRepositoryIcon: React.FunctionComponent<
    React.PropsWithChildren<{
        externalRepo: ExternalRepositoryFields
    }>
> = ({ externalRepo }) => {
    const IconComponent = externalRepoIcon(externalRepo)
    return IconComponent ? (
        <Icon as={IconComponent} aria-label="Code host logo" className="mr-2" />
    ) : (
        <Icon svgPath={mdiCloudQuestion} aria-label="Unknown code host" className="mr-2" />
    )
}
