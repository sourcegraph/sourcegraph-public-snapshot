import React from 'react'

import { gql } from '@sourcegraph/http-client'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { SourceSetContributorsFields } from '../../../../../graphql-operations'
import { personLinkFieldsFragment } from '../../../../../person/PersonLink'
import { PersonList } from '../../../components/person-list/PersonList'

// TODO(sqs): dont fetch all
export const SOURCE_SET_CONTRIBUTORS_FRAGMENT = gql`
    fragment SourceSetContributorsFields on SourceSet {
        contributors {
            edges {
                person {
                    ...PersonLinkFields
                    avatarURL
                }
                authoredLineCount
                authoredLineProportion
                lastCommit {
                    author {
                        date
                    }
                }
            }
            totalCount
            pageInfo {
                hasNextPage
            }
        }
    }
    ${personLinkFieldsFragment}
`

export const SourceSetContributorsSidebarItem: React.FunctionComponent<{
    contributors: NonNullable<SourceSetContributorsFields['contributors']>
    titleLink: string
    className?: string
}> = ({ contributors, titleLink, className }) => (
    // TODO(sqs): For this, could show a visualization horizontal bar where width = % of person's contributions, bg color is recency of last contribution, and text overlay is the person's name
    <PersonList
        title="Contributors"
        titleLink={titleLink}
        titleCount={contributors.totalCount}
        listTag="ol"
        orientation="summary"
        items={contributors.edges.map(contributor => ({
            person: contributor.person,
            text:
                contributor.authoredLineProportion >= 0.01
                    ? `${(contributor.authoredLineProportion * 100).toFixed(0)}%`
                    : '<1%',
            textTooltip: `${contributor.authoredLineCount} ${pluralize('line', contributor.authoredLineCount)}`,
            date: contributor.lastCommit.author.date,
        }))}
        className={className}
    />
)
