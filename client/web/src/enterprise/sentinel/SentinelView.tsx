import { type FC, useState, useLayoutEffect } from 'react'

import { mdiChevronDoubleUp, mdiChevronDoubleDown } from '@mdi/js'
import classNames from 'classnames'
import ShieldHalfFullIcon from 'mdi-react/ShieldHalfFullIcon'

import { PageHeader, Button, Icon, useWindowSize, VIEWPORT_LG } from '@sourcegraph/wildcard'

import { ConnectionError, ConnectionLoading } from '../../components/FilteredConnection/ui'

import { SentinelBanner } from './components/SentinelBanner/SentinelBanner'
import { SummaryTable } from './components/SummaryTable/SummaryTable'
import { VulnerabilityList } from './components/VulnerabilityList/VulnerabilityList'
import { VulnerabilitySidebarView } from './components/VulnerabilitySidebar/VulnerabilitySidebar'
import { useSentinelQuery } from './graphql/useSentinelQuery'

import styles from './SentinelView.module.scss'

export const SentinelView: FC = () => {
    const [showMobileFilters, setShowMobileFilters] = useState(true)
    const [filter, setFilter] = useState({
        severity: '',
        language: '',
        repositoryName: '',
    })
    const { loading, hasNextPage, fetchMore, connection, error } = useSentinelQuery(filter)

    return (
        <section className="w-100">
            <SentinelBanner />
            <div className={styles.pageContainer}>
                <PageHeader path={[{ icon: ShieldHalfFullIcon, text: 'Sentinel' }]} className={styles.header} />
                <FilterButton showMobileFilters={showMobileFilters} onShowMobileFiltersClicked={setShowMobileFilters} />
                <div className={classNames(styles.container, { [styles.full]: !showMobileFilters })}>
                    <div className={styles.main}>
                        {error && <ConnectionError errors={[error.message]} />}
                        {loading && !connection && <ConnectionLoading />}
                        {connection && (
                            <>
                                <SummaryTable />
                                <VulnerabilityList vulnerabilityMatches={connection?.nodes ?? []} />
                                {hasNextPage && (
                                    <Button size="sm" variant="secondary" outline={true} onClick={fetchMore}>
                                        Load more vulnerabilities
                                    </Button>
                                )}
                            </>
                        )}
                    </div>
                    <div className={styles.sidebar}>
                        {showMobileFilters && (
                            <VulnerabilitySidebarView
                                onShowMobileFiltersChanged={setShowMobileFilters}
                                onFilterChosen={setFilter}
                            />
                        )}
                    </div>
                </div>
            </div>
        </section>
    )
}

interface FilterButtonProps {
    showMobileFilters: boolean
    onShowMobileFiltersClicked: React.Dispatch<React.SetStateAction<boolean>>
}
const FilterButton: FC<FilterButtonProps> = ({ showMobileFilters, onShowMobileFiltersClicked }) => {
    const [isDesktopView, setIsDesktopView] = useState(true)

    const { width } = useWindowSize()
    useLayoutEffect(() => {
        if (width > VIEWPORT_LG) {
            onShowMobileFiltersClicked(true)
            setIsDesktopView(true)
        } else {
            setIsDesktopView(false)
        }
    }, [width, onShowMobileFiltersClicked])
    const toggleMobileFiltersClicked = (): void => {
        const newShowFilters = !showMobileFilters
        onShowMobileFiltersClicked(newShowFilters)
    }

    return (isDesktopView && !showMobileFilters) || !isDesktopView ? (
        <div className={styles.filterContainer}>
            <Button
                className={classNames(styles.filtersButton, showMobileFilters && 'active')}
                aria-pressed={showMobileFilters}
                onClick={toggleMobileFiltersClicked}
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
    ) : null
}
