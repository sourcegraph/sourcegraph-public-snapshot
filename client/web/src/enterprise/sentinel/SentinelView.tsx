import { FC, useState, useLayoutEffect } from 'react'
import classNames from 'classnames'
import ShieldHalfFullIcon from 'mdi-react/ShieldHalfFullIcon'
import { mdiChevronDoubleUp, mdiChevronDoubleDown } from '@mdi/js'
// import { useQuery } from '@sourcegraph/http-client'
// import { VulnerabilitiesVariables, VulnerabilitiesResult, Scalers } from '../../graphql-operations'
import { PageHeader, Button, Icon, useWindowSize, VIEWPORT_LG } from '@sourcegraph/wildcard'
// import { RESOLVE_SECURITY_VULNERABILITIES_QUERY } from './SecurityViewQueries'
import styles from './SentinelView.module.scss'
// import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { SummaryTable } from './components/SummaryTable/SummaryTable'
import { SentinelBanner } from './components/SentinelBanner/SentinelBanner'
import { VulnerabilityList } from './components/VulnerabilityList/VulnerabilityList'
import { VulnerabilitySidebarView } from './components/VulnerabilitySidebar/VulnerabilitySidebar'

export const SentinelView: FC = () => {
    // const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
    let mockData = [
        {
            cve: 'CVE-2023-1234', // VulnerabilityMatch.Vulnerability.sourceID
            description: 'Remote code exectuion vulnerability foo in bar.', // VulnerabilityMatch.Vulnerability.details
            dependency: 'vulnerable-package', // VulnerabilityMatch.VulnerabilityAffectedPackage.packageName
            packageManager: 'npm', // VulnerabilityMatch.VulnerabilityAffectedPackage.language
            publishedDate: '1st February 2023', // VulnerabilityMatch.Vulnerability.published
            lastUpdate: '9th February 2023', // VulnerabilityMatch.Vulnerability.modified
            sourceFile: 'github.com/sourcegraph/sourcegraph:README.md', // Not avail yet
            sourceFileLineNumber: 1, // Not needed. Will come from code intel location
            affectedVersion: '<1.2.3', // VulnerabilityMatch.VulnerabilityAffectedPackage.versionConstraint (sort and take first one)
            currentVersion: '1.2.2', // Not avail yet
            severityScore: '9.8', // VulnerabilityMatch.Vulnerability.cvssScore
            severityString: 'High', // VulnerabilityMatch.Vulnerability.severity
            vulnerableCode: [
                // ^ VulnerabilityMatch.location (not avail yet)
                {
                    repository: 'github.com/sourcegraph/sourcegraph',
                    fileName: 'browser/src/libs/code_intelligence/code_intelligence.tsx',
                    group: [
                        {
                            repoName: 'github.com/sourcegraph/sourcegraph',
                            path: 'browser/src/libs/code_intelligence/code_intelligence.tsx',
                            locations: [
                                {
                                    repo: 'github.com/sourcegraph/sourcegraph',
                                    file: 'browser/src/libs/code_intelligence/code_intelligence.tsx',
                                    content: 'const foo = "bar"',
                                    commitID: '1234567890',
                                    range: {
                                        start: {
                                            line: 1,
                                            character: 1,
                                        },
                                        end: {
                                            line: 10,
                                            character: 5,
                                        },
                                    },
                                    url: 'https://google.com',
                                    lines: ['foo', 'bar', 'baz'],
                                    precise: true,
                                },
                            ],
                        },
                    ],
                },
                {
                    repository: 'github.com/sourcegraph/logger',
                    fileName: 'web/src/components/CodeSnippet.tsx',
                    group: [
                        {
                            repoName: 'github.com/sourcegraph/logger',
                            path: 'web/src/components/CodeSnippet.tsx',
                            locations: [
                                {
                                    repo: 'github.com/sourcegraph/sourcegraph',
                                    file: 'browser/src/libs/code_intelligence/code_intelligence.tsx',
                                    content: 'const foo = "bar"',
                                    commitID: '1234567890',
                                    range: {
                                        start: {
                                            line: 1,
                                            character: 1,
                                        },
                                        end: {
                                            line: 10,
                                            character: 5,
                                        },
                                    },
                                    url: 'https://google.com',
                                    lines: ['foo', 'bar', 'baz'],
                                    precise: true,
                                },
                            ],
                        },
                    ],
                },
            ],
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
            vulnerableCode: [
                {
                    repository: 'github.com/sourcegraph/sourcegraph',
                    fileName: 'browser/src/libs/code_intelligence/code_intelligence.tsx',
                    group: [
                        {
                            repoName: 'github.com/sourcegraph/sourcegraph',
                            path: 'browser/src/libs/code_intelligence/code_intelligence.tsx',
                            locations: [
                                {
                                    repo: 'github.com/sourcegraph/sourcegraph',
                                    file: 'browser/src/libs/code_intelligence/code_intelligence.tsx',
                                    content: 'const foo = "bar"',
                                    commitID: '1234567890',
                                    range: {
                                        start: {
                                            line: 1,
                                            character: 1,
                                        },
                                        end: {
                                            line: 10,
                                            character: 5,
                                        },
                                    },
                                    url: 'https://google.com',
                                    lines: ['foo', 'bar', 'baz'],
                                    precise: true,
                                },
                            ],
                        },
                    ],
                },
                {
                    repository: 'github.com/sourcegraph/logger',
                    fileName: 'web/src/components/CodeSnippet.tsx',
                    group: [
                        {
                            repoName: 'github.com/sourcegraph/logger',
                            path: 'web/src/components/CodeSnippet.tsx',
                            locations: [
                                {
                                    repo: 'github.com/sourcegraph/sourcegraph',
                                    file: 'browser/src/libs/code_intelligence/code_intelligence.tsx',
                                    content: 'const foo = "bar"',
                                    commitID: '1234567890',
                                    range: {
                                        start: {
                                            line: 1,
                                            character: 1,
                                        },
                                        end: {
                                            line: 10,
                                            character: 5,
                                        },
                                    },
                                    url: 'https://google.com',
                                    lines: ['foo', 'bar', 'baz'],
                                    precise: true,
                                },
                            ],
                        },
                    ],
                },
            ],
        },
    ]

    const [showMobileFilters, setShowMobileFilters] = useState(true)
    const [isDesktopView, setIsDesktopView] = useState(true)
    const onShowMobileFiltersClicked = (): void => {
        const newShowFilters = !showMobileFilters
        setShowMobileFilters(newShowFilters)
    }
    const { width } = useWindowSize()
    useLayoutEffect(() => {
        if (width > VIEWPORT_LG) {
            setShowMobileFilters(true)
            setIsDesktopView(true)
        } else {
            setIsDesktopView(false)
        }
    }, [width])

    // const repository = 'UmVwb3NpdG9yeTozMQ==' as Scalars['ID']

    // const { data, loading, error } = useQuery<VulnerabilitiesVariables, VulnerabilitiesResult>(
    //     RESOLVE_SECURITY_VULNERABILITIES_QUERY,
    //     {
    //         variables: {
    //             repository,
    //         },
    //         notifyOnNetworkStatusChange: false,
    //         fetchPolicy: 'no-cache',
    //     }
    // )

    // if (loading) {
    //     return <div>Loading...</div>
    // }

    // if (error) {
    //     return <div>"Ruh roh"</div>
    // }

    // console.log('I HAZ DATA', data)

    return (
        <section className="w-100">
            <SentinelBanner />
            <div className={styles.pageContainer}>
                <PageHeader path={[{ icon: ShieldHalfFullIcon, text: 'Sentinel' }]} className={styles.header} />
                {((isDesktopView && !showMobileFilters) || !isDesktopView) && (
                    <FilterButton
                        showMobileFilters={showMobileFilters}
                        onShowMobileFiltersClicked={onShowMobileFiltersClicked}
                    />
                )}
                <div className={classNames(styles.container, { [styles.full]: !showMobileFilters })}>
                    <div className={styles.main}>
                        <SummaryTable />
                        <VulnerabilityList vulnerabilities={mockData} />
                    </div>

                    <div className={styles.sidebar}>
                        {showMobileFilters && (
                            <VulnerabilitySidebarView onShowMobileFiltersChanged={onShowMobileFiltersClicked} />
                        )}
                    </div>
                </div>
            </div>
        </section>
    )
}

interface FilterButtonProps {
    showMobileFilters: boolean
    onShowMobileFiltersClicked: () => void
}
const FilterButton: FC<FilterButtonProps> = ({ showMobileFilters, onShowMobileFiltersClicked }) => (
    <div className={styles.filterContainer}>
        <Button
            className={classNames(styles.filtersButton, showMobileFilters && 'active')}
            aria-pressed={showMobileFilters}
            onClick={onShowMobileFiltersClicked}
            outline={true}
            variant="secondary"
            size="sm"
            aria-label={`${showMobileFilters ? 'Hide' : 'Show'} filters`}
        >
            Filters
            <Icon
                aria-hidden={true}
                className="ml-2"
                svgPath={showMobileFilters ? mdiChevronDoubleUp : mdiChevronDoubleDown}
            />
        </Button>
    </div>
)
