import classNames from 'classnames'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import styles from './CodeMonitoringGettingStarted.module.scss'

export const CodeMonitoringGettingStarted: React.FunctionComponent<{}> = () => (
    <div>
        <div className="d-flex flex-column mb-5">
            <h2>Get started with code monitoring</h2>
            <p className={classNames('text-muted mb-4', styles.subheading)}>
                Watch your code for changes and trigger actions to get notifications, send webhooks, and more.{' '}
                <a href="https://docs.sourcegraph.com/code_monitoring">Learn more.</a>
            </p>
            <Link to="/code-monitoring/new" className={classNames('btn btn-primary', styles.createButton)}>
                Create your first code monitor →
            </Link>
        </div>
        <div className={classNames('container', styles.startingPointsContainer)}>
            <h3 className="mb-3">Starting points for your first monitor</h3>
            <div className="row no-gutters mb-3">
                <div className={classNames('col-6', styles.startingPoint)}>
                    <div className="card">
                        <div className="card-body p-3">
                            <h3>Watch for AWS secrets in commits</h3>
                            <p className="text-muted">
                                Use a search query to watch for new search results, and choose how to receive
                                notifications in response.
                            </p>
                            <a
                                href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points#watch-for-potential-secrets"
                                className="btn btn-secondary"
                            >
                                View in docs →
                            </a>
                        </div>
                    </div>
                </div>
                <div className={classNames('col-6', styles.startingPoint)}>
                    <div className="card">
                        <div className="card-body p-3">
                            <h3>Watch for new consumers of deprecated methods</h3>
                            <p className="text-muted">
                                Keep an eye on commits with new consumers of deprecated methods to keep your code base
                                up-to-date.
                            </p>
                            <a
                                href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points#watch-for-consumers-of-deprecated-endpoints"
                                className="btn btn-secondary"
                            >
                                View in docs →
                            </a>
                        </div>
                    </div>
                </div>
            </div>
            <a className="link" href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points">
                Find more starting points in the docs
            </a>
        </div>
        <div className="container mt-5 px-0">
            <h3 className="mb-3">Learn more about code monitoring</h3>
            <div className="row">
                <div className="col-4">
                    <div>
                        <h4>Core concepts</h4>
                        <p className="text-muted">
                            Craft searches that will monitor your code and trigger actions.{' '}
                            <a
                                href="https://docs.sourcegraph.com/code_monitoring/explanations/core_concepts"
                                className="link"
                            >
                                Read the docs
                            </a>
                        </p>
                    </div>
                </div>
                <div className="col-4">
                    <div>
                        <h4>Starting points and ideas</h4>
                        <p className="text-muted">
                            Find specific examples of useful code monitors to keep on top of security and consistency
                            concerns.{' '}
                            <a
                                href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points"
                                className="link"
                            >
                                Explore starting points
                            </a>
                        </p>
                    </div>
                </div>
                <div className="col-4">
                    <div>
                        <h4>Questions and feedback</h4>
                        <p className="text-muted">
                            We want to hear your feedback.{' '}
                            <a href="mailto:feedback@sourcegraph.com" className="link">
                                Share your thoughts
                            </a>
                        </p>
                    </div>
                </div>
            </div>
        </div>
    </div>
)
