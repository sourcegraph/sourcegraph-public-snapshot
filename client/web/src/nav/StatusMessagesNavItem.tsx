import React from 'react'

import { mdiInformation, mdiAlert, mdiSync, mdiCheckboxMarkedCircle } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'
import { isEqual } from 'lodash'
import { Observable, Subscription } from 'rxjs'
import { catchError, map, repeatWhen, delay, distinctUntilChanged } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, ErrorLike, isErrorLike, repeatUntil } from '@sourcegraph/common'
import { dataOrThrowErrors } from '@sourcegraph/http-client'
import {
    CloudAlertIconRefresh,
    CloudSyncIconRefresh,
    CloudCheckIconRefresh,
} from '@sourcegraph/shared/src/components/icons'
import {
    Button,
    Link,
    Popover,
    PopoverContent,
    PopoverTrigger,
    Position,
    Icon,
    H4,
    Text,
    Tooltip,
} from '@sourcegraph/wildcard'

import { requestGraphQL } from '../backend/graphql'
import { CircleDashedIcon } from '../components/CircleDashedIcon'
import { StatusMessagesResult } from '../graphql-operations'

import { STATUS_MESSAGES } from './StatusMessagesNavItemQueries'

import styles from './StatusMessagesNavItem.module.scss'

function fetchAllStatusMessages(): Observable<StatusMessagesResult['statusMessages']> {
    return requestGraphQL<StatusMessagesResult>(STATUS_MESSAGES).pipe(
        map(dataOrThrowErrors),
        map(data => data.statusMessages)
    )
}

type EntryType = 'not-active' | 'progress' | 'warning' | 'success' | 'error'

interface StatusMessageEntryProps {
    message: string
    linkTo: string
    linkText: string
    entryType: EntryType
    linkOnClick: (event: React.MouseEvent<HTMLAnchorElement, MouseEvent>) => void
    messageHint?: string
    title?: string
}

function entryIcon(entryType: EntryType): JSX.Element {
    switch (entryType) {
        case 'error':
            return (
                <Icon
                    className={classNames('text-danger', styles.icon)}
                    svgPath={mdiInformation}
                    inline={false}
                    aria-label="Error"
                    height={14}
                    width={14}
                />
            )
        case 'warning':
            return (
                <Icon
                    className={classNames('text-warning', styles.icon)}
                    svgPath={mdiAlert}
                    inline={false}
                    aria-label="Warning"
                    height={14}
                    width={14}
                />
            )
        case 'success':
            return (
                <Icon
                    className={classNames('text-success', styles.icon)}
                    svgPath={mdiCheckboxMarkedCircle}
                    inline={false}
                    aria-label="Success"
                    height={14}
                    width={14}
                />
            )
        case 'progress':
            return (
                <Icon
                    className={classNames('text-primary', styles.icon)}
                    svgPath={mdiSync}
                    inline={false}
                    aria-label="In progress"
                    height={14}
                    width={14}
                />
            )
        case 'not-active':
            return (
                <Icon
                    className={classNames(styles.icon, styles.iconOff)}
                    as={CircleDashedIcon}
                    inline={false}
                    aria-label="Not active"
                    height={16}
                    width={16}
                />
            )
    }
}

const getMessageColor = (entryType: EntryType): string => {
    switch (entryType) {
        case 'error':
            return styles.messageError
        case 'warning':
            return styles.messageWarning
        default:
            return ''
    }
}

const getBorderClassname = (entryType: EntryType): string => {
    switch (entryType) {
        case 'error':
            return styles.entryBorderError
        case 'warning':
            return styles.entryBorderWarning
        case 'success':
            return styles.entryBorderSuccess
        case 'progress':
            return styles.entryBorderProgress
        default:
            return ''
    }
}

const StatusMessagesNavItemEntry: React.FunctionComponent<React.PropsWithChildren<StatusMessageEntryProps>> = props => {
    const onLinkClick = (event: React.MouseEvent<HTMLAnchorElement, MouseEvent>): void => {
        props.linkOnClick(event)
    }

    return (
        <div key={props.message} className={styles.entry}>
            <H4 className="d-flex align-items-center mb-0">
                {entryIcon(props.entryType)}
                {props.title ? props.title : 'Your repositories'}
            </H4>
            {props.entryType === 'not-active' ? (
                <div className={classNames('status-messages-nav-item__entry-card border-0', styles.cardInactive)}>
                    <Text className={classNames('text-muted', styles.message)}>{props.message}</Text>
                    <Link className="text-primary" to={props.linkTo} onClick={onLinkClick}>
                        {props.linkText}
                    </Link>
                </div>
            ) : (
                <div
                    className={classNames(
                        'status-messages-nav-item__entry-card',
                        styles.cardActive,
                        getBorderClassname(props.entryType)
                    )}
                >
                    <Text className={classNames(styles.message, getMessageColor(props.entryType))}>
                        {props.message}
                    </Text>
                    {props.messageHint && (
                        <>
                            <small className="text-muted d-inline-block mb-1">{props.messageHint}</small>
                            <br />
                        </>
                    )}
                    <Link className="text-primary" to={props.linkTo} onClick={onLinkClick}>
                        {props.linkText}
                    </Link>
                </div>
            )}
        </div>
    )
}

interface Props {
    history: H.History
    fetchMessages?: () => Observable<StatusMessagesResult['statusMessages']>
}

type Messages = StatusMessagesResult['statusMessages']
type MessagesOrError = Messages | ErrorLike

