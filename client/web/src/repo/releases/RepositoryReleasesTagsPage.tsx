import React, { useCallback, useEffect } from 'react'

import * as H from 'history'
import { Observable, of } from 'rxjs'

import { LoadingSpinner } from '@sourcegraph/wildcard'

import { FilteredConnection, FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { GitRefType, Scalars, GitRefConnectionFields, GitRefFields, RepositoryFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { GitReferenceNode, queryGitReferences as queryGitReferencesFromBackend } from '../GitReference'

interface Props {
    repo: RepositoryFields | undefined
    history: H.History
    location: H.Location
    queryGitReferences?: (args: {
        repo: Scalars['ID']
        first?: number
        query?: string
        type: GitRefType
        withBehindAhead?: boolean
    }) => Observable<GitRefConnectionFields>
}

/** A page that shows all of a repository's tags. */
export const RepositoryReleasesTagsPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    repo,
    history,
    location,
    queryGitReferences: queryGitReferences = queryGitReferencesFromBackend,
}) => {
    useEffect(() => {
        eventLogger.logViewEvent('RepositoryReleasesTags')
    }, [])

    const queryTags = useCallback(
        (args: FilteredConnectionQueryArguments): Observable<GitRefConnectionFields> => {
            if (!repo?.id) {
                return of()
            }

            return queryGitReferences({ ...args, repo: repo.id, type: GitRefType.GIT_TAG })
        },
        [repo?.id, queryGitReferences]
    )

    if (!repo) {
        return <LoadingSpinner />
    }

    return (
        <div className="repository-releases-page">
            <PageTitle title="Tags" />
            <FilteredConnection<GitRefFields>
                className="my-3"
                listClassName="list-group list-group-flush test-filtered-tags-connection"
                noun="tag"
                pluralNoun="tags"
                queryConnection={queryTags}
                nodeComponent={GitReferenceNode}
                ariaLabelFunction={(tagDisplayName: string) =>
                    `View this repository using ${tagDisplayName} as the selected revision`
                }
                defaultFirst={20}
                autoFocus={true}
                history={history}
                location={location}
            />
        </div>
    )
}
