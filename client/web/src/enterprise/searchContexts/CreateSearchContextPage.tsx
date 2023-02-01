import React, { useCallback } from 'react'

import { mdiMagnify } from '@mdi/js'
import { Redirect } from 'react-router'
import { useLocation } from 'react-router-dom-v5-compat'
import { Observable } from 'rxjs'

import {
    Scalars,
    SearchContextInput,
    SearchContextRepositoryRevisionsInput,
    SearchContextFields,
} from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SearchContextProps } from '@sourcegraph/shared/src/search'
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
    extends ThemeProps,
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

    const location = useLocation()

    const query = parseSearchURLQuery(location.search)

    const onSubmit = useCallback(
        (
            id: Scalars['ID'] | undefined,
            searchContext: SearchContextInput,
            repositories: SearchContextRepositoryRevisionsInput[]
        ): Observable<SearchContextFields> => createSearchContext({ searchContext, repositories }, platformContext),
        [createSearchContext, platformContext]
    )

    if (!authenticatedUser) {
        return <Redirect to="/sign-in" />
    }

    return (
        <div className="w-100">
            <Page>
                <div className="container col-sm-8">
                    <PageTitle title="Create context" />
                    <PageHeader
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
                    >
                        <PageHeader.Heading as="h2" styleAs="h1">
                            <PageHeader.Breadcrumb icon={mdiMagnify} to="/search" aria-label="Code Search" />
                            <PageHeader.Breadcrumb to="/contexts">Contexts</PageHeader.Breadcrumb>
                            <PageHeader.Breadcrumb>Create context</PageHeader.Breadcrumb>
                        </PageHeader.Heading>
                    </PageHeader>
                    <SearchContextForm {...props} query={query} onSubmit={onSubmit} />
                </div>
            </Page>
        </div>
    )
}

export const CreateSearchContextPage = withAuthenticatedUser(AuthenticatedCreateSearchContextPage)
