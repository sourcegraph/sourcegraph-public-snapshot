import { FC, useEffect, useState } from 'react'

import classNames from 'classnames'

import { QueryExamples } from '@sourcegraph/branded/src/search-ui/components/QueryExamples'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { QueryState, SearchContextInputProps } from '@sourcegraph/shared/src/search'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { buildCloudTrialURL } from '@sourcegraph/shared/src/util/url'
import { Link, Tooltip, useWindowSize, VIEWPORT_SM } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { BrandLogo } from '../../../components/branding/BrandLogo'
import { CodeInsightsProps } from '../../../insights/types'
import {
    AddCodeHostWidget,
    getShouldShowAddCodeHostWidget,
    useShouldShowAddCodeHostWidget,
} from '../../../onboarding/AddCodeHostWidget'
import { useExperimentalFeatures } from '../../../stores'
import { ThemePreferenceProps } from '../../../theme'
import { eventLogger } from '../../../tracking/eventLogger'

import { SearchPageFooter } from './SearchPageFooter/SearchPageFooter'
import { SearchPageInput } from './SearchPageInput/SearchPageInput'

import styles from './SearchPage.module.scss'
import { usePageLoaderData } from './SearchPage.loader'

export interface SearchPageProps
    extends SettingsCascadeProps<Settings>,
        ThemeProps,
        ThemePreferenceProps,
        TelemetryProps,
        ExtensionsControllerProps<'extHostAPI' | 'executeCommand'>,
        PlatformContextProps<'settings' | 'sourcegraphURL' | 'updateSettings' | 'requestGraphQL'>,
        SearchContextInputProps,
        CodeInsightsProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    autoFocus?: boolean

    // Whether globbing is enabled for filters.
    globbing: boolean
}

console.log('SearchPage chunk is here!')

/**
 * The search page
 */
export const SearchPage: FC<SearchPageProps> = props => {
    const { data } = usePageLoaderData()

    // Use page loader data instead issues API requests form the render function via `useShouldShowAddCodeHostWidget`.
    const shouldShowAddCodeHostWidget = getShouldShowAddCodeHostWidget(
        data?.evaluateFeatureFlag.value,
        props.authenticatedUser?.siteAdmin,
        data?.externalServices.totalCount
    )

    // Note: if we keep hooks that initiates two queries, no additional requests are made because of Apollo Cache!
    // const shouldShowAddCodeHostWidget = useShouldShowAddCodeHostWidget(props.authenticatedUser)

    const { width } = useWindowSize()
    const experimentalQueryInput = useExperimentalFeatures(features => features.searchQueryInput === 'experimental')

    /** The value entered by the user in the query input */
    const [queryState, setQueryState] = useState<QueryState>({
        query: '',
    })

    useEffect(() => {
        if (experimentalQueryInput && props.selectedSearchContextSpec) {
            setQueryState(state =>
                state.query === '' ? { query: `context:${props.selectedSearchContextSpec} ` } : state
            )
        }
    }, [experimentalQueryInput, props.selectedSearchContextSpec])

    useEffect(() => props.telemetryService.logViewEvent('Home'), [props.telemetryService])

    return (
        <div className={classNames('d-flex flex-column align-items-center px-3', styles.searchPage)}>
            <BrandLogo className={styles.logo} isLightTheme={props.isLightTheme} variant="logo" />
            {props.isSourcegraphDotCom && (
                <div className="d-sm-flex flex-row text-center">
                    <div className={classNames(width >= VIEWPORT_SM && 'border-right', 'text-muted mt-3 mr-sm-2 pr-2')}>
                        Search millions of public repositories
                    </div>
                    <div className="mt-3">
                        <Link
                            to={buildCloudTrialURL(props.authenticatedUser)}
                            onClick={() => eventLogger.log('ClickedOnCloudCTA', { cloudCtaType: 'HomeAboveSearch' })}
                        >
                            Search private code
                        </Link>
                    </div>
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
                                <SearchPageInput
                                    {...props}
                                    queryState={queryState}
                                    setQueryState={setQueryState}
                                    source="home"
                                />
                            </div>
                        </Tooltip>
                        <AddCodeHostWidget className="mb-4" telemetryService={props.telemetryService} />
                    </>
                ) : (
                    <SearchPageInput {...props} queryState={queryState} setQueryState={setQueryState} source="home" />
                )}
            </div>
            <div className={classNames(styles.panelsContainer)}>
                {(!!props.authenticatedUser || props.isSourcegraphDotCom) && (
                    <QueryExamples
                        selectedSearchContextSpec={props.selectedSearchContextSpec}
                        telemetryService={props.telemetryService}
                        queryState={queryState}
                        setQueryState={setQueryState}
                        isSourcegraphDotCom={props.isSourcegraphDotCom}
                    />
                )}
            </div>

            <SearchPageFooter {...props} />
        </div>
    )
}
