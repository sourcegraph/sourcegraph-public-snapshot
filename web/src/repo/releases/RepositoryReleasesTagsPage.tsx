import React, { useEffect, useCallback } from 'react'
import * as H from 'history'
import * as GQL from '../../../../shared/src/graphql/schema'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { GitRefNode, queryGitRefs as _queryGitRefs } from '../GitRef'
import { RepositoryReleasesAreaPageProps } from './RepositoryReleasesArea'
import { Observable } from 'rxjs'

interface Props extends RepositoryReleasesAreaPageProps {
    history: H.History
    location: H.Location
    queryGitRefs?: (args: {
        repo: GQL.ID
        first?: number
        query?: string
        type: GQL.GitRefType
        withBehindAhead?: boolean
    }) => Observable<GQL.IGitRefConnection>
}

/** A page that shows all of a repository's tags. */
export const RepositoryReleasesTagsPage: React.FunctionComponent<Props> = ({
    repo,
    history,
    location,
    queryGitRefs = _queryGitRefs,
}) => {
    useEffect(() => {
        eventLogger.logViewEvent('RepositoryReleasesTags')
    }, [])

    const queryTags = useCallback(
        (args: FilteredConnectionQueryArgs): Observable<GQL.IGitRefConnection> =>
            queryGitRefs({ ...args, repo: repo.id, type: GQL.GitRefType.GIT_TAG }),
        [repo.id, queryGitRefs]
    )

    return (
        <div className="repository-releases-page">
            <PageTitle title="Tags" />
            <FilteredConnection<GQL.IGitRef>
                className="my-3"
                listClassName="list-group list-group-flush"
                noun="tag"
                pluralNoun="tags"
                queryConnection={queryTags}
                nodeComponent={GitRefNode}
                defaultFirst={20}
                autoFocus={true}
                history={history}
                location={location}
            />
        </div>
    )
}
