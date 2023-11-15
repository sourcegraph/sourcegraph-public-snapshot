import { type FC, useEffect, useState } from 'react'

import classNames from 'classnames'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { QueryExamples } from '@sourcegraph/branded/src/search-ui/components/QueryExamples'
import type { QueryState } from '@sourcegraph/shared/src/search'
import { getGlobalSearchContextFilter } from '@sourcegraph/shared/src/search/query/query'
import { appendContextFilter, omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Label, Tooltip, useLocalStorage } from '@sourcegraph/wildcard'

import { BrandLogo } from '../../../components/branding/BrandLogo'
import { useFeatureFlag } from '../../../featureFlags/useFeatureFlag'
import { useLegacyContext_onlyInStormRoutes } from '../../../LegacyRouteContext'
import { useV2QueryInput } from '../../../search/useV2QueryInput'
import { GettingStartedTour } from '../../../tour/GettingStartedTour'
import { useShowOnboardingTour } from '../../../tour/hooks'

import { AddCodeHostWidget } from './AddCodeHostWidget'
import { SearchPageFooter } from './SearchPageFooter'
import { SearchPageInput } from './SearchPageInput'
import { TryCodyCtaSection } from './TryCodyCtaSection'
import { TryCodySignUpCtaSection } from './TryCodySignUpCtaSection'

import styles from './SearchPageContent.module.scss'

interface SearchPageContentProps extends TelemetryV2Props {
    shouldShowAddCodeHostWidget?: boolean
}

export const SearchPageContent: FC<SearchPageContentProps> = props => {
    const { shouldShowAddCodeHostWidget } = props

    const { telemetryService, telemetryRecorder, selectedSearchContextSpec, isSourcegraphDotCom, authenticatedUser } =
        useLegacyContext_onlyInStormRoutes()

    const isLightTheme = useIsLightTheme()
    const [v2QueryInput] = useV2QueryInput()

    /** The value entered by the user in the query input */
    const [queryState, setQueryState] = useState<QueryState>({
        query: '',
    })

    useEffect(() => {
        telemetryService.logViewEvent('Home')
        telemetryRecorder.recordEvent('Home', 'viewed')
    }, [telemetryService, telemetryRecorder])
    useEffect(() => {
        // TODO (#48103): Remove/simplify when new search input is released
        // Because the current and the new search input handle the context: selector differently
        // we need properly "translate" the queries when switching between the both versions
        if (selectedSearchContextSpec) {
            setQueryState(state => {
                if (v2QueryInput) {
                    return { query: appendContextFilter(state.query, selectedSearchContextSpec) }
                }
                const contextFilter = getGlobalSearchContextFilter(state.query)?.filter
                if (contextFilter) {
                    return { query: omitFilter(state.query, contextFilter) }
                }
                return state
            })
        }
    }, [v2QueryInput, selectedSearchContextSpec])

    const defaultSimpleSearchToggle = true
    const [simpleSearch, setSimpleSearch] = useLocalStorage('simple.search.toggle', defaultSimpleSearchToggle)
    const [simpleSearchEnabled] = useFeatureFlag('enable-simple-search', false)

    const showOnboardingTour = useShowOnboardingTour({ authenticatedUser, isSourcegraphDotCom })
    const showCodyCTA = !showOnboardingTour

    return (
        <div className={classNames('d-flex flex-column align-items-center px-3', styles.searchPage)}>
            <BrandLogo className={styles.logo} isLightTheme={isLightTheme} variant="logo" />
            {isSourcegraphDotCom && (
                <div className="text-muted mt-3 mr-sm-2 pr-2 text-center">
                    Code search and an AI assistant with the context of the code graph.
                </div>
            )}

            <div className={styles.searchContainer}>
                {simpleSearchEnabled && (
                    <div className="mb-2">
                        <Label htmlFor="simpleSearchToggle" className="mr-2">
                            Simple search
                        </Label>
                        <Toggle
                            id="simpleSearchToggle"
                            value={simpleSearch}
                            onToggle={val => {
                                const arg = { state: val }
                                telemetryService.log('SimpleSearchToggle', arg, arg)
                                telemetryRecorder.recordEvent('SimpleSearch', 'toggled', { privateMetadata: { arg } })
                                setSimpleSearch(val)
                            }}
                        />
                    </div>
                )}

                {shouldShowAddCodeHostWidget ? (
                    <>
                        <Tooltip
                            content="Sourcegraph is not fully functional until a code host is set up"
                            placement="top"
                        >
                            <div className={styles.translucent}>
                                <SearchPageInput
                                    simpleSearch={false}
                                    queryState={queryState}
                                    setQueryState={setQueryState}
                                />
                            </div>
                        </Tooltip>
                        <AddCodeHostWidget className="mb-4" />
                    </>
                ) : (
                    <>
                        <SearchPageInput
                            simpleSearch={simpleSearch && simpleSearchEnabled}
                            queryState={queryState}
                            setQueryState={setQueryState}
                        />
                        {authenticatedUser && showOnboardingTour && (
                            <GettingStartedTour
                                className="mt-5"
                                telemetryService={telemetryService}
                                variant="horizontal"
                                authenticatedUser={authenticatedUser}
                            />
                        )}
                        {showCodyCTA ? (
                            authenticatedUser ? (
                                <TryCodyCtaSection
                                    className="mx-auto my-5"
                                    telemetryService={telemetryService}
                                    isSourcegraphDotCom={isSourcegraphDotCom}
                                />
                            ) : (
                                <TryCodySignUpCtaSection className="mx-auto my-5" telemetryService={telemetryService} />
                            )
                        ) : null}
                    </>
                )}
            </div>
            {(!simpleSearchEnabled || !simpleSearch) && (
                <div className={classNames(styles.panelsContainer)}>
                    {(!!authenticatedUser || isSourcegraphDotCom) && (
                        <QueryExamples
                            selectedSearchContextSpec={selectedSearchContextSpec}
                            telemetryService={telemetryService}
                            isSourcegraphDotCom={isSourcegraphDotCom}
                        />
                    )}
                </div>
            )}

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
