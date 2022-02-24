import * as H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import InformationIcon from 'mdi-react/InformationIcon'
import React, { useMemo } from 'react'
import { Observable } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import * as GQL from '@sourcegraph/shared/src/schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { LoadingSpinner, useObservable, Icon } from '@sourcegraph/wildcard'

import { RepositoryFields } from '../../graphql-operations'

import { DocumentationExamplesListItem } from './DocumentationExamplesListItem'
import { fetchDocumentationReferences } from './graphql'

interface Props extends SettingsCascadeProps {
    location: H.Location
    isLightTheme: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    repo: RepositoryFields
    commitID: string
    pathID: string
    count: number
}

const LOADING = 'loading' as const

export const DocumentationExamplesList: React.FunctionComponent<Props> = ({
    fetchHighlightedFileLineRanges,
    commitID,
    pathID,
    repo,
    count,
    ...props
}) => {
    const referencesLocations =
        useObservable(
            useMemo(
                () =>
                    fetchDocumentationReferences({
                        repo: repo.id,
                        revspec: commitID,
                        pathID,
                        first: count,
                    }).pipe(
                        catchError(error => [asError(error)]),
                        startWith(LOADING)
                    ),
                [repo.id, commitID, pathID, count]
            )
        ) || LOADING

    return (
        <div className="documentation-examples">
            {referencesLocations === LOADING ? (
                <LoadingSpinner />
            ) : (
                (referencesLocations as GQL.ILocationConnection).nodes.map(location => (
                    <DocumentationExamplesListItem
                        key={location.url}
                        item={location}
                        fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                        repo={repo}
                        commitID={commitID}
                        pathID={pathID}
                        {...props}
                    />
                ))
            )}
            {referencesLocations !== LOADING && isErrorLike(referencesLocations) && (
                <span className="ml-2">
                    <Icon as={AlertCircleIcon} /> Error: {referencesLocations}
                </span>
            )}
            {referencesLocations !== LOADING &&
                !isErrorLike(referencesLocations) &&
                referencesLocations.nodes.length === 0 && (
                    <span className="ml-2">
                        <Icon as={InformationIcon} /> None found
                    </span>
                )}
        </div>
    )
}
