import React from 'react'

import { ButtonLink } from '@sourcegraph/shared/src/components/LinkOrButton'

import { PatternTypeProps } from '../..'
import { AuthenticatedUser } from '../../../auth'
import { CodeInsightsIcon } from '../../../insights/components'

interface CreateCodeInsightButtonProps extends Pick<PatternTypeProps, 'patternType'> {
    /** Search query string. */
    query?: string

    /** The currently authenticated user or null. */
    authenticatedUser: Pick<AuthenticatedUser, 'id'> | null

    /** Whether the code insights feature flag is enabled. */
    enableCodeInsights?: boolean
}

/**
 * Displays code insights creation button from search query.
 *
 * Basically it navigates user to search based insight creation UI with
 * prefilled repositories and data series query field according to
 * search page query.
 */
export const CreateCodeInsightButton: React.FunctionComponent<CreateCodeInsightButtonProps> = props => {
    if (!props.enableCodeInsights || !props.query || !props.authenticatedUser) {
        return null
    }

    const searchParameters = new URLSearchParams()
    searchParameters.set('query', `${props.query} patterntype:${props.patternType}`)
    const toURL = `/insights/create/search?${searchParameters.toString()}`

    return (
        <li data-tooltip="Create Insight based on this search query" data-delay={10000} className="nav-item">
            <ButtonLink to={toURL} className="btn btn-sm btn-outline-secondary mr-2 nav-link text-decoration-none">
                <CodeInsightsIcon className="icon-inline mr-1" />
                Create Insight
            </ButtonLink>
        </li>
    )
}
