import { startCase } from 'lodash'
import CloudAlertIcon from 'mdi-react/CloudAlertIcon'
import CloudCheckIcon from 'mdi-react/CloudCheckIcon'
import CloudSyncIcon from 'mdi-react/CloudSyncIcon'
import React from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { Observable, SchedulerLike, Subscription, timer } from 'rxjs'
import { catchError, concatMap, map } from 'rxjs/operators'
import { Link } from '../../../shared/src/components/Link'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { queryGraphQL } from '../backend/graphql'

export function fetchAllStatusMessages(): Observable<GQL.IStatusMessage[]> {
    return queryGraphQL(
        gql`
            query StatusMessages {
                statusMessages {
                    message
                    type
                    metadata {
                        name
                        value
                    }
                }
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.statusMessages)
    )
}

interface StatusMessageEntryProps {
    title: string
    text: string
    showLink?: boolean
    linkTo: string
    linkText: string
    linkOnClick: (event: React.MouseEvent<HTMLAnchorElement, MouseEvent>) => void
}

const StatusMessagesNavItemEntry: React.FunctionComponent<StatusMessageEntryProps> = props => (
    <div key={props.text} className="status-messages-nav-item__entry">
        <h4>{props.title}</h4>
        <p>{props.text}</p>
        {props.showLink && (
            <p className="status-messages-nav-item__entry-link">
                <Link to={props.linkTo} onClick={props.linkOnClick}>
                    {props.linkText}
                </Link>
            </p>
        )}
    </div>
)

interface Props {
    fetchMessages: () => Observable<GQL.IStatusMessage[]>

    /** Scheduler for the refresh timer */
    scheduler?: SchedulerLike

    isSiteAdmin?: boolean
}

interface State {
    messagesOrError: GQL.IStatusMessage[] | ErrorLike
    isOpen: boolean
}

const REFRESH_INTERVAL_MS = 3000

/**
 * Displays a status icon in the navbar reflecting the completion of backend
 * tasks such as repository cloning, and exposes a dropdown menu containing
 * more information on these tasks.
 */
export class StatusMessagesNavItem extends React.PureComponent<Props, State> {
    private subscriptions = new Subscription()

    public state: State = { isOpen: false, messagesOrError: [] }

    private toggleIsOpen = () => this.setState(prevState => ({ isOpen: !prevState.isOpen }))

    public componentDidMount(): void {
        this.subscriptions.add(
            timer(0, REFRESH_INTERVAL_MS, this.props.scheduler)
                .pipe(concatMap(() => this.props.fetchMessages().pipe(catchError(err => [asError(err)]))))
                .subscribe(messagesOrError => this.setState({ messagesOrError }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private renderMessage(message: GQL.IStatusMessage): JSX.Element | null {
        switch (message.type) {
            case GQL.StatusMessageType.CLONING:
                return (
                    <StatusMessagesNavItemEntry
                        key={message.message}
                        title="Repositories cloning"
                        text={message.message}
                        showLink={this.props.isSiteAdmin}
                        linkTo="/site-admin/external-services"
                        linkText="Configure external services"
                        linkOnClick={this.toggleIsOpen}
                    />
                )
            case GQL.StatusMessageType.SYNCERROR:
                const displayName = message.metadata.find(m => m.name === 'ext_svc_name')
                const extSvcID = message.metadata.find(m => m.name === 'ext_svc_id')
                if (displayName !== undefined && extSvcID !== undefined) {
                    return (
                        <StatusMessagesNavItemEntry
                            key={message.message}
                            title={`Syncing external service "${displayName.value}" failed:`}
                            text={message.message}
                            showLink={this.props.isSiteAdmin}
                            linkTo={`/site-admin/external-services/${extSvcID.value}`}
                            linkText={`Edit "${displayName.value}"...`}
                            linkOnClick={this.toggleIsOpen}
                        />
                    )
                }
                return (
                    <StatusMessagesNavItemEntry
                        key={message.message}
                        title={`Syncing repositories failed`}
                        text={message.message}
                        showLink={this.props.isSiteAdmin}
                        linkTo="/site-admin/external-services"
                        linkText="Configure external services"
                        linkOnClick={this.toggleIsOpen}
                    />
                )
        }
    }

    private renderIcon(): JSX.Element | null {
        if (isErrorLike(this.state.messagesOrError)) {
            return <CloudAlertIcon className="icon-inline" />
        }
        if (this.state.messagesOrError.some(({ type }) => type === GQL.StatusMessageType.SYNCERROR)) {
            return (
                <CloudAlertIcon
                    className="icon-inline"
                    data-tooltip={this.state.isOpen ? undefined : 'Syncing repositories failed!'}
                />
            )
        }
        if (this.state.messagesOrError.some(({ type }) => type === GQL.StatusMessageType.CLONING)) {
            return (
                <CloudSyncIcon
                    className="icon-inline"
                    data-tooltip={this.state.isOpen ? undefined : 'Cloning repositories...'}
                />
            )
        }
        return (
            <CloudCheckIcon
                className="icon-inline"
                data-tooltip={this.state.isOpen ? undefined : 'Repositories up to date'}
            />
        )
    }

    public render(): JSX.Element | null {
        return (
            <ButtonDropdown
                isOpen={this.state.isOpen}
                toggle={this.toggleIsOpen}
                className="nav-link py-0 px-0 status-messages-nav-item__nav-link"
            >
                <DropdownToggle caret={false} className="btn btn-icon" nav={true}>
                    {this.renderIcon()}
                </DropdownToggle>

                <DropdownMenu right={true} className="status-messages-nav-item__dropdown-menu">
                    {isErrorLike(this.state.messagesOrError) ? (
                        <div className="status-messages-nav-item__entry">
                            <h4>Failed to load status messages:</h4>
                            <p className="alert alert-danger">{startCase(this.state.messagesOrError.message)}</p>
                        </div>
                    ) : this.state.messagesOrError.length > 0 ? (
                        this.state.messagesOrError.map((m, i) => {
                            if (!isErrorLike(this.state.messagesOrError) && i < this.state.messagesOrError.length - 1) {
                                return (
                                    <div>
                                        {this.renderMessage(m)}
                                        <DropdownItem divider={true}></DropdownItem>
                                    </div>
                                )
                            }
                            return this.renderMessage(m)
                        })
                    ) : (
                        <StatusMessagesNavItemEntry
                            title="Repositories up to date"
                            text="All repositories hosted on the configured external services are cloned."
                            showLink={this.props.isSiteAdmin}
                            linkTo="/site-admin/external-services"
                            linkText="Configure external services"
                            linkOnClick={this.toggleIsOpen}
                        />
                    )}
                </DropdownMenu>
            </ButtonDropdown>
        )
    }
}
