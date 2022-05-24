import React from 'react'

import classNames from 'classnames'
import PlusIcon from 'mdi-react/PlusIcon'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Link, Button, CardBody, Card, Icon, Typography } from '@sourcegraph/wildcard'

import { CodeMonitorSignUpLink } from './CodeMonitoringSignUpLink'

import styles from './CodeMonitoringGettingStarted.module.scss'

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

export const CodeMonitoringGettingStarted: React.FunctionComponent<
    React.PropsWithChildren<CodeMonitoringGettingStartedProps>
> = ({ isLightTheme, isSignedIn }) => {
    const assetsRoot = window.context?.assetsRoot || ''

    return (
        <div>
            <Card className={classNames('mb-5 flex-column flex-lg-row', styles.hero)}>
                <img
                    src={`${assetsRoot}/img/codemonitoring-illustration-${isLightTheme ? 'light' : 'dark'}.svg`}
                    alt="A code monitor observes a depcreated library being used in code and sends an email alert."
                    className={classNames('mr-lg-5', styles.heroImage)}
                />
                <div className="align-self-center">
                    <Typography.H2 className={classNames('mb-3', styles.heading)}>
                        Proactively monitor changes to your codebase
                    </Typography.H2>
                    <p className={classNames('mb-4')}>
                        With code monitoring, you can automatically track changes made across multiple code hosts and
                        repositories.
                    </p>

                    <Typography.H3>Common use cases</Typography.H3>
                    <ul>
                        <li>Identify when bad patterns are committed </li>
                        <li>Identify use of deprecated libraries</li>
                    </ul>
                    {isSignedIn ? (
                        <Button to="/code-monitoring/new" className={styles.createButton} variant="primary" as={Link}>
                            <Icon className="mr-2" as={PlusIcon} />
                            Create a code monitor
                        </Button>
                    ) : (
                        <CodeMonitorSignUpLink
                            className={styles.createButton}
                            eventName="SignUpPLGMonitor_GettingStarted"
                            text="Get started with code monitors"
                        />
                    )}
                </div>
            </Card>
            <div>
                <Typography.H3 className="mb-3">Example code monitors</Typography.H3>

                <div className={classNames('mb-3', styles.startingPointsContainer)}>
                    {exampleCodeMonitors.map(monitor => (
                        <div className={styles.startingPoint} key={monitor.title}>
                            <Card className="h-100">
                                <CardBody className="d-flex flex-column">
                                    <Typography.H3>{monitor.title}</Typography.H3>
                                    <p className="text-muted flex-grow-1">{monitor.description}</p>
                                    <Link to={createCodeMonitorUrl(monitor)}>Create copy of monitor</Link>
                                </CardBody>
                            </Card>
                        </div>
                    ))}
                </div>
            </div>
            <div className="mt-5 px-0">
                <div className="row">
                    <div className="col-4">
                        <div>
                            <Typography.H4>Get started</Typography.H4>
                            <p className="text-muted">
                                Craft searches that will monitor your code and trigger actions such as email
                                notifications.
                            </p>
                            <Link to="/help/code_monitoring" className="link">
                                Code monitoring documentation
                            </Link>
                        </div>
                    </div>
                    <div className="col-4">
                        <div>
                            <Typography.H4>Starting points and ideas</Typography.H4>
                            <p className="text-muted">
                                Find specific examples of useful code monitors to keep on top of security and
                                consistency concerns.
                            </p>
                            <Link to="/help/code_monitoring/how-tos/starting_points" className="link">
                                Explore starting points
                            </Link>
                        </div>
                    </div>
                    {isSignedIn ? (
                        <div className="col-4">
                            <div>
                                <Typography.H4>Questions and feedback</Typography.H4>
                                <p className="text-muted">
                                    Have a question or idea about code monitoring? We want to hear your feedback!
                                </p>
                                <Link to="mailto:feedback@sourcegraph.com" className="link">
                                    Share your thoughts
                                </Link>
                            </div>
                        </div>
                    ) : (
                        <div className="col-4">
                            <Card className={styles.signUpCard}>
                                <Typography.H4>Free for registered users</Typography.H4>
                                <p className="text-muted">Sign up and build your first code monitor today.</p>
                                <CodeMonitorSignUpLink
                                    className={styles.createButton}
                                    eventName="SignUpPLGMonitor_GettingStarted"
                                    text="Get started"
                                />
                            </Card>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
