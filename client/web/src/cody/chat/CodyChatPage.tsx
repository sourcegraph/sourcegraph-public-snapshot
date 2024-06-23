import type { FC } from 'react'

import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import type { SourcegraphContext } from '../../jscontext'

import { CodyChatPage as OldCodyChatPage } from './old-chat/CodyChatPage'

// Lazy loaded new cody chat page, we have to lazy load it
// since new cody web client pulls heavy agent
const LazyNewCodyChatPage = lazyComponent(() => import('./new-chat/NewCodyChatPage'), 'NewCodyChatPage')

interface CodyChatPageProps extends TelemetryV2Props {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser
    context: Pick<SourcegraphContext, 'externalURL'>
}

const AuthenticatedCodyChatPage: FC<CodyChatPageProps> = ({
    isSourcegraphDotCom,
    authenticatedUser,
    context,
    telemetryRecorder,
}) => {
    // We have two different version of Cody Web, first was created as original
    // Cody Web chat, second version (NewCodyChatPage) is a port from VSCode
    // cody extension.
    const newCodyWeb = !useExperimentalFeatures(features => features.newCodyWeb)

    return newCodyWeb ? (
        <LazyNewCodyChatPage isSourcegraphDotCom={isSourcegraphDotCom} />
    ) : (
        <OldCodyChatPage
            isSourcegraphDotCom={isSourcegraphDotCom}
            authenticatedUser={authenticatedUser}
            context={context}
            telemetryRecorder={telemetryRecorder}
        />
    )
}

export const CodyChatPage = withAuthenticatedUser(AuthenticatedCodyChatPage)
