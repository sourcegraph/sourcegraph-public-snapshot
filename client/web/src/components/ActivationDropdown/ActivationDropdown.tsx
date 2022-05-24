import React from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { CircularProgressbar } from 'react-circular-progressbar'
import Confetti from 'react-dom-confetti'
import { concat, of, Subject, Subscription } from 'rxjs'
import { concatMap, delay, filter, map, pairwise, startWith, tap } from 'rxjs/operators'

import { Activation, percentageDone } from '@sourcegraph/shared/src/components/activation/Activation'
import { ActivationChecklist } from '@sourcegraph/shared/src/components/activation/ActivationChecklist'
import { Menu, MenuButton, MenuList, Typography } from '@sourcegraph/wildcard'

import styles from './ActivationDropdown.module.scss'

export interface ActivationDropdownProps {
    history: H.History
    activation: Activation
    /**
     * Forces display of the activation dropdown button. Used for Storybook testing.
     */
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
export class ActivationDropdown extends React.PureComponent<ActivationDropdownProps, State> {
    public state: State = { animate: false, displayEvenIfFullyCompleted: false }
    private componentUpdates = new Subject<ActivationDropdownProps>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(this.props),
                    map(props => props.activation.completed),
                    pairwise(),
                    filter(([previous, current]) => {
                        if (!previous || !current) {
                            return false
                        }
                        return percentageDone(current) > percentageDone(previous)
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

    public render(): JSX.Element | null {
        const show =
            this.props.alwaysShow ||
            this.state.displayEvenIfFullyCompleted ||
            this.state.animate ||
            (this.props.activation.completed !== undefined && percentageDone(this.props.activation.completed) < 100)
        if (!show) {
            return <div className="test-activation-hidden d-none" />
        }

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
                                'bg-transparent align-items-center test-activation-nav-item-toggle',
                                styles.activationDropdownButton,
                                { animate: this.state.animate }
                            )}
                        >
                            <div className={styles.confetti}>
                                <Confetti
                                    active={this.state.animate}
                                    config={{
                                        angle: 210,
                                        ...confettiConfig,
                                    }}
                                />
                            </div>
                            Get started
                            <div className={styles.confetti}>
                                <Confetti
                                    active={this.state.animate}
                                    config={{
                                        angle: 330,
                                        ...confettiConfig,
                                    }}
                                />
                            </div>
                            <span className={styles.progressBarContainer}>
                                <CircularProgressbar
                                    className={classNames('test-activation-progress-bar', styles.circularProgressBar)}
                                    strokeWidth={12}
                                    value={percentageDone(this.props.activation.completed)}
                                />
                            </span>
                        </MenuButton>
                        <MenuList
                            className={styles.activationDropdown}
                            data-testid="activation-dropdown"
                            isOpen={isOpen || this.props.alwaysShow}
                        >
                            <div className={classNames('dropdown-item-text', styles.activationDropdownHeader)}>
                                <Typography.H3 className="mb-0">
                                    {percentageDone(this.props.activation.completed) > 0
                                        ? 'Almost there!'
                                        : 'Welcome to Sourcegraph'}
                                </Typography.H3>
                                <p className="mb-2">Complete the steps below to finish onboarding!</p>
                            </div>
                            <ActivationChecklist
                                {...this.props}
                                steps={this.props.activation.steps}
                                completed={this.props.activation.completed}
                                className={styles.checklist}
                            />
                        </MenuList>
                    </>
                )}
            </Menu>
        )
    }
}
