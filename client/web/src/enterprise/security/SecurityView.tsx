import { FC, useState } from 'react'
import ShieldHalfFullIcon from 'mdi-react/ShieldHalfFullIcon'
import { mdiEyeOff, mdiCheckCircle } from '@mdi/js'
import { useQuery } from '@sourcegraph/http-client'
import classNames from 'classnames'
import { mdiOpenInNew } from '@mdi/js'
import { VulnerabilitiesVariables, VulnerabilitiesResult, Scalers } from '../../graphql-operations'
import { PageHeader, Badge, Card, CardBody, Button, H2, H3, Icon, Text, Link } from '@sourcegraph/wildcard'
import { RESOLVE_SECURITY_VULNERABILITIES_QUERY } from './SecurityViewQueries'
import styles from './SecurityView.module.scss'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'

export interface SecurityViewProps {
    /**
     * Possible dashboard id. All insights on the page will be got from
     * dashboard's info from the user or org settings by the dashboard id.
     * In case if id is undefined we get insights from the final
     * version of merged settings (all insights)
     */
    dashboardId?: string
}

export const SecurityView: FC<SecurityViewProps> = props => {
    // const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
    let mockData = [
        {
            cve: 'CVE-2023-1234',
            description: 'Remote code exectuion vulnerability foo in bar.',
            dependency: 'vulnerable-package',
            packageManager: 'npm',
            publishedDate: '1st February 2023',
            lastUpdate: '9th February 2023',
            sourceFile: 'github.com/sourcegraph/sourcegraph:README.md',
            sourceFileLineNumber: 1,
            affectedVersion: '<1.2.3',
            currentVersion: '1.2.2',
            severityScore: '9.8',
            severityString: 'High',
        },
        {
            cve: 'CVE-2023-1234',
            description: 'Remote code exectuion vulnerability foo in bar.',
            dependency: 'vulnerable-package',
            packageManager: 'npm',
            publishedDate: '1st February 2023',
            lastUpdate: '9th February 2023',
            sourceFile: 'github.com/sourcegraph/sourcegraph:README.md',
            sourceFileLineNumber: 1,
            affectedVersion: '<1.2.3',
            currentVersion: '1.2.2',
            severityScore: '9.8',
            severityString: 'High',
        },
    ]

    const repository = 'UmVwb3NpdG9yeTozMQ==' as Scalars['ID']

    const { data, loading, error } = useQuery<VulnerabilitiesVariables, VulnerabilitiesResult>(
        RESOLVE_SECURITY_VULNERABILITIES_QUERY,
        {
            variables: {
                repository,
            },
            notifyOnNetworkStatusChange: false,
            fetchPolicy: 'no-cache',
        }
    )

    if (loading) {
        return <div>Loading...</div>
    }

    if (error) {
        return <div>"Ruh roh"</div>
    }

    console.log('I HAZ DATA', data)

    return (
        <div className={styles.pageContainer}>
            <PageHeader
                path={[{ icon: ShieldHalfFullIcon, text: 'Sentinel' }]}
                // actions={<CodeInsightHeaderActions dashboardId={absoluteDashboardId} telemetryService={telemetryService} />}
                className={styles.header}
            />
            <div className={styles.container}>
                {/* <Card as={CardBody} className={styles.heroSection}>
                <aside className={styles.heroVideoBlock}>
                    <img
                        src="https://a.storyblok.com/f/151984/1912x1264/3dd15d94d0/ssc-vuln1.gif"
                        alt="get started video"
                        className={styles.heroImage}
                    />
                </aside>
                <SecurityDescription className={styles.heroDescriptionBlock} />
            </Card> */}

                <div className={styles.contents}>
                    <div className={styles.bar}>
                        <div className={styles.barContainer}>
                            <div className={styles.barItem}>
                                <div className={styles.amount}>1</div>
                                <div className={styles.subtitle}>Total Vulnerabilities</div>
                            </div>
                            <div className={styles.barItem}>
                                <div className={styles.amount}>1/10</div>
                                <div className={styles.subtitle}>Critical Severity</div>
                            </div>
                            <div className={styles.barItem}>
                                <div className={styles.amount}>10/245</div>
                                <div className={styles.subtitle}>High Severity</div>
                            </div>
                            <div className={styles.barItem}>
                                <div className={styles.amount}>10/245</div>
                                <div className={styles.subtitle}>Medium Severity</div>
                            </div>
                            <div className={styles.barItem}>
                                <div className={styles.amount}>10/245</div>
                                <div className={styles.subtitle}>Repos with Vulnerabilities</div>
                            </div>
                        </div>
                    </div>
                    <div>
                        <ul className={styles.vulnerabilityList}>
                            {mockData.map((data, index) => {
                                return (
                                    <li className={styles.listItem}>
                                        <SecurityCard data={data} />
                                    </li>
                                )
                            })}{' '}
                        </ul>
                    </div>
                </div>

                <div className={styles.sidebar}>
                    <SecurityFilter />
                </div>
            </div>
        </div>
    )
}

