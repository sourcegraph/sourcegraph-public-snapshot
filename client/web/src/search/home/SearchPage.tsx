import React, { useEffect, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { QueryState, SearchContextInputProps } from '@sourcegraph/shared/src/search'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { buildCloudTrialURL } from '@sourcegraph/shared/src/util/url'
import { Link, Tooltip, useWindowSize, VIEWPORT_SM } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { BrandLogo } from '../../components/branding/BrandLogo'
import { CodeInsightsProps } from '../../insights/types'
import { AddCodeHostWidget, useShouldShowAddCodeHostWidget } from '../../onboarding/AddCodeHostWidget'
import { useExperimentalFeatures } from '../../stores'
import { ThemePreferenceProps } from '../../theme'
import { eventLogger } from '../../tracking/eventLogger'

import { QueryExamplesHomepage } from './QueryExamplesHomepage'
import { SearchPageFooter } from './SearchPageFooter'
import { SearchPageInput } from './SearchPageInput'

import styles from './SearchPage.module.scss'

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
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    autoFocus?: boolean

    // Whether globbing is enabled for filters.
    globbing: boolean
}

/**
 * The search page
 */
export const SearchPage: React.FunctionComponent<React.PropsWithChildren<SearchPageProps>> = props => {
    const homepageUserInvitation = useExperimentalFeatures(features => features.homepageUserInvitation) ?? false
    const showCollaborators = window.context.allowSignup && homepageUserInvitation && props.isSourcegraphDotCom
    const { width } = useWindowSize()
    const shouldShowAddCodeHostWidget = useShouldShowAddCodeHostWidget(props.authenticatedUser)

    /** The value entered by the user in the query input */
    const [queryState, setQueryState] = useState<QueryState>({
        query: '',
    })

    useEffect(() => props.telemetryService.logViewEvent('Home'), [props.telemetryService])

    return (
        <div className={classNames('d-flex flex-column align-items-center px-3', styles.searchPage)}>
            <BrandLogo className={styles.logo} isLightTheme={props.isLightTheme} variant="logo" />
            {props.isSourcegraphDotCom && (
                <div className="d-sm-flex flex-row text-center">
                    <div className={classNames(width >= VIEWPORT_SM && 'border-right', 'text-muted mt-3 mr-sm-2 pr-2')}>
                        Search millions of open source repositories
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
            <div
                className={classNames(styles.panelsContainer, {
                    [styles.panelsContainerWithCollaborators]: showCollaborators,
                })}
            >
                {(!!props.authenticatedUser || props.isSourcegraphDotCom) && (
                    <QueryExamplesHomepage
                        selectedSearchContextSpec={props.selectedSearchContextSpec}
                        telemetryService={props.telemetryService}
                        queryState={queryState}
                        setQueryState={setQueryState}
                        isSourcegraphDotCom={props.isSourcegraphDotCom}
                        authenticatedUser={props.authenticatedUser}
                    />
                )}
            </div>

            <SearchPageFooter {...props} />
        </div>
    )
}
