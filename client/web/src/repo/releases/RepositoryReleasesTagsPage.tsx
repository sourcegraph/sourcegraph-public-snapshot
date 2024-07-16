import React, { useCallback, useEffect } from 'react'

import { of, type Observable } from 'rxjs'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { FilteredConnection, type FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import {
    GitRefType,
    type GitRefConnectionFields,
    type GitRefFields,
    type RepositoryFields,
    type Scalars,
} from '../../graphql-operations'
import {
    GitReferenceNode,
    queryGitReferences as queryGitReferencesFromBackend,
    type GitReferenceNodeProps,
} from '../GitReference'

interface Props extends TelemetryV2Props {
    repo: RepositoryFields | undefined
    isPackage?: boolean
    queryGitReferences?: (args: {
        repo: Scalars['ID']
        first?: number | null
        query?: string
        type: GitRefType
        withBehindAhead?: boolean
    }) => Observable<GitRefConnectionFields>
}

/** A page that shows all of a repository's tags. */
export const RepositoryReleasesTagsPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    repo,
    isPackage,
    queryGitReferences = queryGitReferencesFromBackend,
    telemetryRecorder,
}) => {
    useEffect(() => {
        EVENT_LOGGER.logViewEvent('RepositoryReleasesTags')
        telemetryRecorder.recordEvent('repo.releasesTags', 'view')
    }, [telemetryRecorder])

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
