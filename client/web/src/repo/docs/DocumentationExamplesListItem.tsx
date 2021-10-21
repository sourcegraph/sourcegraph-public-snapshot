import * as H from 'history'
import React, { useMemo } from 'react'
import { Observable } from 'rxjs'
import { catchError, map, startWith } from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/codeintellify/lib/errors'
import { CodeExcerpt, FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { RepoFileLink } from '@sourcegraph/shared/src/components/RepoFileLink'
import { RepoIcon } from '@sourcegraph/shared/src/components/RepoIcon'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { asError } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { Timestamp } from '../../components/time/Timestamp'
import { RepositoryFields } from '../../graphql-operations'
import { PersonLink } from '../../person/PersonLink'

import { fetchDocumentationBlame } from './graphql'

interface Props extends SettingsCascadeProps {
    location: H.Location
    isLightTheme: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    repo: RepositoryFields
    commitID: string
    pathID: string
    item: GQL.ILocation
}

const contextLines = 1

const LOADING = 'loading' as const

export const DocumentationExamplesListItem: React.FunctionComponent<Props> = ({
    fetchHighlightedFileLineRanges,
    repo,
    commitID,
    pathID,
    item,
    ...props
}) => {
    const fetchHighlightedFileRangeLines = React.useCallback(
        (isFirst, startLine, endLine) =>
            fetchHighlightedFileLineRanges(
                {
                    repoName: item.resource.repository.name,
                    commitID: item.resource.commit.oid,
                    filePath: item.resource.path,
                    disableTimeout: false,
                    ranges: [
                        {
                            startLine: (item.range?.start.line || 0) - contextLines,
                            endLine: (item.range?.end.line || 0) + contextLines + 1,
                        },
                    ],
                },
                false
            ).pipe(
                map(lines =>
                    // Hack to remove newlines which cause duplicate newlines when copying/pasting code snippets.
                    lines[0].map(line => line.replace(/\r\n|\r|\n/g, ''))
                )
            ),
        [item, fetchHighlightedFileLineRanges]
    )

    const blameHunks =
        useObservable(
            useMemo(
                () =>
                    fetchDocumentationBlame({
                        repo: item.resource.repository.name,
                        revspec: item.resource.commit.oid,
                        path: item.resource.path,
                        startLine: item.range?.start.line || 0,
                        endLine: item.range?.end.line || 0,
                    }).pipe(
                        catchError(error => [asError(error)]),
                        startWith(LOADING)
                    ),
                [item]
            )
        ) || LOADING

    return (
        <div className="documentation-examples-list-item mt-2">
            <div className="p-2">
                <RepoIcon
                    repoName={item.resource.repository.name}
                    className="icon-inline text-muted flex-shrink-0 mr-2"
                />
                <RepoFileLink
                    repoName={item.resource.repository.name}
                    repoURL={item.resource.repository.url}
                    filePath={item.resource.path}
                    // Hack because the backend incorrectly returns /-/tree, and linking to that does
                    // redirect to /-/blob, but doesn't redirect to the right line range on the page.
                    fileURL={item.url.replace('/-/tree/', '/-/blob/')}
                    className="documentation-examples-list-item__repo-file-link"
                />
                {blameHunks !== LOADING && !isErrorLike(blameHunks) && blameHunks.length > 0 && (
                    <span className="float-right text-muted">
                        by <PersonLink person={blameHunks[0].author.person} />{' '}
                        <Link to={blameHunks[0].commit.url}>
                            <Timestamp date={blameHunks[0].author.date} />
                        </Link>
                    </span>
                )}
            </div>
            <CodeExcerpt
                key={item.url}
                repoName={item.resource.repository.name}
                commitID={item.resource.commit.oid}
                filePath={item.resource.path}
                startLine={(item.range?.start.line || 0) - contextLines}
                endLine={(item.range?.end.line || 0) + contextLines + 1}
                highlightRanges={[
                    {
                        line: item.range?.start.line || 0,
                        character: item.range?.start.character || 0,
                        highlightLength: (item.range?.end.character || 0) - (item.range?.start.character || 0),
                    },
                ]}
                className="documentation-examples-list-item__code-excerpt"
                fetchHighlightedFileRangeLines={fetchHighlightedFileRangeLines}
                isFirst={false}
                {...props}
            />
        </div>
    )
}
