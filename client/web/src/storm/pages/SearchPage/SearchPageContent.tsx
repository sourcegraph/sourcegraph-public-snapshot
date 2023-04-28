import { FC, useEffect, useState } from 'react'

import { mdiArrowRight } from '@mdi/js'
import classNames from 'classnames'

import { QueryExamples } from '@sourcegraph/branded/src/search-ui/components/QueryExamples'
import { QueryState } from '@sourcegraph/shared/src/search'
import { getGlobalSearchContextFilter } from '@sourcegraph/shared/src/search/query/query'
import { appendContextFilter, omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Icon, Link, Tooltip, Text, Badge } from '@sourcegraph/wildcard'

import { CodyIcon } from '../../../cody/CodyIcon'
import { BrandLogo } from '../../../components/branding/BrandLogo'
import { useFeatureFlag } from '../../../featureFlags/useFeatureFlag'
import { useLegacyContext_onlyInStormRoutes } from '../../../LegacyRouteContext'
import { useExperimentalQueryInput } from '../../../search/useExperimentalSearchInput'

import { AddCodeHostWidget } from './AddCodeHostWidget'
import { SearchPageFooter } from './SearchPageFooter'
import { SearchPageInput } from './SearchPageInput'

import styles from './SearchPageContent.module.scss'

interface SearchPageContentProps {
    shouldShowAddCodeHostWidget?: boolean
}

export const SearchPageContent: FC<SearchPageContentProps> = props => {
    const { shouldShowAddCodeHostWidget } = props

    const { telemetryService, selectedSearchContextSpec, isSourcegraphDotCom, authenticatedUser, ownEnabled } =
        useLegacyContext_onlyInStormRoutes()

    const isLightTheme = useIsLightTheme()
    const [experimentalQueryInput] = useExperimentalQueryInput()
    const [ownFeatureFlagEnabled] = useFeatureFlag('search-ownership')
    const enableOwnershipSearch = ownEnabled && ownFeatureFlagEnabled

    /** The value entered by the user in the query input */
    const [queryState, setQueryState] = useState<QueryState>({
        query: '',
    })

    useEffect(() => telemetryService.logViewEvent('Home'), [telemetryService])
    useEffect(() => {
        // TODO (#48103): Remove/simplify when new search input is released
        // Because the current and the new search input handle the context: selector differently
        // we need properly "translate" the queries when switching between the both versions
        if (selectedSearchContextSpec) {
            setQueryState(state => {
                if (experimentalQueryInput) {
                    return { query: appendContextFilter(state.query, selectedSearchContextSpec) }
                }
                const contextFilter = getGlobalSearchContextFilter(state.query)?.filter
                if (contextFilter) {
                    return { query: omitFilter(state.query, contextFilter) }
                }
                return state
            })
        }
    }, [experimentalQueryInput, selectedSearchContextSpec])

    return (
        <div className={classNames('d-flex flex-column align-items-center px-3', styles.searchPage)}>
            <BrandLogo className={styles.logo} isLightTheme={isLightTheme} variant="logo" />
            {isSourcegraphDotCom && (
                <div className="text-muted mt-3 mr-sm-2 pr-2 text-center">
                    Searching millions of public repositories
                </div>
            )}

            <div className={styles.searchContainer}>
                {shouldShowAddCodeHostWidget ? (
                    <>
                        <Tooltip
                            content="Sourcegraph is not fully functional until a code host is set up"
                            placement="top"
                        >
                            <div className={styles.translucent}>
                                <SearchPageInput queryState={queryState} setQueryState={setQueryState} />
                            </div>
                        </Tooltip>
                        <AddCodeHostWidget className="mb-4" />
                    </>
                ) : (
                    <>
                        <SearchPageInput queryState={queryState} setQueryState={setQueryState} />
                        {window.context.codyEnabled && (
                            <div className="d-flex justify-content-center mt-4">
                                <Text className="text-muted">
                                    <Badge variant="merged">Experimental</Badge>{' '}
                                    <Link
                                        to="/search/cody"
                                        onClick={() =>
                                            telemetryService.log('ClickedOnTryCodySearchCTA', {
                                                location: 'SearchPage',
                                            })
                                        }
                                    >
                                        Ask Cody to construct a query from natural language <CodyIcon />{' '}
                                        <Icon svgPath={mdiArrowRight} aria-hidden={true} />
                                    </Link>
                                </Text>
                            </div>
                        )}
                    </>
                )}
            </div>
            <div className={classNames(styles.panelsContainer)}>
                {(!!authenticatedUser || isSourcegraphDotCom) && (
                    <QueryExamples
                        selectedSearchContextSpec={selectedSearchContextSpec}
                        telemetryService={telemetryService}
                        queryState={queryState}
                        setQueryState={setQueryState}
                        isSourcegraphDotCom={isSourcegraphDotCom}
                        enableOwnershipSearch={enableOwnershipSearch}
                    />
                )}
            </div>

            <SearchPageFooter />
        </div>
    )
}

interface ShouldShowAddCodeHostWidgetOptions {
    isAddCodeHostWidgetEnabled?: boolean
    isSiteAdmin?: boolean
    externalServicesCount?: number
}

export function getShouldShowAddCodeHostWidget({
    isAddCodeHostWidgetEnabled,
    isSiteAdmin,
    externalServicesCount,
}: ShouldShowAddCodeHostWidgetOptions): boolean {
    return !!isAddCodeHostWidgetEnabled && !!isSiteAdmin && externalServicesCount === 0
}
