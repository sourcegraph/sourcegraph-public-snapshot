import React, { useCallback, useEffect } from 'react'

import { type Observable, of } from 'rxjs'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { FilteredConnection, type FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import {
    GitRefType,
    type Scalars,
    type GitRefConnectionFields,
    type GitRefFields,
    type RepositoryFields,
} from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import {
    GitReferenceNode,
    type GitReferenceNodeProps,
    queryGitReferences as queryGitReferencesFromBackend,
} from '../GitReference'

interface Props extends TelemetryV2Props {
    repo: RepositoryFields | undefined
    isPackage?: boolean
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
    isPackage,
    queryGitReferences: queryGitReferences = queryGitReferencesFromBackend,
    telemetryRecorder,
}) => {
    useEffect(() => {
        telemetryRecorder.recordEvent('repositoryReleasesTags', 'viewed')
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
            <PageTitle title={isPackage ? 'Versions' : 'Tags'} />
            <FilteredConnection<GitRefFields, Partial<GitReferenceNodeProps>>
                className="my-3"
                listClassName="list-group list-group-flush test-filtered-tags-connection"
                {...(isPackage ? { noun: 'version', pluralNoun: 'versions' } : { noun: 'tag', pluralNoun: 'tags' })}
                queryConnection={queryTags}
                nodeComponent={GitReferenceNode}
                nodeComponentProps={{ isPackageVersion: isPackage }}
                ariaLabelFunction={(tagDisplayName: string) =>
                    `View this repository using ${tagDisplayName} as the selected revision`
                }
                defaultFirst={20}
                autoFocus={true}
            />
        </div>
    )
}
