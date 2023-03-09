import { FC, useState, useLayoutEffect } from 'react'
import classNames from 'classnames'
import ShieldHalfFullIcon from 'mdi-react/ShieldHalfFullIcon'
import { mdiChevronDoubleUp, mdiChevronDoubleDown } from '@mdi/js'

import { PageHeader, Button, Icon, useWindowSize, VIEWPORT_LG } from '@sourcegraph/wildcard'
import { useSentinelQuery } from './graphql/useSentinelQuery'
import styles from './SentinelView.module.scss'
import { SummaryTable } from './components/SummaryTable/SummaryTable'
import { SentinelBanner } from './components/SentinelBanner/SentinelBanner'
import { VulnerabilityList } from './components/VulnerabilityList/VulnerabilityList'
import { VulnerabilitySidebarView } from './components/VulnerabilitySidebar/VulnerabilitySidebar'

export const SentinelView: FC = () => {
    const [showMobileFilters, setShowMobileFilters] = useState(true)

    const { vulnerabilities, loading, error, refetch } = useSentinelQuery({ severity: '' })

    if (loading) {
        return <div>Loading...</div>
    }

    if (error) {
        return <div>"Ruh roh"</div>
    }

    return (
        <section className="w-100">
            <SentinelBanner />
            <div className={styles.pageContainer}>
                <PageHeader path={[{ icon: ShieldHalfFullIcon, text: 'Sentinel' }]} className={styles.header} />
                <FilterButton showMobileFilters={showMobileFilters} onShowMobileFiltersClicked={setShowMobileFilters} />
                <div className={classNames(styles.container, { [styles.full]: !showMobileFilters })}>
                    <div className={styles.main}>
                        <SummaryTable vulnerabilityMatches={vulnerabilities} />
                        <VulnerabilityList vulnerabilityMatches={vulnerabilities} />
                    </div>

                    <div className={styles.sidebar}>
                        {showMobileFilters && (
                            <VulnerabilitySidebarView
                                onShowMobileFiltersChanged={setShowMobileFilters}
                                onFilterChosen={refetch}
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
    }, [width])
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
