import React from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { SourceLocationSetContributorsFields } from '../../../../../graphql-operations'

import { PersonList } from '../../../components/person-list/PersonList'

export const SourceSetContributorsSidebarItem: React.FunctionComponent<{
    contributors: NonNullable<SourceLocationSetContributorsFields['contributors']>
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
