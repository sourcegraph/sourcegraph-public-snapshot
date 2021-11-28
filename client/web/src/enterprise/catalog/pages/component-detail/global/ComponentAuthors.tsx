import classNames from 'classnames'
import React from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'

import { Timestamp } from '../../../../../components/time/Timestamp'
import { CatalogComponentAuthorsFields } from '../../../../../graphql-operations'
import { PersonLink } from '../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../user/UserAvatar'

import styles from './ComponentAuthors.module.scss'

interface Props {
    catalogComponent: CatalogComponentAuthorsFields
    className?: string
    headerClassName?: string
    titleClassName?: string
    bodyClassName?: string
}

export const ComponentAuthors: React.FunctionComponent<Props> = ({
    catalogComponent: { authors },
    className,
    headerClassName,
    titleClassName,
    bodyClassName,
}) =>
    authors && authors.length > 0 ? (
        <div className={className}>
            <header className={headerClassName}>
                <h3 className={titleClassName}>Authors</h3>
            </header>
            <ol className={classNames('list-group list-group-horizontal', bodyClassName)}>
                {authors.map(author => (
                    <li
                        key={author.person.email}
                        className={classNames('list-group-item text-center pt-2', styles.author)}
                    >
                        <div>
                            <UserAvatar className="icon-inline" user={author.person} />
                        </div>
                        <PersonLink person={author.person} className="text-muted small text-truncate d-block" />
                        <div
                            className={classNames(styles.percent)}
                            title={`${author.authoredLineCount} ${pluralize('line', author.authoredLineCount)}`}
                        >
                            {author.authoredLineProportion >= 0.01
                                ? `${(author.authoredLineProportion * 100).toFixed(0)}%`
                                : '<1%'}
                        </div>
                        <div className={classNames('text-muted', styles.lastCommit)}>
                            <Timestamp date={author.lastCommit.author.date} noAbout={true} />
                        </div>
                    </li>
                ))}
            </ol>
        </div>
    ) : (
        <p>Unable to determine authorship</p>
    )