const SecurityFilter: React.FunctionComponent<Props> = ({ className }) => (
    <aside className={styles.filter}>
        <div className={styles.filterContainer}>
            <p>Filter</p>
        </div>
    </aside>
)

interface Props {
    className?: string
    data?: any
}

const SecurityCard: React.FunctionComponent<Props> = ({
    className,
    data: {
        cve,
        description,
        dependency,
        packageManager,
        publishedDate,
        lastUpdate,
        sourceFile,
        sourceFileLineNumber,
        currentVersion,
        severityScore,
        severityString,
    },
}) => {
    var result = description.match(/[^\.!\?]+[\.!\?]+/g)
    console.log('dis results', result)

    return (
        <div className={styles.vulnerabilityCard}>
            <div className={styles.vulnerabilityContainer}>
                <div className={styles.vulnerabilityHeader}>
                    <div className={styles.vulnerabilityTitle}>
                        <div>
                            {/* <Badge variant="warning">5.3 Medium</Badge> */}
                            <Badge variant="warning">
                                {severityScore} {severityString}
                            </Badge>
                            <span className={styles.cve}>{cve}</span>
                        </div>
                    </div>
                    <div className={styles.vulnerabilityButtonContainer}>
                        <Button variant="secondary">
                            <Icon aria-hidden={true} svgPath={mdiEyeOff} />
                            Ignore
                        </Button>
                        <Button variant="secondary">
                            <Icon aria-hidden={true} svgPath={mdiCheckCircle} />
                            Mark as Resolved
                        </Button>
                    </div>
                </div>
                <div className={styles.vulnerabilityDescription}>
                    <div className={styles.vulnerabilityDescriptionContainer}>
                        {/* <H3>{title}</H3> */}
                        <p>{description}</p>
                    </div>
                    <div className={styles.vulnerabilityVersions}>
                        <div>
                            <div className={styles.versionTitle}>Affected Version</div>
                            <div className={classNames(styles.versionNumber, styles.red)}>Unknown</div>
                        </div>
                        <div>
                            <div className={styles.versionTitle}>Patch Version</div>
                            <div className={classNames(styles.versionNumber, styles.green)}>{currentVersion}</div>
                        </div>
                    </div>
                    <div className={styles.descriptionTable}>
                        <div>
                            <div className={styles.descriptionSubheader}>Dependency</div>
                            <div className={styles.descriptionValue}>{dependency}</div>
                        </div>
                        <div>
                            <div className={styles.descriptionSubheader}>Package Manager</div>
                            <div className={styles.descriptionValue}>{packageManager}</div>
                        </div>
                        <div>
                            <div className={styles.descriptionSubheader}>Published Date</div>
                            <div className={styles.descriptionValue}>{publishedDate}</div>
                        </div>
                        <div>
                            <div className={styles.descriptionSubheader}>Last Updated</div>
                            <div className={styles.descriptionValue}>{lastUpdate}</div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

export const SecurityDescription: React.FunctionComponent<Props> = ({ className }) => (
    <section className={className}>
        <H2>Fix your vulnerabilities, now</H2>

        <Text>
            Security provides precise answers about the vulnerabilities currently present in your codebase. It
            transforms the way you see your codebase and lets you fix your vulnerabilities in seconds.
        </Text>

        <div>
            <H3>Use Security to...</H3>

            <ul>
                <li>Fix vulnerabilities and deprecations</li>
                <li>Detect versions of packages, or infrastructure that are affected</li>
                <li>Ensure removal of security vulnerabilities</li>
                <li>Track code smells, ownership, and configurations</li>
                <li>
                    <Link to="#" rel="noopener">
                        See more use cases
                    </Link>
                </li>
            </ul>
        </div>

        <H3>Resources</H3>
        <ul>
            <li>
                <Link to="#" target="_blank" rel="noopener">
                    Documentation <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                </Link>
            </li>
            <li>
                <Link to="https://about.sourcegraph.com/code-insights" target="_blank" rel="noopener">
                    Product page <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                </Link>
            </li>
        </ul>
    </section>
)
