import H from 'history'
import RocketIcon from 'mdi-react/RocketIcon'
import React from 'react'
import CircularProgressbar from 'react-circular-progressbar'
import Confetti from 'react-dom-confetti'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { Activation, percentageDone } from './Activation'
import { ActivationChecklistItem } from './ActivationChecklist'

interface Props {
    history: H.History
    activation: Activation
}

interface State {
    isOpen: boolean
    animate: boolean
}

/**
 * Renders the activation status navlink item, a dropdown button that shows activation
 * status in the navbar.
 */
export class ActivationDropdown extends React.PureComponent<Props, State> {
    public state: State = { isOpen: false, animate: false }

    private toggleIsOpen = () => this.setState(prevState => ({ isOpen: !prevState.isOpen }))

    /**
     * Tracks the last remove-animation timeout that was added, which should be cleared when
     * the component is unmounted.
     */
    private removeAnimationTimeout?: NodeJS.Timeout

    public componentDidUpdate(prevProps: Readonly<Props>, prevState: Readonly<State>): void {
        const oldDone = percentageDone(prevProps.activation.completed)
        const newDone = percentageDone(this.props.activation.completed)
        if (newDone > oldDone && !this.state.animate) {
            this.setState({ animate: true })
            this.removeAnimationTimeout = setTimeout(() => {
                this.setState({ animate: false })
                this.removeAnimationTimeout = undefined
            }, 1500)
        }
    }

    public componentWillUnmount(): void {
        if (this.removeAnimationTimeout) {
            clearTimeout(this.removeAnimationTimeout)
        }
    }

    public render(): JSX.Element {
        const show = this.state.animate || percentageDone(this.props.activation.completed) < 100
        const confettiConfig = {
            spread: '68',
            startVelocity: '23',
            elementCount: '81',
            dragFriction: '0.09',
            duration: '5700',
            delay: '20',
            width: '10px',
            height: '10px',
            colors: ['#a864fd', '#29cdff', '#78ff44', '#ff718d', '#fdff6a'],
        }
        return (
            <ButtonDropdown
                isOpen={this.state.isOpen}
                toggle={this.toggleIsOpen}
                className={`${show ? '' : 'activation-dropdown-button--hidden'} nav-link p-0`}
            >
                <DropdownToggle
                    caret={false}
                    className={`${
                        this.state.animate ? 'animate' : ''
                    } activation-dropdown-button__animated-button bg-transparent d-flex align-items-center e2e-user-nav-item-toggle`}
                    nav={true}
                >
                    <Confetti
                        active={this.state.animate}
                        config={
                            {
                                angle: '210',
                                ...confettiConfig,
                            } as any
                        }
                    />
                    Setup
                    <Confetti
                        active={this.state.animate}
                        config={
                            {
                                angle: '330',
                                ...confettiConfig,
                            } as any
                        }
                    />
                    <span className="activation-dropdown-button__progress-bar-container">
                        <CircularProgressbar
                            className="activation-dropdown-button__circular-progress-bar"
                            strokeWidth={12}
                            percentage={percentageDone(this.props.activation.completed)}
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
                    {this.props.activation && this.props.activation.completed ? (
                        this.props.activation.steps.map(s => (
                            <div key={s.id} className="dropdown-item" onClick={this.toggleIsOpen}>
                                <ActivationChecklistItem
                                    {...s}
                                    history={this.props.history}
                                    done={
                                        (this.props.activation.completed && this.props.activation.completed[s.id]) ||
                                        false
                                    }
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
