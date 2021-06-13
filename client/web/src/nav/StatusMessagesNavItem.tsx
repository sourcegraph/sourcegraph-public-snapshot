import classNames from 'classnames'
import * as H from 'history'
import { isEqual } from 'lodash'
import { MdiReactIconProps } from 'mdi-react'
import AlertIcon from 'mdi-react/AlertIcon'
import CheckboxCircleIcon from 'mdi-react/CheckboxMarkedCircleIcon'
import CloudAlertIconCurrent from 'mdi-react/CloudAlertIcon'
import CloudCheckIconCurrent from 'mdi-react/CloudCheckIcon'
import CloudOffOutlineIcon from 'mdi-react/CloudOffOutlineIcon'
import CloudSyncIconCurrent from 'mdi-react/CloudSyncIcon'
import InformationCircleIcon from 'mdi-react/InformationCircleIcon'
import SyncIcon from 'mdi-react/SyncIcon'
import React from 'react'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'
import { Observable, Subscription, of } from 'rxjs'
import { catchError, map, repeatWhen, delay, distinctUntilChanged, switchMap } from 'rxjs/operators'

import {
    CloudAlertIconRefresh,
    CloudSyncIconRefresh,
    CloudCheckIconRefresh,
    IconProps,
} from '@sourcegraph/shared/src/components/icons'

import { Link } from '../../../shared/src/components/Link'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import { asError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { repeatUntil } from '../../../shared/src/util/rxjs/repeatUntil'
import { requestGraphQL } from '../backend/graphql'
import { ErrorAlert } from '../components/alerts'
import { CircleDashedIcon } from '../components/CircleDashedIcon'
import { queryExternalServices } from '../components/externalServices/backend'
import { StatusMessagesResult } from '../graphql-operations'

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
    isRedesignEnabled?: boolean
}

type Icon =
    | React.FunctionComponent<IconProps>
    | React.ComponentClass<MdiReactIconProps, any>
    | React.FunctionComponent<MdiReactIconProps>

interface IconsToShow {
    CloudAlertIcon: Icon
    CloudSyncIcon: Icon
    CloudCheckIcon: Icon
    SyncIcon: Icon
}

function iconsToShow(isRedesignEnabled = false): IconsToShow {
    const CloudAlertIcon = isRedesignEnabled ? CloudAlertIconRefresh : CloudAlertIconCurrent
    const CloudSyncIcon = isRedesignEnabled ? CloudSyncIconRefresh : CloudSyncIconCurrent
    const CloudCheckIcon = isRedesignEnabled ? CloudCheckIconRefresh : CloudCheckIconCurrent
    const SyncIconMdi = isRedesignEnabled ? CloudSyncIconRefresh : SyncIcon
    return { CloudAlertIcon, CloudSyncIcon, CloudCheckIcon, SyncIcon: SyncIconMdi }
}

function entryIcon(entryType: EntryType): JSX.Element {
    switch (entryType) {
        case 'error': {
            return <InformationCircleIcon size={14} className="text-danger status-messages-nav-item__entry-icon" />
        }
        case 'warning':
            return <AlertIcon size={14} className="text-warning status-messages-nav-item__entry-icon" />
        case 'success':
            return <CheckboxCircleIcon size={14} className="text-success status-messages-nav-item__entry-icon" />
        case 'progress':
            return <SyncIcon size={14} className="text-primary status-messages-nav-item__entry-icon" />
        case 'not-active':
            return (
                <CircleDashedIcon
                    size={16}
                    className="status-messages-nav-item__entry-icon status-messages-nav-item__entry-icon--off"
                />
            )
    }
}

const getMessageColor = (entryType: string): string => {
    const messageClass = 'status-messages-nav-item__entry-message'
    switch (entryType) {
        case 'error':
            return `${messageClass}--error`
        case 'warning':
            return `${messageClass}--warning`
    }

    return ''
}

const StatusMessagesNavItemEntry: React.FunctionComponent<StatusMessageEntryProps> = props => (
    <div key={props.message} className="status-messages-nav-item__entry">
        <h4 className="d-flex align-items-center mb-0">
            {entryIcon(props.entryType)}
            {props.title ? props.title : 'Your repositories'}
        </h4>
        {props.entryType === 'not-active' ? (
            <div className="status-messages-nav-item__entry-card status-messages-nav-item__entry-card--inactive border-0">
                <p className="text-muted status-messages-nav-item__entry-message">{props.message}</p>
                <Link className="text-primary" to={props.linkTo} onClick={props.linkOnClick}>
                    {props.linkText}
                </Link>
            </div>
        ) : (
            <div
                className={classNames(
                    'status-messages-nav-item__entry-card status-messages-nav-item__entry-card--active',
                    `status-messages-nav-item__entry--border-${props.entryType}`
                )}
            >
                <p className={classNames('status-messages-nav-item__entry-message', getMessageColor(props.entryType))}>
                    {props.message}
                </p>
                {props.messageHint && (
                    <>
                        <small className="text-muted d-inline-block mb-1">{props.messageHint}</small>
                        <br />
                    </>
                )}
                <Link className="text-primary" to={props.linkTo} onClick={props.linkOnClick}>
                    {props.linkText}
                </Link>
            </div>
        )}
        {props.progressHint && <small className="text-muted">{props.progressHint}</small>}
    </div>
)

