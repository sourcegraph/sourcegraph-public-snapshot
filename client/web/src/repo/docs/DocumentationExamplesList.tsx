import * as H from 'history'
import React, { useMemo } from 'react'

import { fetchDocumentationReferences } from './graphql'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { catchError, startWith } from 'rxjs/operators'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { Observable } from 'rxjs'
import { RepositoryFields } from '../../graphql-operations'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { DocumentationExamplesListItem } from './DocumentationExamplesListItem'
import InformationIcon from 'mdi-react/InformationIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'

interface Props extends SettingsCascadeProps, VersionContextProps {
    location: H.Location
    isLightTheme: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    repo: RepositoryFields
    commitID: string
    pathID: string
}

const LOADING = 'loading' as const

export const DocumentationExamplesList: React.FunctionComponent<Props> = ({
    fetchHighlightedFileLineRanges,
    commitID,
    pathID,
    repo,
    ...props
}) => {
    const referencesLocations =
        useObservable(
            useMemo(
                () =>
                    fetchDocumentationReferences({
                        repo: repo.id,
                        revspec: commitID,
                        pathID: pathID,
                        first: 3,
                    }).pipe(
                        catchError(error => [asError(error)]),
                        startWith(LOADING)
                    ),
                [repo.id, commitID, pathID]
            )
        ) || LOADING

    return (
        <div className="documentation-examples">
            {referencesLocations === LOADING ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                referencesLocations.nodes.map((location, i) => (
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
            {referencesLocations !== LOADING && referencesLocations.nodes.length === 0 && (
                <span className="ml-2">
                    <InformationIcon className="icon-inline" /> None found
                </span>
            )}
        </div>
    )
}
