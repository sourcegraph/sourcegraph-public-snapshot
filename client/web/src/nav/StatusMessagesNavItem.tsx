import React from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { isEqual, upperFirst } from 'lodash'
import AlertIcon from 'mdi-react/AlertIcon'
import CheckboxCircleIcon from 'mdi-react/CheckboxMarkedCircleIcon'
import CloudOffOutlineIcon from 'mdi-react/CloudOffOutlineIcon'
import InformationCircleIcon from 'mdi-react/InformationCircleIcon'
import SyncIcon from 'mdi-react/SyncIcon'
import { Observable, Subscription, of } from 'rxjs'
import { catchError, map, repeatWhen, delay, distinctUntilChanged, switchMap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, ErrorLike, isErrorLike, repeatUntil } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
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
    Typography,
} from '@sourcegraph/wildcard'

import { requestGraphQL } from '../backend/graphql'
import { CircleDashedIcon } from '../components/CircleDashedIcon'
import { queryExternalServices } from '../components/externalServices/backend'
import { StatusMessagesResult } from '../graphql-operations'
import { eventLogger } from '../tracking/eventLogger'

import styles from './StatusMessagesNavItem.module.scss'

function fetchAllStatusMessages(): Observable<StatusMessagesResult['statusMessages']> {
    return requestGraphQL<StatusMessagesResult>(
        gql`
            query StatusMessages {
                statusMessages {
                    ...StatusMessageFields
                }
            }

            fragment StatusMessageFields on StatusMessage {
                type: __typename

                ... on CloningProgress {
                    message
                }

                ... on IndexingProgress {
                    message
                }

                ... on SyncError {
                    message
                }

                ... on IndexingError {
                    message
                }

                ... on ExternalServiceSyncError {
                    message
                    externalService {
                        id
                        displayName
                    }
                }
            }
        `
    ).pipe(
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
    progressHint?: string
    title?: string
}

function entryIcon(entryType: EntryType): JSX.Element {
    switch (entryType) {
        case 'error':
            return <InformationCircleIcon size={14} className={classNames('text-danger', styles.icon)} />
        case 'warning':
            return <AlertIcon size={14} className={classNames('text-warning', styles.icon)} />
        case 'success':
            return <CheckboxCircleIcon size={14} className={classNames('text-success', styles.icon)} />
        case 'progress':
            return <SyncIcon size={14} className={classNames('text-primary', styles.icon)} />
        case 'not-active':
            return <CircleDashedIcon size={16} className={classNames(styles.icon, styles.iconOff)} />
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
        const payload = { notificationType: props.entryType }
        eventLogger.log('UserNotificationsLinkClicked', payload, payload)
        props.linkOnClick(event)
    }

    return (
        <div key={props.message} className={styles.entry}>
            <Typography.H4 className="d-flex align-items-center mb-0">
                {entryIcon(props.entryType)}
                {props.title ? props.title : 'Your repositories'}
            </Typography.H4>
            {props.entryType === 'not-active' ? (
                <div className={classNames('status-messages-nav-item__entry-card border-0', styles.cardInactive)}>
                    <p className={classNames('text-muted', styles.message)}>{props.message}</p>
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
                    <p className={classNames(styles.message, getMessageColor(props.entryType))}>{props.message}</p>
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
            {props.progressHint && <small className="text-muted">{props.progressHint}</small>}
        </div>
    )
}

interface User {
    id: string
    username: string
    isSiteAdmin: boolean
}

interface Props {
    user: User
    history: H.History
    fetchMessages?: () => Observable<StatusMessagesResult['statusMessages']>
}

enum ExternalServiceNoActivityReasons {
    NoCodehosts = 'NoCodehosts',
    NoRepos = 'NoRepos',
}

type ExternalServiceNoActivityReason = keyof typeof ExternalServiceNoActivityReasons
type Message = StatusMessagesResult['statusMessages'] | ExternalServiceNoActivityReason
type MessageOrError = Message | ErrorLike

const isNoActivityReason = (status: MessageOrError): status is ExternalServiceNoActivityReason =>
    typeof status === 'string'

interface State {
    messagesOrError: MessageOrError
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
        let first = true
        this.subscriptions.add(
            queryExternalServices({
                namespace: this.props.user.id,
                first: null,
                after: null,
            })
                .pipe(
                    switchMap(({ nodes: services }) => {
                        if (!this.props.user.isSiteAdmin) {
                            if (services.length === 0) {
                                return of(ExternalServiceNoActivityReasons.NoCodehosts)
                            }

                            if (
                                !services.some(service => service.repoCount !== 0) &&
                                services.every(service => service.lastSyncError === null && service.warning === null)
                            ) {
                                return of(ExternalServiceNoActivityReasons.NoRepos)
                            }
                        }

                        return (this.props.fetchMessages ?? fetchAllStatusMessages)()
                    }),
                    catchError(error => [asError(error) as ErrorLike]),
                    // Poll on REFRESH_INTERVAL_MS, or REFRESH_INTERVAL_AFTER_ERROR_MS if there is an error.
                    repeatUntil(messagesOrError => isErrorLike(messagesOrError), { delay: REFRESH_INTERVAL_MS }),
                    repeatWhen(completions => completions.pipe(delay(REFRESH_INTERVAL_MS))),
                    distinctUntilChanged((a, b) => isEqual(a, b))
                )
                .subscribe(messagesOrError => {
                    this.setState({ messagesOrError })

                    if (first) {
                        this.trackUserNotificationsEvent('loaded')
                        first = false
                    }
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private renderMessage(noActivityOrStatus: Message, isSiteAdmin: boolean): JSX.Element | JSX.Element[] {
        const userSettings = `/users/${this.props.user.username}/settings`
        const makeGetCodeHostLink = (isSiteAdmin: boolean) => (id: string): string =>
            isSiteAdmin ? `/site-admin/external-services/${id}` : `${userSettings}/code-hosts`

        const roleLinks = {
            admin: {
                viewRepositories: '/site-admin/repositories',
                manageRepositories: '/site-admin/external-services',
                manageCodeHosts: '/site-admin/external-services',
                getCodeHostLink: makeGetCodeHostLink(true),
            },
            nonAdmin: {
                viewRepositories: `${userSettings}/repositories`,
                manageRepositories: `${userSettings}/repositories/manage`,
                manageCodeHosts: `${userSettings}/code-hosts`,
                getCodeHostLink: makeGetCodeHostLink(false),
            },
        }

        const links = isSiteAdmin ? roleLinks.admin : roleLinks.nonAdmin

        // no status messages
        if (Array.isArray(noActivityOrStatus) && noActivityOrStatus.length === 0) {
            return (
                <StatusMessagesNavItemEntry
                    key="up-to-date"
                    message="Repositories available for search"
                    linkTo={links.viewRepositories}
                    linkText="Manage repositories"
                    linkOnClick={this.toggleIsOpen}
                    entryType="success"
                />
            )
        }

        // no code hosts or no repos
        if (isNoActivityReason(noActivityOrStatus)) {
            if (noActivityOrStatus === ExternalServiceNoActivityReasons.NoRepos) {
                return (
                    <StatusMessagesNavItemEntry
                        key={noActivityOrStatus}
                        message="Add repositories to start searching your code on Sourcegraph."
                        linkTo={links.manageRepositories}
                        linkText="Add repositories"
                        linkOnClick={this.toggleIsOpen}
                        entryType="not-active"
                    />
                )
            }
            return (
                <StatusMessagesNavItemEntry
                    key={noActivityOrStatus}
                    message="Connect with a code host to start adding your code to Sourcegraph."
                    linkTo={links.manageCodeHosts}
                    linkText="Connect with code host"
                    linkOnClick={this.toggleIsOpen}
                    entryType="not-active"
                />
            )
        }

        const cloningProgress = noActivityOrStatus.find(message_ => message_.type === 'CloningProgress')
        const indexing = noActivityOrStatus.find(message_ => message_.type === 'IndexingProgress')

        if (cloningProgress && indexing) {
            noActivityOrStatus = noActivityOrStatus.filter(message_ => message_.type !== 'IndexingProgress')
        }

        return noActivityOrStatus.map(status => {
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
                            progressHint={indexing && indexing.type === 'IndexingProgress' ? indexing.message : ''}
                        />
                    )
                case 'IndexingProgress':
                    return (
                        <StatusMessagesNavItemEntry
                            key={status.message}
                            message="Repositories available for search"
                            linkTo={links.viewRepositories}
                            linkText="Manage repositories"
                            linkOnClick={this.toggleIsOpen}
                            entryType="success"
                            progressHint={status.message}
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
                            linkTo={links.viewRepositories}
                            linkText="Manage repositories"
                            linkOnClick={this.toggleIsOpen}
                            entryType="error"
                        />
                    )
                case 'IndexingError':
                    return (
                        <StatusMessagesNavItemEntry
                            key={status.message}
                            message={status.message}
                            messageHint="Your repositories are up-to-date, but search speed may be slower than usual."
                            linkTo={links.viewRepositories}
                            linkText="Troubleshoot"
                            linkOnClick={this.toggleIsOpen}
                            entryType="warning"
                        />
                    )
            }
        })
    }

    private renderIcon(): JSX.Element | null {
        if (isErrorLike(this.state.messagesOrError)) {
            return (
                <Icon
                    role="img"
                    data-tooltip="Sorry, we couldn’t fetch notifications!"
                    as={CloudAlertIconRefresh}
                    size="md"
                    aria-label="Sorry, we couldn’t fetch notifications!"
                />
            )
        }

        let codeHostMessage = this.state.isOpen
            ? undefined
            : this.state.messagesOrError === ExternalServiceNoActivityReasons.NoCodehosts
            ? 'No code host connections'
            : 'No repositories'
        if (isNoActivityReason(this.state.messagesOrError)) {
            return (
                <Icon
                    role="img"
                    data-tooltip={codeHostMessage}
                    as={CloudOffOutlineIcon}
                    size="md"
                    aria-hidden={this.state.isOpen}
                    aria-label={codeHostMessage}
                />
            )
        }

        if (
            this.state.messagesOrError.some(({ type }) => type === 'ExternalServiceSyncError' || type === 'SyncError')
        ) {
            codeHostMessage = this.state.isOpen ? undefined : 'Syncing repositories failed!'
            return (
                <Icon
                    role="img"
                    data-tooltip={codeHostMessage}
                    as={CloudAlertIconRefresh}
                    size="md"
                    aria-hidden={this.state.isOpen}
                    aria-label={codeHostMessage}
                />
            )
        }
        if (this.state.messagesOrError.some(({ type }) => type === 'CloningProgress')) {
            codeHostMessage = this.state.isOpen ? undefined : 'Cloning repositories...'
            return (
                <Icon
                    img="img"
                    data-tooltip={codeHostMessage}
                    as={CloudSyncIconRefresh}
                    size="md"
                    aria-hidden={this.state.isOpen}
                    aria-label={codeHostMessage}
                />
            )
        }
        codeHostMessage = this.state.isOpen ? undefined : 'Repositories up-to-date'
        return (
            <Icon
                role="img"
                data-tooltip={codeHostMessage}
                as={CloudCheckIconRefresh}
                size="md"
                aria-hidden={this.state.isOpen}
                aria-label={codeHostMessage}
            />
        )
    }

    private getOpenedNotificationsPayload(messagesOrError: MessageOrError): { status: string[] } {
        const messageTypes =
            typeof messagesOrError === 'string'
                ? [messagesOrError]
                : isErrorLike(messagesOrError)
                ? ['error']
                : messagesOrError.map(message => message.type)

        return { status: messageTypes.length === 0 ? ['success'] : messageTypes }
    }

    private trackUserNotificationsEvent(eventName: string): void {
        if (window.context.sourcegraphDotComMode && this.state.messagesOrError) {
            const payload = this.getOpenedNotificationsPayload(this.state.messagesOrError)
            eventLogger.log(`UserNotifications${upperFirst(eventName)}`, payload, payload)
        }
    }

    public render(): JSX.Element | null {
        const { isOpen } = this.state

        if (isOpen) {
            this.trackUserNotificationsEvent('opened')
        }

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
                            this.renderMessage(this.state.messagesOrError, this.props.user.isSiteAdmin)
                        )}
                    </div>
                </PopoverContent>
            </Popover>
        )
    }
}
