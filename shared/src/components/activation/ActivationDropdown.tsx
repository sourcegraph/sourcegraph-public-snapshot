import H from 'history'
import React from 'react'
import { CircularProgressbar } from 'react-circular-progressbar'
import Confetti from 'react-dom-confetti'
import { concat, of, Subject, Subscription } from 'rxjs'
import { concatMap, delay, filter, map, pairwise, startWith, tap } from 'rxjs/operators'
import { Activation, percentageDone } from './Activation'
import { ActivationChecklist } from './ActivationChecklist'
import { Menu, MenuButton, MenuPopover, MenuList } from '@reach/menu-button'
import classNames from 'classnames'

interface Props {
    history: H.History
    activation: Activation
    // For storybook testing
    alwaysShow?: boolean
}

interface State {
    displayEvenIfFullyCompleted: boolean
    animate: boolean
}

const animationDurationMillis = 3260
/**
 * Renders the activation status navlink item, a dropdown button that shows activation
 * status in the navbar.
 */
export class ActivationDropdown extends React.PureComponent<Props, State> {
    public state: State = { animate: false, displayEvenIfFullyCompleted: false }
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
            this.props.alwaysShow ||
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
            colors: ['#a864fd', '#29cdff', '#78ff44', '#ff718d', '#fdff6a'],
        }
        return (
            <Menu>
                {({ isOpen }) => (
                    <>
                        <MenuButton
                            className={classNames(
                                'activation-dropdown-button activation-dropdown-button__animated-button bg-transparent align-items-center e2e-activation-nav-item-toggle',
                                { animate: this.state.animate },
                                { 'activation-dropdown-button--hidden': !show }
                            )}
                        >
                            <div className="activation-dropdown-button__confetti">
                                <Confetti
                                    active={this.state.animate}
                                    config={{
                                        angle: 210,
                                        ...confettiConfig,
                                    }}
                                />
                            </div>
                            Get started
                            <div className="activation-dropdown-button__confetti">
                                <Confetti
                                    active={this.state.animate}
                                    config={{
                                        angle: 330,
                                        ...confettiConfig,
                                    }}
                                />
                            </div>
                            <span className="activation-dropdown-button__progress-bar-container">
                                <CircularProgressbar
                                    className="activation-dropdown-button__circular-progress-bar"
                                    strokeWidth={12}
                                    value={percentageDone(this.props.activation.completed)}
                                />
                            </span>
                        </MenuButton>
                        <MenuPopover>
                            <MenuList className={classNames('activation-dropdown', 'dropdown-menu', { show: isOpen })}>
                                <div className="dropdown-item-text activation-dropdown-header">
                                    <span className="mb-2">
                                        {percentageDone(this.props.activation.completed) > 0
                                            ? 'Almost there!'
                                            : 'Welcome to Sourcegraph'}
                                    </span>
                                    <p className="mb-2">Complete the steps below to finish onboarding!</p>
                                </div>
                                <ActivationChecklist
                                    {...this.props}
                                    steps={this.props.activation.steps}
                                    completed={this.props.activation.completed}
                                    className="activation-dropdown__checklist"
                                />
                            </MenuList>
                        </MenuPopover>
                    </>
                )}
            </Menu>
        )
    }
}
