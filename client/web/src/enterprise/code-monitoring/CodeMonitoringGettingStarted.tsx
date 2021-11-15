import classNames from 'classnames'
import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { CodeMonitorInfo } from './CodeMonitorInfo'
import styles from './CodeMonitoringGettingStarted.module.scss'
import { CodeMonitorSignUpLink } from './CodeMonitoringSignUpLink'

interface CodeMonitoringGettingStartedProps extends ThemeProps {
    isSignedIn: boolean
}

export const CodeMonitoringGettingStarted: React.FunctionComponent<CodeMonitoringGettingStartedProps> = ({
    isLightTheme,
    isSignedIn,
}) => {
    const assetsRoot = window.context?.assetsRoot || ''

    return (
        <div>
            <div className={classNames('mb-5 card flex-lg-row align-items-start', styles.hero)}>
                <div className="w-100">
                    <h2 className={classNames('mb-3', styles.heading)}>Proactively monitor changes to your codebase</h2>
                    <p className={classNames('mb-4')}>
                        With code monitoring, you can automatically track changes made across multiple code hosts and
                        repositories.
                    </p>

                    <h3>Common use cases</h3>
                    <ul>
                        <li>Identify when bad patterns are committed </li>
                        <li>Identify use of deprecated libraries</li>
                    </ul>
                    {isSignedIn ? (
                        <Link to="/code-monitoring/new" className={classNames('btn btn-primary', styles.createButton)}>
                            <PlusIcon className="icon-inline mr-2" />
                            Create a code monitor
                        </Link>
                    ) : (
                        <CodeMonitorSignUpLink
                            className={styles.createButton}
                            eventName="SignUpPLGMonitor_GettingStarted"
                            text="Sign up to create a code monitor"
                        />
                    )}
                </div>
                <CodeMonitorInfo className={classNames('mt-3 mt-lg-0 ml-lg-3 mb-0', styles.alert)} />
            </div>
            <div className={classNames('container', styles.startingPointsContainer)}>
                <h3 className="mb-3">Starting points for your first monitor</h3>
                <div className="row no-gutters code-monitoring-page__start-points-panel-container mb-3">
                    <div className={classNames('col-6', styles.startingPoint)}>
                        <div className="card h-100">
                            <div className="card-body p-3 d-flex flex-column flex-md-row">
                                <img
                                    className="mr-3 mt-3 mb-3 pt-1 pb-1"
                                    src={`${assetsRoot}/img/codemonitoring-search-${
                                        isLightTheme ? 'light' : 'dark'
                                    }.svg`}
                                    alt=""
                                />
                                <div className="flex">
                                    <h3 className="mb-3">
                                        <a href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points#get-notified-when-a-file-changes">
                                            Get notified when a file changes
                                        </a>
                                    </h3>
                                    <p className="text-muted">Use a search query to watch for changes to a file.</p>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div className={classNames('col-6', styles.startingPoint)}>
                        <div className="card h-100">
                            <div className="card-body p-3 d-flex flex-column flex-md-row">
                                <img
                                    src={`${assetsRoot}/img/codemonitoring-notify-${
                                        isLightTheme ? 'light' : 'dark'
                                    }.svg`}
                                    alt=""
                                    className="mr-3"
                                />
                                <div>
                                    <h3 className="mb-3">
                                        <a href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points#watch-for-consumers-of-deprecated-endpoints">
                                            Watch for new uses of deprecated methods
                                        </a>
                                    </h3>
                                    <p className="text-muted">
                                        Keep an eye on commits with new consumers of deprecated methods to keep your
                                        codebase up-to-date.
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            <div className="container mt-5 px-0">
                <div className="row">
                    <div className="col-4">
                        <div>
                            <h4>Get started</h4>
                            <p className="text-muted">
                                Craft searches that will monitor your code and trigger actions such as email
                                notifications.
                            </p>
                            <a href="https://docs.sourcegraph.com/code_monitoring" className="link">
                                Code monitoring documentation
                            </a>
                        </div>
                    </div>
                    <div className="col-4">
                        <div>
                            <h4>Starting points and ideas</h4>
                            <p className="text-muted">
                                Find specific examples of useful code monitors to keep on top of security and
                                consistency concerns.
                            </p>
                            <a
                                href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points"
                                className="link"
                            >
                                Explore starting points
                            </a>
                        </div>
                    </div>
                    {isSignedIn ? (
                        <div className="col-4">
                            <div>
                                <h4>Questions and feedback</h4>
                                <p className="text-muted">
                                    Have a question or idea about code monitoring? We want to hear your feedback!
                                </p>
                                <a href="mailto:feedback@sourcegraph.com" className="link">
                                    Share your thoughts
                                </a>
                            </div>
                        </div>
                    ) : (
                        <div className="col-4">
                            <div className={classNames('card', styles.signUpCard)}>
                                <h4>Free for registered users</h4>
                                <p className="text-muted">Sign up and build your first code monitor today.</p>
                                <CodeMonitorSignUpLink
                                    className={styles.createButton}
                                    eventName="SignUpPLGMonitor_GettingStarted"
                                    text="Sign up now"
                                />
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
