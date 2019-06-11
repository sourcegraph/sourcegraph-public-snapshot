import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import CheckboxMarkedCircleOutlineIcon from 'mdi-react/CheckboxMarkedCircleOutlineIcon'
import React from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { Observable, Subject, Subscription } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
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
}

interface State {
    messages: GQL.IStatusMessage[]
    isOpen: boolean
}

const refreshIntervalMs = 3000

export class StatusMessagesNavItem extends React.PureComponent<Props, State> {
    private notificationUpdates = new Subject<void>()
    private subscriptions = new Subscription()

    private toggleIsOpen = () => this.setState(prevState => ({ isOpen: !prevState.isOpen }))
    public state: State = { isOpen: false, messages: [] }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.notificationUpdates.pipe(switchMap(() => fetchAllStatusMessages())).subscribe(messages => {
                this.setState({ messages })
                setTimeout(() => this.notificationUpdates.next(), refreshIntervalMs)
            })
        )
        this.notificationUpdates.next()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private renderMessage(message: GQL.IStatusMessage): JSX.Element | null {
        switch (message.type) {
            case GQL.StatusMessageType.CURRENTLYCLONING:
                return (
                    <Link to="/site-admin/repositories?filter=cloning" className="dropdown-item">
                        {message.message}
                    </Link>
                )
            default:
                return <DropdownItem>{message.message}</DropdownItem>
        }
    }

    public render(): JSX.Element | null {
        const types = this.state.messages.map(n => n.type)
        const hasMessages = this.state.messages.length > 0
        const currentlyCloning = types.some(t => t === GQL.StatusMessageType.CURRENTLYCLONING)
        return (
            <ButtonDropdown isOpen={this.state.isOpen} toggle={this.toggleIsOpen} className="nav-link py-0">
                <DropdownToggle caret={false} className="bg-transparent d-flex align-items-center" nav={true}>
                    {currentlyCloning ? (
                        <LoadingSpinner className="icon-inline" data-tooltip="Updating repositories..." />
                    ) : (
                        <CheckboxMarkedCircleOutlineIcon className="icon-inline" />
                    )}
                </DropdownToggle>

                <DropdownMenu right={true} className="status-messages-nav-item__dropdown-menu">
                    {hasMessages ? (
                        this.state.messages.map(m => this.renderMessage(m))
                    ) : (
                        <DropdownItem>All repositories up to date</DropdownItem>
                    )}
                </DropdownMenu>
            </ButtonDropdown>
        )
    }
}
