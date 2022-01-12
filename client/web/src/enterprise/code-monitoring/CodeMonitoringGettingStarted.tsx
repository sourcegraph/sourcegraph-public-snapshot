import classNames from 'classnames'
import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button } from '@sourcegraph/wildcard'

import styles from './CodeMonitoringGettingStarted.module.scss'
import { CodeMonitorSignUpLink } from './CodeMonitoringSignUpLink'

interface CodeMonitoringGettingStartedProps extends ThemeProps {
    isSignedIn: boolean
}

interface ExampleCodeMonitor {
    title: string
    description: string
    monitorName: string
    monitorQuery: string
}

const exampleCodeMonitors: ExampleCodeMonitor[] = [
    {
        title: 'Uses of a deprecated method',
        description:
            'Get notified when a deprecated method is added or removed. This example uses leftPad in JavaScript files.',
        monitorName: 'Uses of leftPad in JavaScript',
        monitorQuery: 'lang:JavaScript require(("|\')left-pad("|\')) patternType:regexp type:diff ',
    },
    {
        title: 'New library usage',
        description:
            'After you add a new library, you can watch your codebase for its usage and get notified when it’s imported or certain functions from it are used. This example uses faker in TypeScript.',
        monitorName: 'New uses of faker in TypeScript',
        monitorQuery:
            'lang:TypeScript import faker from "faker" OR import faker from \'faker\' type:diff select:commit.diff.added ',
    },
    {
        title: 'Bad coding patterns',
        description:
            'Get notified when someone uses a pattern that your team is trying to avoid. This example uses React class components in JavaScript.',
        monitorName: 'New React class components in JavaScript',
        monitorQuery:
            'lang:JavaScript class \\w extends React.Component type:diff patternType:regexp select:commit.diff.added ',
    },
    {
        title: 'IP address range',
        description:
            'Detect the usage of banned or invalid IP addresses in your code. This example uses local IP address in the 192.168.1.x range.',
        monitorName: 'New uses of local IP addresses',
        monitorQuery: '^192\\.168\\.1\\.([1-9]|[1-9]d|100)$ type:diff select:commit.diff.added patternType:regexp ',
    },
]

const createCodeMonitorUrl = (example: ExampleCodeMonitor): string => {
    const searchParameters = new URLSearchParams()
    searchParameters.set('trigger-query', example.monitorQuery)
    searchParameters.set('description', example.monitorName)
    return `/code-monitoring/new?${searchParameters.toString()}`
}

export const CodeMonitoringGettingStarted: React.FunctionComponent<CodeMonitoringGettingStartedProps> = ({
    isLightTheme,
    isSignedIn,
}) => {
    const assetsRoot = window.context?.assetsRoot || ''

    return (
        <div>
            <div className={classNames('mb-5 card flex-column flex-lg-row', styles.hero)}>
                <img
                    src={`${assetsRoot}/img/codemonitoring-illustration-${isLightTheme ? 'light' : 'dark'}.svg`}
                    alt="A code monitor observes a depcreated library being used in code and sends an email alert."
                    className={classNames('mr-lg-5', styles.heroImage)}
                />
                <div className="align-self-center">
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
                        <Button to="/code-monitoring/new" className={styles.createButton} variant="primary" as={Link}>
                            <PlusIcon className="icon-inline mr-2" />
                            Create a code monitor
                        </Button>
                    ) : (
                        <CodeMonitorSignUpLink
                            className={styles.createButton}
                            eventName="SignUpPLGMonitor_GettingStarted"
                            text="Sign up to create a code monitor"
                        />
                    )}
                </div>
            </div>
            <div>
                <h3 className="mb-3">Example code monitors</h3>

                <div className={classNames('mb-3', styles.startingPointsContainer)}>
                    {exampleCodeMonitors.map(monitor => (
                        <div className={styles.startingPoint} key={monitor.title}>
                            <div className="card h-100">
                                <div className="card-body p-3 d-flex flex-column">
                                    <h3>{monitor.title}</h3>
                                    <p className="text-muted flex-grow-1">{monitor.description}</p>
                                    <Link to={createCodeMonitorUrl(monitor)}>Create copy of monitor</Link>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
            <div className="mt-5 px-0">
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
