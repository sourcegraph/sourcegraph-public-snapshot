import * as H from 'history'
import React from 'react'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { RepoIcon } from '@sourcegraph/shared/src/components/RepoIcon'

import { map } from 'rxjs/operators'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { CodeExcerpt, FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { Observable } from 'rxjs'
import { RepositoryFields } from '../../graphql-operations'
import { RepoFileLink } from '@sourcegraph/shared/src/components/RepoFileLink'

interface Props extends SettingsCascadeProps, VersionContextProps {
    location: H.Location
    isLightTheme: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    repo: RepositoryFields
    commitID: string
    pathID: string
    item: GQL.ILocation
}

const contextLines = 1

export const DocumentationExamplesListItem: React.FunctionComponent<Props> = ({
    fetchHighlightedFileLineRanges,
    repo,
    commitID,
    pathID,
    item,
    ...props
}) => {
    const fetchHighlightedFileRangeLines = React.useCallback(
        (isFirst, startLine, endLine, isLightTheme) => {
            return fetchHighlightedFileLineRanges(
                {
                    repoName: item.resource.repository.name,
                    commitID: item.resource.commit.oid,
                    filePath: item.resource.path,
                    disableTimeout: false,
                    isLightTheme,
                    ranges: [
                        {
                            startLine: (item.range?.start.line || 0) - contextLines,
                            endLine: (item.range?.end.line || 0) + contextLines + 1,
                        },
                    ],
                },
                false
            ).pipe(
                map(lines => {
                    // Hack to remove newlines which cause duplicate newlines when copying/pasting code snippets.
                    return lines[0].map(line => line.replace(/(?:\r\n|\r|\n)/g, ''))
                })
            )
        },
        [repo, commitID, item, fetchHighlightedFileLineRanges]
    )

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
