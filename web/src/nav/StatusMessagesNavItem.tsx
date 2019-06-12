import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import CheckboxMarkedCircleOutlineIcon from 'mdi-react/CheckboxMarkedCircleOutlineIcon'
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
}

interface State {
    messages: GQL.IStatusMessage[]
    isOpen: boolean
}

const REFRESH_INTERVAL_MS = 3000

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
                    <Link to="/site-admin/repositories?filter=cloning" className="dropdown-item">
                        {message.message}
                    </Link>
                )
            default:
                return <DropdownItem>{message.message}</DropdownItem>
        }
    }

    public render(): JSX.Element | null {
        const hasMessages = this.state.messages.length > 0
        const cloning = this.state.messages.some(({ type }) => type === GQL.StatusMessageType.CLONING)
        return (
            <ButtonDropdown isOpen={this.state.isOpen} toggle={this.toggleIsOpen} className="nav-link py-0">
                <DropdownToggle caret={false} className="bg-transparent d-flex align-items-center" nav={true}>
                    {cloning ? (
                        <LoadingSpinner className="icon-inline" data-tooltip="Updating repositories..." />
                    ) : (
                        <CheckboxMarkedCircleOutlineIcon className="icon-inline" />
                    )}
                </DropdownToggle>

                <DropdownMenu right={true} className="status-messages-nav-item__dropdown-menu">
                    {hasMessages ? (
                        this.state.messages.map(this.renderMessage.bind(this))
                    ) : (
                        <DropdownItem>All repositories up to date</DropdownItem>
                    )}
                </DropdownMenu>
            </ButtonDropdown>
        )
    }
}
