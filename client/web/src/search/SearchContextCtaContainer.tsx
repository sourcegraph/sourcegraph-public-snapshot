import { SearchContextCtaPromptProps, SearchContextCtaPrompt } from '@sourcegraph/search-ui'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'

import { useSearchContextCta, useShowSearchContextCta, useAuthenticatedUser } from '../stores'

export interface SearchContextCtaContainerProps
    extends Omit<SearchContextCtaPromptProps, 'onDismiss' | 'hasUserAddedExternalServices' | 'authenticatedUser'> {}

export const SearchContextCtaContainer: React.FunctionComponent<SearchContextCtaContainerProps> = props => {
    const authenticatedUser = useAuthenticatedUser()
    const showSearchContextCta = useShowSearchContextCta()
    const hasUserAddedExternalServices = useSearchContextCta(state => state.hasUserAddedExternalServices)
    const [contextCtaDismissed, setContextCtaDismissed] = useTemporarySetting('search.contexts.ctaDismissed', false)

    if (showSearchContextCta && !contextCtaDismissed) {
        return (
            <SearchContextCtaPrompt
                telemetryService={props.telemetryService}
                authenticatedUser={authenticatedUser}
                hasUserAddedExternalServices={hasUserAddedExternalServices}
                onDismiss={() => setContextCtaDismissed(true)}
                isExternalServicesUserModeAll={props.isExternalServicesUserModeAll}
            />
        )
    }

    return null
}