interface User {
    id: string
    username: string
    isSiteAdmin: boolean
}

interface Props {
    user: User
    history: H.History
    fetchMessages?: () => Observable<StatusMessagesResult['statusMessages']>
    isRedesignEnabled?: boolean
}

enum ExternalServiceNoActivityReasons {
    NO_CODEHOSTS = 'NO_CODEHOSTS',
    NO_REPOS = 'NO_REPOS',
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

const REFRESH_INTERVAL_AFTER_ERROR_MS = 3000
const REFRESH_INTERVAL_MS = 10000

/**
 * Displays a status icon in the navbar reflecting the completion of backend
 * tasks such as repository cloning, and exposes a dropdown menu containing
 * more information on these tasks.
 */
export class StatusMessagesNavItem extends React.PureComponent<Props, State> {
    private subscriptions = new Subscription()

    public state: State = { isOpen: false, messagesOrError: [] }

    private toggleIsOpen = (): void => this.setState(previousState => ({ isOpen: !previousState.isOpen }))

    public componentDidMount(): void {
        this.subscriptions.add(
            queryExternalServices({
                namespace: this.props.user.id,
                first: null,
                after: null,
            })
                .pipe(
                    switchMap(({ nodes: services }) => {
                        if (services.length === 0) {
                            return of(ExternalServiceNoActivityReasons.NO_CODEHOSTS)
                        }

                        if (
                            !services.some(service => service.repoCount !== 0) &&
                            services.every(service => service.lastSyncError === null && service.warning === null)
                        ) {
                            return of(ExternalServiceNoActivityReasons.NO_REPOS)
                        }

                        return (this.props.fetchMessages ?? fetchAllStatusMessages)()
                    }),
                    catchError(error => [asError(error) as ErrorLike]),
                    // Poll on REFRESH_INTERVAL_MS, or REFRESH_INTERVAL_AFTER_ERROR_MS if there is an error.
                    repeatUntil(messagesOrError => isErrorLike(messagesOrError), { delay: REFRESH_INTERVAL_MS }),
                    repeatWhen(completions => completions.pipe(delay(REFRESH_INTERVAL_AFTER_ERROR_MS))),
                    distinctUntilChanged((a, b) => isEqual(a, b))
                )
                .subscribe(messagesOrError => {
                    this.setState({ messagesOrError })
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private renderMessage(message: Message, isSiteAdmin: boolean): JSX.Element | JSX.Element[] {
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
        if (Array.isArray(message) && message.length === 0) {
            return (
                <StatusMessagesNavItemEntry
                    key="up-to-date"
                    message="Repositories available for search"
                    linkTo={links.viewRepositories}
                    linkText="Manage repositories"
                    linkOnClick={this.toggleIsOpen}
                    entryType="success"
                    isRedesignEnabled={this.props.isRedesignEnabled}
                />
            )
        }

        // no code hosts or no repos
        if (isNoActivityReason(message)) {
            if (message === ExternalServiceNoActivityReasons.NO_REPOS) {
                return (
                    <StatusMessagesNavItemEntry
                        key={message}
                        message="Add repositories to start searching your code on Sourcegraph."
                        linkTo={links.manageRepositories}
                        linkText="Add repositories"
                        linkOnClick={this.toggleIsOpen}
                        entryType="not-active"
                        isRedesignEnabled={this.props.isRedesignEnabled}
                    />
                )
            }
            return (
                <StatusMessagesNavItemEntry
                    key={message}
                    message="Connect with a code host to start adding your code to Sourcegraph."
                    linkTo={links.manageCodeHosts}
                    linkText="Connect with code host"
                    linkOnClick={this.toggleIsOpen}
                    entryType="not-active"
                    isRedesignEnabled={this.props.isRedesignEnabled}
                />
            )
        }

        const cloningProgress = message.find(message_ => message_.type === 'CloningProgress')
        const indexing = message.find(message_ => message_.type === 'IndexingProgress')

        if (cloningProgress && indexing) {
            message = message.filter(message_ => message_.type !== 'IndexingProgress')
        }

        return message.map(message_ => {
            switch (message_.type) {
                case 'CloningProgress':
                    return (
                        <StatusMessagesNavItemEntry
                            key={message_.message}
                            message={message_.message}
                            messageHint="Your repositories may not be up to date."
                            linkTo={links.viewRepositories}
                            linkText="View status"
                            linkOnClick={this.toggleIsOpen}
                            entryType="progress"
                            progressHint={indexing && indexing.type === 'IndexingProgress' ? indexing.message : ''}
                            isRedesignEnabled={this.props.isRedesignEnabled}
                        />
                    )
                case 'IndexingProgress':
                    return (
                        <StatusMessagesNavItemEntry
                            key={message_.message}
                            message="Repositories available for search"
                            linkTo={links.viewRepositories}
                            linkText="Manage repositories"
                            linkOnClick={this.toggleIsOpen}
                            entryType="success"
                            progressHint={message_.message}
                            isRedesignEnabled={this.props.isRedesignEnabled}
                        />
                    )
                case 'ExternalServiceSyncError':
                    return (
                        <StatusMessagesNavItemEntry
                            key={message_.externalService.id}
                            message={`Can't connect to ${message_.externalService.displayName}`}
                            messageHint="Verify the code host configuration."
                            linkTo={links.getCodeHostLink(message_.externalService.id)}
                            linkText="Manage code hosts"
                            linkOnClick={this.toggleIsOpen}
                            entryType="error"
                            isRedesignEnabled={this.props.isRedesignEnabled}
                        />
                    )
                case 'SyncError':
                    return (
                        <StatusMessagesNavItemEntry
                            key={message_.message}
                            message={message_.message}
                            messageHint="Your repositories may not be up to date."
                            linkTo={links.viewRepositories}
                            linkText="Manage repositories"
                            linkOnClick={this.toggleIsOpen}
                            entryType="error"
                            isRedesignEnabled={this.props.isRedesignEnabled}
                        />
                    )
                case 'IndexingError':
                    return (
                        <StatusMessagesNavItemEntry
                            key={message_.message}
                            message={message_.message}
                            messageHint="Your repositories are up to date, but search speed may be slower than usual."
                            linkTo={links.viewRepositories}
                            linkText="Troubleshoot"
                            linkOnClick={this.toggleIsOpen}
                            entryType="warning"
                            isRedesignEnabled={this.props.isRedesignEnabled}
                        />
                    )
            }
        })
    }

    private renderIcon(): JSX.Element | null {
        const { CloudAlertIcon, SyncIcon, CloudCheckIcon } = iconsToShow(true)

        if (isErrorLike(this.state.messagesOrError)) {
            return <CloudAlertIcon className="icon-inline-md" />
        }

        if (isNoActivityReason(this.state.messagesOrError)) {
            return (
                <CloudOffOutlineIcon
                    className="icon-inline-md"
                    data-tooltip={
                        this.state.isOpen
                            ? undefined
                            : this.state.messagesOrError === 'NO_CODEHOSTS'
                            ? 'No code host connections'
                            : 'No repositories'
                    }
                />
            )
        }

        if (this.state.messagesOrError.some(({ type }) => type === 'ExternalServiceSyncError')) {
            return (
                <CloudAlertIcon
                    className="icon-inline-md"
                    data-tooltip={this.state.isOpen ? undefined : 'Syncing repositories failed!'}
                />
            )
        }
        if (this.state.messagesOrError.some(({ type }) => type === 'CloningProgress')) {
            return (
                <SyncIcon
                    className="icon-inline-md"
                    data-tooltip={this.state.isOpen ? undefined : 'Cloning repositories...'}
                />
            )
        }
        return (
            <CloudCheckIcon
                className="icon-inline-md"
                data-tooltip={this.state.isOpen ? undefined : 'Repositories up to date'}
            />
        )
    }

    public render(): JSX.Element | null {
        return (
            <ButtonDropdown
                isOpen={this.state.isOpen}
                toggle={this.toggleIsOpen}
                className="nav-link py-0 px-0 percy-hide chromatic-ignore"
            >
                <DropdownToggle caret={false} className="btn btn-link" nav={true}>
                    {this.renderIcon()}
                </DropdownToggle>

                <DropdownMenu right={true} className="status-messages-nav-item__dropdown-menu p-0">
                    <div className="status-messages-nav-item__dropdown-menu-content">
                        <small className="d-inline-block text-muted status-messages-nav-item__entry-sync">
                            Code sync status
                        </small>
                        {isErrorLike(this.state.messagesOrError) ? (
                            <ErrorAlert
                                className="status-messages-nav-item__entry"
                                prefix="Failed to load status messages"
                                error={this.state.messagesOrError}
                            />
                        ) : (
                            this.renderMessage(this.state.messagesOrError, this.props.user.isSiteAdmin)
                        )}
                    </div>
                </DropdownMenu>
            </ButtonDropdown>
        )
    }
}
