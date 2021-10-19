import MagnifyIcon from 'mdi-react/MagnifyIcon'
import React, { useCallback } from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'

import {
    Scalars,
    SearchContextInput,
    SearchContextRepositoryRevisionsInput,
} from '@sourcegraph/shared/src/graphql-operations'
import { ISearchContext } from '@sourcegraph/shared/src/graphql/schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Page } from '@sourcegraph/web/src/components/Page'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { SearchContextProps } from '../../search'

import { SearchContextForm } from './SearchContextForm'

export interface CreateSearchContextPageProps
    extends RouteComponentProps,
        ThemeProps,
        TelemetryProps,
        Pick<SearchContextProps, 'createSearchContext' | 'deleteSearchContext'> {
    authenticatedUser: AuthenticatedUser
}

export const AuthenticatedCreateSearchContextPage: React.FunctionComponent<CreateSearchContextPageProps> = props => {
    const { authenticatedUser, createSearchContext } = props
    const onSubmit = useCallback(
        (
            id: Scalars['ID'] | undefined,
            searchContext: SearchContextInput,
            repositories: SearchContextRepositoryRevisionsInput[]
        ): Observable<ISearchContext> => createSearchContext({ searchContext, repositories }),
        [createSearchContext]
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
                                <a
                                    href="https://docs.sourcegraph.com/code_search/explanations/features#search-contexts"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    Learn more
                                </a>
                            </span>
                        }
                        className="mb-3"
                    />
                    <SearchContextForm {...props} onSubmit={onSubmit} />
                </div>
            </Page>
        </div>
    )
}

export const CreateSearchContextPage = withAuthenticatedUser(AuthenticatedCreateSearchContextPage)
