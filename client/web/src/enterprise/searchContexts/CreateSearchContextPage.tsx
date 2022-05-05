import React, { useCallback } from 'react'

import MagnifyIcon from 'mdi-react/MagnifyIcon'
import { Redirect, RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'

import { SearchContextProps } from '@sourcegraph/search'
import {
    Scalars,
    SearchContextInput,
    SearchContextRepositoryRevisionsInput,
} from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { ISearchContext } from '@sourcegraph/shared/src/schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { PageHeader, Link } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { parseSearchURLQuery } from '../../search'

import { SearchContextForm } from './SearchContextForm'

export interface CreateSearchContextPageProps
    extends RouteComponentProps,
        ThemeProps,
        TelemetryProps,
        Pick<SearchContextProps, 'createSearchContext' | 'deleteSearchContext'>,
        PlatformContextProps<'requestGraphQL'> {
    authenticatedUser: AuthenticatedUser
    isSourcegraphDotCom: boolean
}

export const AuthenticatedCreateSearchContextPage: React.FunctionComponent<
    React.PropsWithChildren<CreateSearchContextPageProps>
> = props => {
    const { authenticatedUser, createSearchContext, platformContext } = props

    const query = parseSearchURLQuery(props.location.search)

    const onSubmit = useCallback(
        (
            id: Scalars['ID'] | undefined,
            searchContext: SearchContextInput,
            repositories: SearchContextRepositoryRevisionsInput[]
        ): Observable<ISearchContext> => createSearchContext({ searchContext, repositories }, platformContext),
        [createSearchContext, platformContext]
    )

    if (!authenticatedUser) {
        return <Redirect to="/sign-in" />
    }

    return (
        <div className="w-100">
            <Page>
                <div className="container col-8">
                    <PageTitle title="Create context" />
                    <PageHeader
                        path={[
                            {
                                icon: MagnifyIcon,
                                to: '/search',
                                ariaLabel: 'Code Search',
                            },
                            {
                                to: '/contexts',
                                text: 'Contexts',
                            },
                            { text: 'Create context' },
                        ]}
                        description={
                            <span className="text-muted">
                                A search context represents a group of repositories at specified branches or revisions
                                that will be targeted by search queries.{' '}
                                <Link
                                    to="/help/code_search/explanations/features#search-contexts"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    Learn more
                                </Link>
                            </span>
                        }
                        className="mb-3"
                    />
                    <SearchContextForm {...props} query={query} onSubmit={onSubmit} />
                </div>
            </Page>
        </div>
    )
}

export const CreateSearchContextPage = withAuthenticatedUser(AuthenticatedCreateSearchContextPage)
