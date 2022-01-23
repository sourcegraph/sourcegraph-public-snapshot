import React from 'react'

import { gql } from '@sourcegraph/http-client'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { SourceSetCodeOwnersFields } from '../../../../../graphql-operations'
import { PersonList } from '../../../components/person-list/PersonList'

// TODO(sqs): dont fetch all
export const SOURCE_SET_CODE_OWNERS_FRAGMENT = gql`
    fragment SourceSetCodeOwnersFields on SourceSet {
        codeOwners {
            edges {
                node {
                    ...PersonLinkFields
                    avatarURL
                }
                fileCount
                fileProportion
            }
            totalCount
            pageInfo {
                hasNextPage
            }
        }
    }
`

export const SourceSetCodeOwnersSidebarItem: React.FunctionComponent<{
    codeOwners: NonNullable<SourceSetCodeOwnersFields['codeOwners']>
    titleLink: string
    className?: string
}> = ({ codeOwners, titleLink, className }) => (
    <PersonList
        title="Code owners"
        titleLink={titleLink}
        titleCount={codeOwners.totalCount}
        listTag="ol"
        orientation="summary"
        items={codeOwners.edges.map(codeOwner => ({
            person: codeOwner.node,
            text: codeOwner.fileProportion >= 0.01 ? `${(codeOwner.fileProportion * 100).toFixed(0)}%` : '<1%',
            textTooltip: `Owns ${codeOwner.fileCount} ${pluralize('line', codeOwner.fileCount)}`,
        }))}
        className={className}
    />
)
