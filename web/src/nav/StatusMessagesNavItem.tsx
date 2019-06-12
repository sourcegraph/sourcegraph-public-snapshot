import CloudCheckIcon from 'mdi-react/CloudCheckIcon'
import CloudSyncIcon from 'mdi-react/CloudSyncIcon'
import React from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { Observable, Subscription, timer } from 'rxjs'
import { concatMap, map } from 'rxjs/operators'
import { Link } from '../../../shared/src/components/Link'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { queryGraphQL } from '../backend/graphql'

function fetchAllStatusMessages(): Observable<GQL.IStatusMessage[]> {
    return queryGraphQL(
        gql`
            query {
                statusMessages {
                    message
                    type
                }
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.statusMessages)
    )
}

interface Props {
    messages?: GQL.IStatusMessage[]
    isSiteAdmin?: boolean
}

interface State {
    messages: GQL.IStatusMessage[]
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

    private toggleIsOpen = () => this.setState(prevState => ({ isOpen: !prevState.isOpen }))
    public state: State = { isOpen: false, messages: [] }

    public componentDidMount(): void {
        if (this.props.messages) {
            this.setState({ messages: this.props.messages })
        }

        this.subscriptions.add(
            timer(0, REFRESH_INTERVAL_MS)
                .pipe(concatMap(fetchAllStatusMessages))
                .subscribe(messages => this.setState({ messages }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private renderMessage(message: GQL.IStatusMessage): JSX.Element | null {
        switch (message.type) {
            case GQL.StatusMessageType.CLONING:
                return (
                    <div key={message.message} className="status-messages-nav-item__entry">
                        <h4 className="status-messages-nav-item__entry-title">Repositories updating</h4>
                        <p className="status-messages-nav-item__entry-copy">{message.message}</p>
                        {this.props.isSiteAdmin && (
                            <p className="status-messages-nav-item__entry-link">
                                <Link to={'/site-admin/external-services'}>Configure external services</Link>
                            </p>
                        )}
                    </div>
                )
            default:
                return <DropdownItem>{message.message}</DropdownItem>
        }
    }

    public render(): JSX.Element | null {
        const hasMessages = this.state.messages.length > 0
        const cloning = this.state.messages.some(({ type }) => type === GQL.StatusMessageType.CLONING)
        return (
            <ButtonDropdown
                isOpen={this.state.isOpen}
                toggle={this.toggleIsOpen}
                className="nav-link py-0 status-messages-nav-item__nav-link"
            >
                <DropdownToggle caret={false} className="bg-transparent d-flex align-items-center" nav={true}>
                    {cloning ? (
                        <CloudSyncIcon
                            className="icon-inline"
                            {...(!this.state.isOpen && { 'data-tooltip': 'Updating repositories...' })}
                        />
                    ) : (
                        <CloudCheckIcon
                            className="icon-inline"
                            {...(!this.state.isOpen && { 'data-tooltip': 'Repositories up to date' })}
                        />
                    )}
                </DropdownToggle>

                <DropdownMenu right={true} className="status-messages-nav-item__dropdown-menu">
                    {hasMessages ? (
                        this.state.messages.map(this.renderMessage.bind(this))
                    ) : (
                        <div className="status-messages-nav-item__entry">
                            <h4 className="status-messages-nav-item__entry-title">Repositories up to date</h4>
                            <p className="status-messages-nav-item__entry-copy">
                                All repositories hosted on the configured external services are up to date.
                            </p>
                            {this.props.isSiteAdmin && (
                                <p className="status-messages-nav-item__entry-link">
                                    <Link to={'/site-admin/external-services'}>Configure external services</Link>
                                </p>
                            )}
                        </div>
                    )}
                </DropdownMenu>
            </ButtonDropdown>
        )
    }
}
