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

import { AuthenticatedUser } from '../auth'
import { Page } from '../components/Page'
import { PageTitle } from '../components/PageTitle'
import { SearchContextProps } from '../search'

import { SearchContextForm } from './SearchContextForm'

export interface CreateSearchContextPageProps
    extends RouteComponentProps,
        ThemeProps,
        TelemetryProps,
        Pick<SearchContextProps, 'createSearchContext'> {
    authenticatedUser: AuthenticatedUser | null
}

export const CreateSearchContextPage: React.FunctionComponent<CreateSearchContextPageProps> = props => {
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
                    <PageTitle title="Create a new search context" />
                    <h1 className="mb-1">Create a new search context</h1>
                    <div className="text-muted mb-4">
                        A search context represents a group of repositories at specified branches or revisions that will
                        be targeted by search queries.{' '}
                        <a
                            href="https://docs.sourcegraph.com/code_search/explanations/features#search-contexts-experimental"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            Learn more
                        </a>
                    </div>
                    <SearchContextForm {...props} onSubmit={onSubmit} authenticatedUser={authenticatedUser} />
                </div>
            </Page>
        </div>
    )
}
