import H from 'history'
import RocketIcon from 'mdi-react/RocketIcon'
import React from 'react'
import CircularProgressbar from 'react-circular-progressbar'
import ButtonDropdown from 'reactstrap/lib/ButtonDropdown'
import DropdownItem from 'reactstrap/lib/DropdownItem'
import DropdownMenu from 'reactstrap/lib/DropdownMenu'
import DropdownToggle from 'reactstrap/lib/DropdownToggle'
import { Subscription } from 'rxjs'
import { Props as CommandListProps } from '../../commandPalette/CommandList'
import { PopoverButton } from '../PopoverButton'
import { ActivationStatus, percentageDone } from './Activation'
import { ActivationChecklist, ActivationChecklistItem, ActivationChecklistProps } from './ActivationChecklist'

export interface ActivationDropdownProps {
    history: H.History
    activation: ActivationStatus
}

interface State {
    isOpen: boolean

    animate: boolean
    show?: boolean
    completed?: { [key: string]: boolean }
}

export class ActivationDropdown extends React.PureComponent<ActivationDropdownProps, State> {
    public state: State = { isOpen: false, animate: false }
    private subscriptions = new Subscription()

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
        this.props.activation.ensureInit()
    }
    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }
    public render(): JSX.Element {
        return (
            <ButtonDropdown
                isOpen={this.state.isOpen}
                toggle={this.toggleIsOpen}
                className={`${this.state.show ? '' : 'activation-status-dropdown-button--hidden'} nav-link p-0`}
            >
                <DropdownToggle
                    caret={false}
                    className={`${
                        this.state.animate ? 'animate' : ''
                    } first-use-button bg-transparent d-flex align-items-center e2e-user-nav-item-toggle`}
                    nav={true}
                >
                    Setup
                    <span className="activation-status-dropdown-button__progress-bar-container">
                        <CircularProgressbar strokeWidth={12} percentage={percentageDone(this.state.completed)} />
                    </span>
                </DropdownToggle>
                <DropdownMenu right={true}>
                    <DropdownItem header={true} className="py-1">
                        <div className="activation-status-dropdown-header">
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

// interface ActivationPopoverDropdownProps extends ActivationChecklistProps {
//     onClick: () => void
// }

// /**
//  * Presents the site admin activation checklist in a navbar dropdown.
//  */
// export class ActivationPopoverDropdown extends React.PureComponent<ActivationPopoverDropdownProps> {
//     public render(): JSX.Element {
//         return (
//             <div
//                 className="activation-container command-list list-group list-group-flush rounded"
//                 onClick={this.props.onClick}
//             >
//                 <div className="list-group-item">
//                     <div className="activation-container__header">
//                         <h2>
//                             <RocketIcon className="icon-inline" /> Hi there!
//                         </h2>
//                         <div>Complete the steps below to finish onboarding to Sourcegraph.</div>
//                     </div>
//                 </div>
//                 <div className="list-group-item">{<ActivationChecklist {...this.props} />}</div>
//             </div>
//         )
//     }
// }

// export interface ActivationPopoverButtonProps extends CommandListProps {
//     history: H.History
//     activation: ActivationStatus
// }
// export interface ActivationPopoverButtonState {
//     animate: boolean
//     show?: boolean
//     completed?: { [key: string]: boolean }

//     dropdownHideOnChange?: any // controls dismissing the dropdown
// }

// /**
//  * The nav bar button that displays the current user activation status (percentage of
//  * activation steps completed and a dropdown list of all activation items and their
//  * completion status).
//  */
// export class ActivationPopoverButton extends React.PureComponent<
//     ActivationPopoverButtonProps,
//     ActivationPopoverButtonState
// > {
//     private subscriptions = new Subscription()
//     constructor(props: ActivationPopoverButtonProps) {
//         super(props)
//         this.state = { animate: false }
//     }
//     public componentWillMount(): void {
//         this.subscriptions.add(
//             this.props.activation.completed.subscribe(completed => {
//                 if (!completed) {
//                     // Completion is unloaded
//                     this.setState({ show: false })
//                     return
//                 }
//                 if (!this.state.completed && completed && percentageDone(completed) >= 100) {
//                     // Completion was previously unloaded and now is 100% done
//                     this.setState({ show: false })
//                     return
//                 }
//                 if (
//                     completed &&
//                     this.state.completed &&
//                     percentageDone(completed) > percentageDone(this.state.completed)
//                 ) {
//                     // Completition was loaded and is being updated
//                     this.setState({ completed, animate: true })
//                     setTimeout(() => {
//                         if (percentageDone(this.state.completed) >= 100) {
//                             this.setState({ animate: false, show: false })
//                         } else {
//                             this.setState({ animate: false, show: true })
//                         }
//                     }, 2000)
//                     return
//                 }

//                 // Completion is being updated (catch all)
//                 this.setState({ completed, show: true })
//             })
//         )
//         this.props.activation.ensureInit()
//     }
//     public componentWillUnmount(): void {
//         this.subscriptions.unsubscribe()
//     }
//     private onClick = (e: React.MouseEvent<HTMLElement, MouseEvent>) => {
//         e.preventDefault()
//     }
//     private onClickDropdown = () => {
//         this.setState({ dropdownHideOnChange: !this.state.dropdownHideOnChange })
//     }
//     public render(): JSX.Element | null {
//         return (
//             <div className={`activation-status-dropdown-button-container ${this.state.show ? '' : 'hidden'}`}>
//                 <PopoverButton
//                     className={`first-use-button activation-status-dropdown-button ${
//                         this.state.animate ? 'animate' : ''
//                     }`}
//                     {...this.state}
//                     popoverClassName="rounded"
//                     placement="auto-end"
//                     hideOnChange={this.state.dropdownHideOnChange}
//                     hideCaret={true}
//                     popoverElement={
//                         <ActivationPopoverDropdown
//                             steps={this.props.activation.steps}
//                             history={this.props.history}
//                             onClick={this.onClickDropdown}
//                             {...this.state}
//                         />
//                     }
//                 >
//                     <a className="nav-link activation-status-dropdown-button__nav-link" href="#" onClick={this.onClick}>
//                         Setup
//                         <span className="activation-status-dropdown-button__progress-bar-container">
//                             <CircularProgressbar strokeWidth={12} percentage={percentageDone(this.state.completed)} />
//                         </span>
//                     </a>
//                 </PopoverButton>
//             </div>
//         )
//     }
// }
