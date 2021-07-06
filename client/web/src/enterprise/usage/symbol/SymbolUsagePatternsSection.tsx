import H from 'history'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React from 'react'
import { of } from 'rxjs'

import { FileLocations } from '@sourcegraph/branded/src/components/panel/views/FileLocations'
import { Location } from '@sourcegraph/extension-api-types'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'

import { SymbolUsagePatternFields } from '../../../graphql-operations'
import { fetchHighlightedFileLineRanges } from '../../../repo/backend'

const HACK_CSS = (
    <style>
        {
            'td.line { display: none; } .code-excerpt .code { padding-left: 0.25rem !important; } .result-container__header { display: none; } .result-container { border: solid 1px var(--border-color) !important; border-width: 1px !important; margin: 1rem; }'
        }
    </style>
)

export const SymbolUsagePatternGQLFragment = gql`
    fragment SymbolUsagePatternFields on SymbolUsagePattern {
        description

        exampleLocations {
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
    }
`

interface Props extends SettingsCascadeProps, ThemeProps, VersionContextProps {
    usagePatterns: SymbolUsagePatternFields[]

    location: H.Location
}

export const SymbolUsagePatternsSection: React.FunctionComponent<Props> = ({ usagePatterns, ...props }) => {
    const a = 1
    return (
        <>
            {HACK_CSS}
            {usagePatterns.map(({ description, exampleLocations }, index) => (
                <div key={index}>
                    <h4>
                        <Markdown dangerousInnerHTML={renderMarkdown(description)} />
                    </h4>
                    {exampleLocations.map(({ description, location }, index_) => (
                        <div key={index_}>
                            <p>{description}</p>
                            <FileLocations
                                locations={of([
                                    {
                                        uri: makeRepoURI({
                                            repoName: location.resource.repository.name,
                                            commitID: location.resource.commit.oid,
                                            filePath: location.resource.path,
                                        }),
                                        range: location.range!,
                                    },
                                ])}
                                icon={SourceRepositoryIcon}
                                fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                                parentContainerIsEmpty={true}
                                {...props}
                            />
                        </div>
                    ))}
                </div>
            ))}
        </>
    )
}
