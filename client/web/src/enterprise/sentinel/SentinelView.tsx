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
                {((isDesktopView && !showMobileFilters) || !isDesktopView) && (
                    <FilterButton
                        showMobileFilters={showMobileFilters}
                        onShowMobileFiltersClicked={onShowMobileFiltersClicked}
                    />
                )}
                <div className={classNames(styles.container, { [styles.full]: !showMobileFilters })}>
                    <div className={styles.main}>
                        <SummaryTable vulnerabilityMatches={vulnerabilities} />
                        <VulnerabilityList vulnerabilityMatches={vulnerabilities} />
                    </div>

                    <div className={styles.sidebar}>
                        {showMobileFilters && (
                            <VulnerabilitySidebarView
                                onShowMobileFiltersChanged={onShowMobileFiltersClicked}
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
