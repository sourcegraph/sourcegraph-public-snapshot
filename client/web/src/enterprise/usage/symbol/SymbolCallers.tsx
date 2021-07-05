import H from 'history'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React from 'react'
import { of } from 'rxjs'

import { FileLocations } from '@sourcegraph/branded/src/components/panel/views/FileLocations'
import { Location } from '@sourcegraph/extension-api-types'
import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'

import { SymbolCallerEdgeFields } from '../../../graphql-operations'
import { PersonLink, personLinkFieldsFragment } from '../../../person/PersonLink'
import { fetchHighlightedFileLineRanges } from '../../../repo/backend'

const HACK_CSS = (
    <style>
        {
            'td.line { display: none; } .code-excerpt .code { padding-left: 0.25rem !important; } .result-container__header { display: none; } .result-container { border: solid 1px var(--border-color) !important; border-width: 1px !important; margin: 1rem; }'
        }
    </style>
)

export const SymbolCallerEdgeGQLFragment = gql`
    fragment SymbolCallerEdgeFields on SymbolCallerEdge {
        person {
            ...PersonLinkFields
        }

        locations {
            nodes {
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
    ${personLinkFieldsFragment}
`

interface Props extends SettingsCascadeProps, ThemeProps, VersionContextProps {
    symbolCallers: SymbolCallerEdgeFields[]

    location: H.Location
}

export const SymbolCallersSection: React.FunctionComponent<Props> = ({ symbolCallers, ...props }) => {
    const a = 1
    return (
        <>
            {HACK_CSS}
            {symbolCallers.map(({ person, locations }, index) => (
                <div key={index}>
                    <h3>
                        <PersonLink person={person} /> ({locations.nodes.length})
                    </h3>
                    <FileLocations
                        locations={of(
                            locations.nodes.map<Location>(location => ({
                                uri: makeRepoURI({
                                    repoName: location.resource.repository.name,
                                    commitID: location.resource.commit.oid,
                                    filePath: location.resource.path,
                                }),
                                range: location.range!,
                            }))
                        )}
                        icon={SourceRepositoryIcon}
                        fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                        parentContainerIsEmpty={true}
                        {...props}
                    />
                </div>
            ))}
        </>
    )
}
