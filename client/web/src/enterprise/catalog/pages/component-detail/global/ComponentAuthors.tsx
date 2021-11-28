import React from 'react'

import { Timestamp } from '../../../../../components/time/Timestamp'
import { CatalogComponentAuthorsFields } from '../../../../../graphql-operations'
import { PersonLink } from '../../../../../person/PersonLink'

interface Props {
    catalogComponent: CatalogComponentAuthorsFields
    className?: string
    headerClassName?: string
    titleClassName?: string
}

export const ComponentAuthors: React.FunctionComponent<Props> = ({
    catalogComponent: { authors },
    className,
    headerClassName,
    titleClassName,
}) =>
    authors && authors.length > 0 ? (
        <div className={className}>
            <header className={headerClassName}>
                <h3 className={titleClassName}>Authors</h3>
            </header>
            <ol className="list-group list-group-flush">
                {authors.slice(0, 100 /* TODO(sqs): show all */).map(author => (
                    <li key={author.person.email} className="list-group-item">
                        <PersonLink person={author.person} /> {(author.authoredLineProportion * 100).toFixed(1)}%{' '}
                        <Timestamp date={author.lastCommit.author.date} />
                    </li>
                ))}
            </ol>
        </div>
    ) : (
        <p>No changes found</p>
    )
