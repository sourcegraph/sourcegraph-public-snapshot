import type { FC } from 'react'

import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'

import type { SourcegraphContext } from '../../jscontext'

import { NewCodyChatPage } from './new-chat/NewCodyChatPage'
import { CodyChatPage as OldCodyChatPage } from './old-chat/CodyChatPage'

interface CodyChatPageProps extends TelemetryV2Props {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser | null
    context: Pick<SourcegraphContext, 'externalURL'>
}

export const CodyChatPage: FC<CodyChatPageProps> = props => {
    const { isSourcegraphDotCom, authenticatedUser, context, telemetryRecorder } = props

    // We have two different version of Cody Web, first was created as original
    // Cody Web chat, second version (NewCodyChatPage) is a port from VSCode
    // cody extension.
    const newCodyWeb = useExperimentalFeatures(features => features.newCodyWeb)

    return newCodyWeb ? (
        <NewCodyChatPage />
    ) : (
        <OldCodyChatPage
            isSourcegraphDotCom={isSourcegraphDotCom}
            authenticatedUser={authenticatedUser}
            context={context}
            telemetryRecorder={telemetryRecorder}
        />
    )
}
