import React from 'react'

import MagnifyIcon from 'mdi-react/MagnifyIcon'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { ButtonLink, Icon } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'

interface CreateSearchContextButtonProps {
    /** Search query string. */
    query?: string

    /** The currently authenticated user or null. */
    authenticatedUser: Pick<AuthenticatedUser, 'id'> | null
}

export const CreateSearchContextButton: React.FunctionComponent<
    React.PropsWithChildren<CreateSearchContextButtonProps>
> = props => {
    if (!props.query || !props.authenticatedUser) {
        return null
    }

    const contextFilter = findFilter(props.query, FilterType.context, FilterKind.Global)
    if (!contextFilter || contextFilter.value?.value !== 'global') {
        return null
    }

    const query = omitFilter(props.query, contextFilter)
    const searchParameters = new URLSearchParams()
    searchParameters.set('q', query)
    const toURL = `/contexts/new?${searchParameters.toString()}`

    return (
        <li data-tooltip="Create search context based on this query" data-delay={10000} className="nav-item mr-2">
            <ButtonLink to={toURL} className="text-decoration-none" variant="secondary" outline={true} size="sm">
                <Icon className="mr-1" as={MagnifyIcon} />
                Create context
            </ButtonLink>
        </li>
    )
}
