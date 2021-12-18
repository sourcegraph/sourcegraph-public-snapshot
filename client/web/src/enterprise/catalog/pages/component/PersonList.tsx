import classNames from 'classnames'
import React from 'react'

import { Timestamp } from '../../../../components/time/Timestamp'
import { PersonLinkFields } from '../../../../graphql-operations'
import { PersonLink } from '../../../../person/PersonLink'
import { UserAvatar } from '../../../../user/UserAvatar'

import styles from './PersonList.module.scss'
import { ScrollListHorizontal, ScrollListVertical } from './ScrollList'

interface Item {
    person: PersonLinkFields & { avatarURL: string | null }
    text: string | React.ReactFragment
    textTooltip?: string
    date?: string
    action?: React.ReactFragment
}

interface Props
    extends Pick<
        React.ComponentPropsWithoutRef<typeof ScrollListHorizontal>,
        'title' | 'description' | 'listTag' | 'className' | 'listClassName'
    > {
    items: Item[]
    orientation: 'horizontal' | 'vertical'
    primaryText?: 'person' | 'text'
    className?: string
}

export const PersonList: React.FunctionComponent<Props> = ({ items, orientation, primaryText = 'text', ...props }) =>
    orientation === 'horizontal' ? (
        <ScrollListHorizontal {...props}>
            {items.map(({ person, text, textTooltip, date }) => (
                <li
                    key={person.email}
                    className={classNames('list-group-item text-center pt-2', styles.itemHorizontal)}
                >
                    <UserAvatar className="icon-inline" user={person} />
                    <PersonLink
                        person={person}
                        className={classNames('small text-truncate d-block', primaryText === 'text' && 'text-muted')}
                    />
                    <div
                        className={classNames(styles.itemText, primaryText === 'person' && 'text-muted')}
                        title={textTooltip}
                    >
                        {text}
                    </div>
                    {date && (
                        <div className={classNames('text-muted', styles.itemDate)}>
                            <Timestamp date={date} noAbout={true} />
                        </div>
                    )}
                </li>
            ))}
        </ScrollListHorizontal>
    ) : (
        <ScrollListVertical {...props}>
            {items.map(({ person, text, textTooltip, date, action }) => (
                <li key={person.email} className={classNames('list-group-item p-2 d-flex align-items-center')}>
                    <UserAvatar className="icon-inline mr-2 " user={person} size={28} />
                    <div className="flex-1">
                        <PersonLink
                            person={person}
                            className={classNames(
                                'small text-truncate d-block',
                                primaryText === 'text' && 'text-muted'
                            )}
                        />
                        <div
                            className={classNames(styles.itemText, primaryText === 'person' && 'text-muted')}
                            title={textTooltip}
                        >
                            {text}
                            {date && (
                                <span className={classNames('ml-1 text-muted', styles.itemDate)}>
                                    <Timestamp date={date} noAbout={true} />
                                </span>
                            )}
                        </div>
                    </div>
                    {action}
                </li>
            ))}
        </ScrollListVertical>
    )
