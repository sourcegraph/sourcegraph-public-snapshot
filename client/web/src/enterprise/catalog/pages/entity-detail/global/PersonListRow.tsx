import classNames from 'classnames'
import React from 'react'

import { Timestamp } from '../../../../../components/time/Timestamp'
import { PersonLinkFields } from '../../../../../graphql-operations'
import { PersonLink } from '../../../../../person/PersonLink'
import { UserAvatar } from '../../../../../user/UserAvatar'

import styles from './PersonListRow.module.scss'
import { ScrollListRow } from './ScrollListRow'

interface Item {
    person: PersonLinkFields & { avatarURL: string | null }
    text: string
    textTooltip?: string
    date?: string
}

interface Props extends Pick<React.ComponentPropsWithoutRef<typeof ScrollListRow>, 'title' | 'listTag' | 'className'> {
    items: Item[]
    className?: string
}

export const PersonListRow: React.FunctionComponent<Props> = ({ items, ...props }) => (
    <ScrollListRow {...props}>
        {items.map(({ person, text, textTooltip, date }) => (
            <li key={person.email} className={classNames('list-group-item text-center pt-2', styles.item)}>
                <UserAvatar className="icon-inline" user={person} />
                <PersonLink person={person} className="text-muted small text-truncate d-block" />
                <div className={classNames(styles.itemText)} title={textTooltip}>
                    {text}
                </div>
                {date && (
                    <div className={classNames('text-muted', styles.itemDate)}>
                        <Timestamp date={date} noAbout={true} />
                    </div>
                )}
            </li>
        ))}
    </ScrollListRow>
)
