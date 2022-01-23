import classNames from 'classnames'
import React from 'react'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { Badge } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../../components/time/Timestamp'
import { PersonLinkFields } from '../../../../graphql-operations'
import { PersonLink } from '../../../../person/PersonLink'
import { UserAvatar } from '../../../../user/UserAvatar'
import { formatDistanceShortStrict } from '../../../../util/date'
import { ScrollListHorizontal, ScrollListVertical } from '../scroll-list/ScrollList'

import styles from './PersonList.module.scss'

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
        'title' | 'titleLink' | 'titleCount' | 'description' | 'listTag' | 'className' | 'listClassName'
    > {
    items: Item[]
    orientation: 'horizontal' | 'vertical' | 'summary' | 'summary2'
    primaryText?: 'person' | 'text'
    className?: string
}

export const PersonList: React.FunctionComponent<Props> = ({ items, orientation, primaryText = 'text', ...props }) =>
    orientation === 'summary' ? (
        <ListSummary
            {...props}
            items={items}
            detailItemComponent={PersonListItemVerticalOneLine}
            otherItemComponent={PersonListItemOther}
        />
    ) : orientation === 'summary2' ? (
        <ListSummary2 {...props} items={items} itemComponent={PersonListItemSummary2Component} />
    ) : orientation === 'horizontal' ? (
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
            {items.map(item => (
                <PersonListItemVertical
                    key={item.person.email}
                    item={item}
                    primaryText={primaryText}
                    avatarSize={28}
                    className="list-group-item p-2"
                />
            ))}
        </ScrollListVertical>
    )

const PersonListItemVertical: React.FunctionComponent<{
    item: Item
    avatarSize?: number
    primaryText: Props['primaryText']
    className?: string
}> = ({ item: { person, text, textTooltip, date, action }, avatarSize, primaryText, className }) => (
    <li className={classNames('d-flex flex-wrap align-items-center', className)}>
        <UserAvatar className="icon-inline mr-2 " user={person} size={avatarSize} />
        <div className="flex-1 mr-2">
            <PersonLink
                person={person}
                className={classNames('small text-truncate d-block', primaryText === 'text' && 'text-muted')}
            />
            <div className={classNames(styles.itemText, primaryText === 'person' && 'text-muted')} title={textTooltip}>
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
)

const PersonListItemVerticalOneLine: React.FunctionComponent<{
    item: Omit<Item, 'action'>
    avatarSize?: number
    primaryText: Props['primaryText']
    className?: string
}> = ({ item: { person, text, textTooltip, date }, avatarSize, primaryText, className }) => (
    <li className={classNames('d-flex align-items-center', className)}>
        <UserAvatar className="icon-inline mr-2 flex-shrink-0" user={person} size={avatarSize} />
        <PersonLink
            person={person}
            className={classNames('small text-truncate d-block mr-2', primaryText === 'text' && 'text-muted')}
        />
        <div
            className={classNames('text-nowrap', styles.itemText, primaryText === 'person' && 'text-muted')}
            title={textTooltip}
        >
            {text}
            {date && (
                <span className={classNames('ml-1 text-muted', styles.itemDate)}>
                    <Timestamp date={date} noAbout={true} />
                </span>
            )}
        </div>
    </li>
)

const PersonListItemOther: React.FunctionComponent<{
    item: Omit<Item, 'text' | 'date' | 'action'>
    avatarSize?: number
    className?: string
}> = ({ item: { person, textTooltip }, avatarSize, className }) => (
    <li className={classNames(className)}>
        <LinkOrSpan to={person.user?.url} title={textTooltip}>
            <UserAvatar className="icon-inline mr-2 " user={person} size={avatarSize} />
        </LinkOrSpan>
    </li>
)

const DEFAULT_LIST_SUMMARY_DETAIL = 3
const DEFAULT_LIST_SUMMARY_OTHER = 6

const ListSummary: React.FunctionComponent<
    Omit<Props, 'orientation'> & {
        detailItemComponent: React.ComponentType<{
            item: Omit<Item, 'action'>
            primaryText: Props['primaryText']
            avatarSize?: number
            className?: string
        }>
        otherItemComponent: React.ComponentType<{
            item: Omit<Item, 'text' | 'date' | 'action'>
            avatarSize?: number
            className?: string
        }>
    }
