import * as H from 'history'
import ErrorIcon from 'mdi-react/ErrorIcon'
import InformationIcon from 'mdi-react/InformationIcon'
import React, { useMemo } from 'react'
import { Observable } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

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
                <LoadingSpinner className="icon-inline" />
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
                    <ErrorIcon className="icon-inline" /> Error: {referencesLocations}
                </span>
            )}
            {referencesLocations !== LOADING &&
                !isErrorLike(referencesLocations) &&
                referencesLocations.nodes.length === 0 && (
                    <span className="ml-2">
                        <InformationIcon className="icon-inline" /> None found
                    </span>
                )}
        </div>
    )
}
