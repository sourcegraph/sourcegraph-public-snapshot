import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import CircularProgressbar from 'react-circular-progressbar'
import Confetti from 'react-dom-confetti'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { concat, of, Subject, Subscription } from 'rxjs'
import { concatMap, delay, filter, map, pairwise, startWith, tap } from 'rxjs/operators'
import { Activation, percentageDone } from './Activation'
import { ActivationChecklistItem } from './ActivationChecklist'

interface Props {
    history: H.History
    activation: Activation
}

interface State {
    displayEvenIfFullyCompleted: boolean
    isOpen: boolean
    animate: boolean
}

const animationDurationMillis = 3260

/**
 * Renders the activation status navlink item, a dropdown button that shows activation
 * status in the navbar.
 */
export class ActivationDropdown extends React.PureComponent<Props, State> {
    public state: State = { isOpen: false, animate: false, displayEvenIfFullyCompleted: false }
    private toggleIsOpen = () => this.setState(prevState => ({ isOpen: !prevState.isOpen }))
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(this.props),
                    map(props => props.activation.completed),
                    pairwise(),
                    filter(([prev, cur]) => {
                        if (!prev || !cur) {
                            return false
                        }
                        return percentageDone(cur) > percentageDone(prev)
                    }),
                    tap(didIncrease => {
                        if (didIncrease) {
                            this.setState({ displayEvenIfFullyCompleted: true })
                        }
                    }),
                    concatMap(() => concat(of(true), of(false).pipe(delay(animationDurationMillis))))
                )
                .subscribe(animate => this.setState({ animate }))
        )
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element {
        const show =
            this.state.displayEvenIfFullyCompleted ||
            this.state.animate ||
            (this.props.activation.completed !== undefined && percentageDone(this.props.activation.completed) < 100)
        const confettiConfig = {
            spread: 68,
            startVelocity: 12,
            elementCount: 81,
            dragFriction: 0.09,
            duration: animationDurationMillis,
            delay: 20,
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
                    } activation-dropdown-button__animated-button bg-transparent d-flex align-items-center e2e-activation-nav-item-toggle`}
                    nav={true}
                >
                    <Confetti
                        active={this.state.animate}
                        config={{
                            angle: 210,
                            ...confettiConfig,
                        }}
                    />
                    Setup
                    <Confetti
                        active={this.state.animate}
                        config={{
                            angle: 330,
                            ...confettiConfig,
                        }}
                    />
                    <CircularProgressbar
                        className="activation-dropdown-button__circular-progress-bar"
                        strokeWidth={12}
                        percentage={percentageDone(this.props.activation.completed)}
                    />
                </DropdownToggle>
                <DropdownMenu className="activation-dropdown" right={true}>
                    <div className="dropdown-item-text activation-dropdown-header">
                        <h3 className="mb-1">Get started with Sourcegraph</h3>
                        <p className="mb-0">
                            Welcome to Sourcegraph! Complete the steps below to finish setting up your instance.
                        </p>
                    </div>
                    <DropdownItem divider={true} />
                    {this.props.activation && this.props.activation.completed ? (
                        this.props.activation.steps.map(step => (
                            <div
                                key={step.id}
                                className="activation-dropdown-item dropdown-item"
                                onClick={this.toggleIsOpen}
                            >
                                <ActivationChecklistItem
                                    {...step}
                                    history={this.props.history}
                                    done={
                                        (this.props.activation.completed && this.props.activation.completed[step.id]) ||
                                        false
                                    }
                                />
                            </div>
                        ))
                    ) : (
                        <div className="activation-dropdown-button__loader">
                            <LoadingSpinner className="icon-inline" />
                        </div>
                    )}
                </DropdownMenu>
            </ButtonDropdown>
        )
    }
}
