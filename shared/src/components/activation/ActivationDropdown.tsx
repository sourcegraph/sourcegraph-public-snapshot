import H from 'history'
import RocketIcon from 'mdi-react/RocketIcon'
import React from 'react'
import CircularProgressbar from 'react-circular-progressbar'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { Subscription } from 'rxjs'
import { ActivationStatus, percentageDone } from './Activation'
import { ActivationChecklistItem } from './ActivationChecklist'

export interface Props {
    history: H.History
    activation: ActivationStatus
}

interface State {
    isOpen: boolean
    animate: boolean
    show?: boolean
    completed?: { [key: string]: boolean }
}

/**
 * Renders the activation status navlink item, a dropdown button that shows activation
 * status in the navbar.
 */
export class ActivationDropdown extends React.PureComponent<Props, State> {
    private subscriptions = new Subscription()
    public state: State = { isOpen: false, animate: false }

    private toggleIsOpen = () => this.setState(prevState => ({ isOpen: !prevState.isOpen }))

    public componentWillMount(): void {
        this.subscriptions.add(
            this.props.activation.completed.subscribe(completed => {
                if (!completed) {
                    // Completion is unloaded
                    this.setState({ show: false })
                    return
                }
                if (!this.state.completed && completed && percentageDone(completed) >= 100) {
                    // Completion was previously unloaded and now is 100% done
                    this.setState({ show: false })
                    return
                }
                if (
                    completed &&
                    this.state.completed &&
                    percentageDone(completed) > percentageDone(this.state.completed)
                ) {
                    // Completition was loaded and is being updated
                    this.setState({ completed, animate: true })
                    setTimeout(() => {
                        if (percentageDone(this.state.completed) >= 100) {
                            this.setState({ animate: false, show: false })
                        } else {
                            this.setState({ animate: false, show: true })
                        }
                    }, 1500)
                    return
                }

                // Completion is being updated (catch all)
                this.setState({ completed, show: true })
            })
        )
    }
    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }
    public render(): JSX.Element {
        return (
            <ButtonDropdown
                isOpen={this.state.isOpen}
                toggle={this.toggleIsOpen}
                className={`${this.state.show ? '' : 'activation-dropdown-button--hidden'} nav-link p-0`}
            >
                <DropdownToggle
                    caret={false}
                    className={`${
                        this.state.animate ? 'animate' : ''
                    } activation-dropdown-button__animated-button bg-transparent d-flex align-items-center e2e-user-nav-item-toggle`}
                    nav={true}
                >
                    Setup
                    <span className="activation-dropdown-button__progress-bar-container">
                        <CircularProgressbar
                            className="activation-dropdown-button__circular-progress-bar"
                            strokeWidth={12}
                            percentage={percentageDone(this.state.completed)}
                        />
                    </span>
                </DropdownToggle>
                <DropdownMenu right={true}>
                    <DropdownItem header={true} className="py-1">
                        <div className="activation-dropdown-header">
                            <h2>
                                <RocketIcon className="icon-inline" /> Hi there!
                            </h2>
                            <div>Complete the steps below to finish onboarding to Sourcegraph.</div>
                        </div>
                    </DropdownItem>
                    <DropdownItem divider={true} />
                    {this.state.completed ? (
                        this.props.activation.steps.map(s => (
                            <div key={s.id} className="dropdown-item" onClick={this.toggleIsOpen}>
                                <ActivationChecklistItem
                                    {...s}
                                    history={this.props.history}
                                    done={(this.state.completed && this.state.completed[s.id]) || false}
                                />
                            </div>
                        ))
                    ) : (
                        <div>Loading...</div>
                    )}
                </DropdownMenu>
            </ButtonDropdown>
        )
    }
}
