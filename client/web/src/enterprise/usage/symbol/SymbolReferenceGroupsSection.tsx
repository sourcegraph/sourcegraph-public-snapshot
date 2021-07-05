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

import { SymbolReferenceGroup } from '../../../graphql-operations'
import { fetchHighlightedFileLineRanges } from '../../../repo/backend'

const HACK_CSS = (
    <style>
        {
            'td.line { display: none; } .code-excerpt .code { padding-left: 0.25rem !important; } .result-container__header { display: none; } .result-container { border: solid 1px var(--border-color) !important; border-width: 1px !important; margin: 1rem; }'
        }
    </style>
)

export const SymbolReferenceGroupGQLFragment = gql`
    fragment SymbolReferenceGroup on ReferenceGroup {
        name

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
`

interface Props extends SettingsCascadeProps, ThemeProps, VersionContextProps {
    referenceGroups: SymbolReferenceGroup[]

    location: H.Location
}

export const SymbolReferenceGroupsSection: React.FunctionComponent<Props> = ({ referenceGroups, ...props }) => {
    const a = 1
    return (
        <>
            {HACK_CSS}
            {referenceGroups.map((referenceGroup, index) => (
                <div key={index}>
                    <h3>{referenceGroup.name}</h3>
                    <FileLocations
                        locations={of(
                            referenceGroup.locations.nodes.map<Location>(reference => ({
                                uri: makeRepoURI({
                                    repoName: reference.resource.repository.name,
                                    commitID: reference.resource.commit.oid,
                                    filePath: reference.resource.path,
                                }),
                                range: reference.range!,
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
