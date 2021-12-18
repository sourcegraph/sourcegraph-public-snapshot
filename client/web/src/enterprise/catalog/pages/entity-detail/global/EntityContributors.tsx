import React from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { ComponentAuthorsFields } from '../../../../../graphql-operations'

import { PersonList } from './PersonList'

interface Props {
    component: ComponentAuthorsFields
    className?: string
}

export const EntityContributors: React.FunctionComponent<Props> = ({ component: { authors }, className }) => (
    <PersonList
        title="Contributors"
        listTag="ol"
        orientation="vertical"
        items={
            authors
                ? authors.map(author => ({
                      person: author.person,
                      text:
                          author.authoredLineProportion >= 0.01
                              ? `${(author.authoredLineProportion * 100).toFixed(0)}%`
                              : '<1%',
                      textTooltip: `${author.authoredLineCount} ${pluralize('line', author.authoredLineCount)}`,
                      date: author.lastCommit.author.date,
                  }))
                : []
        }
        className={className}
    />
)
