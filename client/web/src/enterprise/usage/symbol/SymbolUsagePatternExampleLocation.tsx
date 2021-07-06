import H from 'history'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React from 'react'

import { Location } from '@sourcegraph/extension-api-types'
import { FileMatch } from '@sourcegraph/shared/src/components/FileMatch'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'

import { SymbolUsagePatternExampleLocationFields } from '../../../graphql-operations'
import { fetchHighlightedFileLineRanges } from '../../../repo/backend'

export const SymbolUsagePatternExampleLocationGQLFragment = gql`
    fragment SymbolUsagePatternExampleLocationFields on SymbolUsagePatternExampleLocationEdge {
        description
        location {
            range {
                start {
                    line
                    character
                }
                end {
                    line
                    character
                }
            }
            resource {
                path
                commit {
                    oid
                }
                repository {
                    name
                }
            }
        }
    }
`

interface Props extends SettingsCascadeProps, ThemeProps, VersionContextProps {
    exampleLocation: SymbolUsagePatternExampleLocationFields

    location: H.Location
}

export const SymbolUsagePatternExampleLocation: React.FunctionComponent<Props> = ({ exampleLocation, ...props }) => {
    const a = 1
    return (
        <div className="pb-2">
            <Markdown
                dangerousInnerHTML={renderMarkdown(exampleLocation.description)}
                className="text-muted py-1 px-2"
            />
            <FileMatch
                result={{
                    type: 'file',
                    repository: exampleLocation.location.resource.repository.name,
                    version: exampleLocation.location.resource.commit.oid,
                    name: exampleLocation.location.resource.path,
                    lineMatches: [
                        {
                            line: '',
                            lineNumber: exampleLocation.location.range!.start.line,
                            offsetAndLengths: [
                                [
                                    exampleLocation.location.range!.start.character,
                                    exampleLocation.location.range!.end.character -
                                        exampleLocation.location.range!.start.character,
                                ],
                            ],
                        },
                    ],
                }}
                fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                icon={SourceRepositoryIcon}
                expanded={true}
                onSelect={() => {}}
                showAllMatches={true}
                {...props}
            />
        </div>
    )
}
