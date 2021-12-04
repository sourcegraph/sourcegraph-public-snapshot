import React from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { CatalogComponentAuthorsFields } from '../../../../../graphql-operations'

import { PersonList } from './PersonList'

interface Props {
    catalogComponent: CatalogComponentAuthorsFields
    className?: string
}

export const ComponentAuthors: React.FunctionComponent<Props> = ({ catalogComponent: { authors }, className }) => (
    <PersonList
        title="Authors"
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
