import type { FC } from 'react'

import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'

import { NewCodySidebar, type Repository } from './new-cody-sidebar/NewCodySidebar'
import { CodySidebar as OldCodySidebar } from './old-cody-sidebar/CodySidebar'

interface CodySidebarProps extends TelemetryV2Props {
    filePath: string | undefined
    repository: Repository
    authenticatedUser: AuthenticatedUser | null
    onClose: () => void
}

/**
 * Cody sidebar component, it's used on repository and blob UI page.
 * Contains different version of the cody web chat.
 */
export const CodySidebar: FC<CodySidebarProps> = props => {
    const { filePath, repository, authenticatedUser, telemetryRecorder, onClose } = props

    // We have two different version of Cody Web, first was created as original
    // Cody Web chat, second version (NewCodyChatPage) is a port from VSCode
    // cody extension.
    const newCodyWeb = useExperimentalFeatures(features => features.newCodyWeb)

    return newCodyWeb ? (
        <NewCodySidebar
            filePath={filePath}
            repository={repository}
            isAuthorized={authenticatedUser !== null}
            onClose={onClose}
        />
    ) : (
        <OldCodySidebar telemetryRecorder={telemetryRecorder} authenticatedUser={authenticatedUser} onClose={onClose} />
    )
}
