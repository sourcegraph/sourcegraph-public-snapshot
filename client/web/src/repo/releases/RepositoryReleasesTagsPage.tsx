import React, { useCallback, useEffect } from 'react'

import * as H from 'history'
import { Observable } from 'rxjs'

import { FilteredConnection, FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { GitRefType, Scalars, GitRefConnectionFields, GitRefFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { GitReferenceNode, queryGitReferences as queryGitReferencesFromBackend } from '../GitReference'

import { RepositoryReleasesAreaPageProps } from './RepositoryReleasesArea'

interface Props extends RepositoryReleasesAreaPageProps {
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
        (args: FilteredConnectionQueryArguments): Observable<GitRefConnectionFields> =>
            queryGitReferences({ ...args, repo: repo.id, type: GitRefType.GIT_TAG }),
        [repo.id, queryGitReferences]
    )

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
                defaultFirst={20}
                autoFocus={true}
                history={history}
                location={location}
            />
        </div>
    )
}