> = ({
    title,
    titleLink,
    titleCount,
    description,
    items,
    primaryText,
    listTag: ListTag,
    detailItemComponent: DetailItemComponent,
    otherItemComponent: OtherItemComponent,
    listClassName,
    className,
}) => (
    <div className={className}>
        <h4 className="font-weight-bold">
            <LinkOrSpan to={titleLink} className="text-body">
                {title}
                {titleCount !== undefined && (
                    <Badge variant="secondary" small={true} pill={true} className="ml-1">
                        {titleCount}
                    </Badge>
                )}
            </LinkOrSpan>
        </h4>
        {description}
        <ListTag className={classNames('list-unstyled d-flex flex-wrap align-items-center', listClassName)}>
            {items.slice(0, DEFAULT_LIST_SUMMARY_DETAIL).map(item => (
                <DetailItemComponent
                    key={item.person.email}
                    item={item}
                    avatarSize={19}
                    primaryText={primaryText}
                    className="w-100 pb-2"
                />
            ))}
            {items
                .slice(DEFAULT_LIST_SUMMARY_DETAIL, DEFAULT_LIST_SUMMARY_DETAIL + DEFAULT_LIST_SUMMARY_OTHER)
                .map(item => (
                    <OtherItemComponent
                        key={item.person.email}
                        item={{
                            ...item,
                            // TODO(sqs): better textTooltip
                            textTooltip: `${item.person.displayName}: ${
                                typeof item.text === 'string' ? item.text : ''
                            } (${item.textTooltip || ''})`,
                        }}
                        avatarSize={19}
                        className="pb-1"
                    />
                ))}
        </ListTag>
    </div>
)

const LIST_SUMMARY2_MAX = 3

const ListSummary2: React.FunctionComponent<
    Omit<Props, 'orientation'> & {
        itemComponent: React.ComponentType<{
            item: Omit<Item, 'action'>
            avatarSize: number
            className?: string
        }>
    }
> = ({
    title,
    titleLink,
    titleCount,
    description,
    items,
    listTag: ListTag,
    itemComponent: ItemComponent,
    listClassName,
    className,
}) => (
    <div className={className}>
        <h4 className="font-weight-bold">
            <LinkOrSpan to={titleLink} className="text-body">
                {title}
                {titleCount !== undefined && (
                    <Badge variant="secondary" small={true} pill={true} className="ml-1">
                        {titleCount}
                    </Badge>
                )}
            </LinkOrSpan>
        </h4>
        {description}
        <ListTag className={classNames('list-unstyled mb-0', styles.listSummary2, listClassName)}>
            {items.slice(0, LIST_SUMMARY2_MAX).map(item => (
                <ItemComponent key={item.person.email} item={item} avatarSize={19} className="pb-2" />
            ))}
        </ListTag>
    </div>
)

const PersonListItemSummary2Component: React.FunctionComponent<{
    item: Omit<Item, 'action'>
    avatarSize: number
    className?: string
}> = ({ item: { person, text, textTooltip, date }, avatarSize, className }) => (
    <li className={classNames(className)}>
        <LinkOrSpan
            to={person.user?.url}
            title={
                // TODO(sqs): better textTooltip
                `${person.displayName}: ${typeof text === 'string' ? text : ''} (${textTooltip || ''})`
            }
            className="d-flex align-items-stretch"
        >
            <UserAvatar
                className={classNames(styles.listItemSummary2ComponentAvatar)}
                user={person}
                size={avatarSize}
            />
            <span
                className="small border border-left-0 pr-1 text-nowrap"
                style={{
                    position: 'relative',
                    left: `${((-1 * avatarSize) / 2).toFixed(2)}px`,
                    paddingLeft: `calc(0.25*var(--spacer) + ${(avatarSize / 2).toFixed(2)}px)`,
                    zIndex: 0,
                    fontSize: '80%',
                }}
            >
                {text} {date && <span className="text-muted">{formatDistanceShortStrict(date)}</span>}
            </span>
        </LinkOrSpan>
    </li>
)
