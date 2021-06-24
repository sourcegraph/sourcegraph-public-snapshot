import classNames from 'classnames'
import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import styles from './CodeMonitoringGettingStarted.module.scss'

export const CodeMonitoringGettingStarted: React.FunctionComponent<ThemeProps> = ({ isLightTheme }) => {
    const assetsRoot = window.context?.assetsRoot || ''

    return (
        <div>
            <div className={classNames('mb-5 card flex-lg-row', styles.hero)}>
                <img
                    src={`${assetsRoot}/img/codemonitoring-illustration-${isLightTheme ? 'light' : 'dark'}.svg`}
                    alt="A code monitor observes a bearer token being added to code and sends an email alert."
                    className={classNames('flex-shrink-0', styles.heroImage)}
                />
                <div>
                    <h2 className={classNames('mb-3', styles.heading)}>Proactively monitor changes to your codebase</h2>

                    <p className={classNames('mb-4')}>
                        With code monitoring, you can automatically track changes made across multiple code hosts and
                        repositories.
                    </p>

                    <h3>Common use cases</h3>
                    <ul>
                        <li>Watch for secrets in commits</li>
                        <li>Identify when bad patterns are commited </li>
                        <li>Identify use of depricated libraries</li>
                    </ul>
                    <Link to="/code-monitoring/new" className={classNames('btn btn-primary', styles.createButton)}>
                        <PlusIcon className="icon-inline mr-2" />
                        Create a code monitor
                    </Link>
                </div>
            </div>
            <div className={classNames('container', styles.startingPointsContainer)}>
                <h3 className="mb-3">Starting points for your first monitor</h3>
                <div className="row no-gutters code-monitoring-page__start-points-panel-container mb-3">
                    <div className={classNames('col-6', styles.startingPoint)}>
                        <div className="card">
                            <div className="card-body p-3 d-flex">
                                <img
                                    src={`${assetsRoot}/img/codemonitoring-search-${
                                        isLightTheme ? 'light' : 'dark'
                                    }.svg`}
                                    alt=""
                                    className="mr-3"
                                />
                                <div>
                                    <h3 className="mb-3">
                                        <a href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points#watch-for-potential-secrets">
                                            Watch for AWS secrets in commits
                                        </a>
                                    </h3>
                                    <p className="text-muted">
                                        Use a search query to watch for new search results, and choose how to receive
                                        notifications in response.
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div className={classNames('col-6', styles.startingPoint)}>
                        <div className="card">
                            <div className="card-body p-3 d-flex">
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
                </div>
            </div>
        </div>
    )
}