interface State {
    messagesOrError: MessagesOrError
    isOpen: boolean
}

const REFRESH_INTERVAL_MS = 60000

/**
 * Displays a status icon in the navbar reflecting the completion of backend
 * tasks such as repository cloning, and exposes a dropdown menu containing
 * more information on these tasks.
 */
export class StatusMessagesNavItem extends React.PureComponent<Props, State> {
    private subscriptions = new Subscription()

    public state: State = { isOpen: false, messagesOrError: [] }

    private toggleIsOpen = (): void => {
        this.setState(previousState => ({ isOpen: !previousState.isOpen }))
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            (this.props.fetchMessages ?? fetchAllStatusMessages)()
                .pipe(
                    catchError(error => [asError(error) as ErrorLike]),
                    // Poll on REFRESH_INTERVAL_MS
                    repeatUntil(messagesOrError => isErrorLike(messagesOrError), { delay: REFRESH_INTERVAL_MS }),
                    repeatWhen(completions => completions.pipe(delay(REFRESH_INTERVAL_MS))),
                    distinctUntilChanged(isEqual)
                )
                .subscribe(messagesOrError => {
                    this.setState({ messagesOrError })
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private renderMessage(messages: Messages): JSX.Element | JSX.Element[] {
        const links = {
            viewRepositories: '/site-admin/repositories',
            manageRepositories: '/site-admin/external-services',
            manageCodeHosts: '/site-admin/external-services',
            getCodeHostLink: (id: string) => `/site-admin/external-services/${id}`,
        }

        // no status messages
        if (messages.length === 0) {
            return (
                <StatusMessagesNavItemEntry
                    key="up-to-date"
                    message="Repositories available for search"
                    linkTo={links.viewRepositories}
                    linkText="View repositories"
                    linkOnClick={this.toggleIsOpen}
                    entryType="success"
                />
            )
        }

        return messages.map(status => {
            switch (status.type) {
                case 'CloningProgress':
                    return (
                        <StatusMessagesNavItemEntry
                            key={status.message}
                            message={status.message}
                            messageHint="Your repositories may not be up-to-date."
                            linkTo={links.viewRepositories}
                            linkText="View status"
                            linkOnClick={this.toggleIsOpen}
                            entryType="progress"
                        />
                    )
                case 'ExternalServiceSyncError':
                    return (
                        <StatusMessagesNavItemEntry
                            key={status.externalService.id}
                            message={`Can't connect to ${status.externalService.displayName}`}
                            messageHint="Verify the code host configuration."
                            linkTo={links.getCodeHostLink(status.externalService.id)}
                            linkText="Manage code hosts"
                            linkOnClick={this.toggleIsOpen}
                            entryType="error"
                        />
                    )
                case 'SyncError':
                    return (
                        <StatusMessagesNavItemEntry
                            key={status.message}
                            message={status.message}
                            messageHint="Your repositories may not be up-to-date."
                            linkTo={links.viewRepositories + '?status=failed-fetch'}
                            linkText="Manage repositories"
                            linkOnClick={this.toggleIsOpen}
                            entryType="error"
                        />
                    )
            }
        })
    }

    private renderIcon(): JSX.Element | null {
        if (isErrorLike(this.state.messagesOrError)) {
            return (
                <Tooltip content="Sorry, we couldn’t fetch notifications!">
                    <Icon aria-label="Sorry, we couldn’t fetch notifications!" as={CloudAlertIconRefresh} size="md" />
                </Tooltip>
            )
        }

        let codeHostMessage
        let icon
        if (
            this.state.messagesOrError.some(({ type }) => type === 'ExternalServiceSyncError' || type === 'SyncError')
        ) {
            codeHostMessage = 'Syncing repositories failed!'
            icon = CloudAlertIconRefresh
        } else if (this.state.messagesOrError.some(({ type }) => type === 'CloningProgress')) {
            codeHostMessage = 'Cloning repositories...'
            icon = CloudSyncIconRefresh
        } else {
            codeHostMessage = 'Repositories up-to-date'
            icon = CloudCheckIconRefresh
        }

        return (
            <Tooltip content={this.state.isOpen ? undefined : codeHostMessage}>
                <Icon
                    as={icon}
                    size="md"
                    {...(this.state.isOpen ? { 'aria-hidden': true } : { 'aria-label': codeHostMessage })}
                />
            </Tooltip>
        )
    }

    public render(): JSX.Element | null {
        return (
            <Popover isOpen={this.state.isOpen} onOpenChange={event => this.setState({ isOpen: event.isOpen })}>
                <PopoverTrigger
                    className="nav-link py-0 px-0 percy-hide chromatic-ignore"
                    as={Button}
                    variant="link"
                    aria-label={this.state.isOpen ? 'Hide status messages' : 'Show status messages'}
                >
                    {this.renderIcon()}
                </PopoverTrigger>

                <PopoverContent position={Position.bottomEnd} className={classNames('p-0', styles.dropdownMenu)}>
                    <div className={styles.dropdownMenuContent}>
                        <small className={classNames('d-inline-block text-muted', styles.sync)}>Code sync status</small>
                        {isErrorLike(this.state.messagesOrError) ? (
                            <ErrorAlert
                                className={styles.entry}
                                prefix="Failed to load status messages"
                                error={this.state.messagesOrError}
                            />
                        ) : (
                            this.renderMessage(this.state.messagesOrError)
                        )}
                    </div>
                </PopoverContent>
            </Popover>
        )
    }
}
