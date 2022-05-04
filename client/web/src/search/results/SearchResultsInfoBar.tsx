import React, { useCallback, useMemo, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import ArrowCollapseUpIcon from 'mdi-react/ArrowCollapseUpIcon'
import ArrowExpandDownIcon from 'mdi-react/ArrowExpandDownIcon'
import BookmarkOutlineIcon from 'mdi-react/BookmarkOutlineIcon'
import FormatQuoteOpenIcon from 'mdi-react/FormatQuoteOpenIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuIcon from 'mdi-react/MenuIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'

import { ContributableMenu } from '@sourcegraph/client-api'
import { SearchPatternTypeProps, CaseSensitivityProps } from '@sourcegraph/search'
import { ActionItem } from '@sourcegraph/shared/src/actions/ActionItem'
import { ActionsContainer } from '@sourcegraph/shared/src/actions/ActionsContainer'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, ButtonLink, Link, useLocalStorage, Icon } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { CodeMonitoringLogo } from '../../code-monitoring/CodeMonitoringLogo'
import { BookmarkRadialGradientIcon, CodeMonitorRadialGradientIcon } from '../../components/CtaIcons'
import { SearchPatternType } from '../../graphql-operations'
import { defaultPopperModifiers } from '../input/tour-options'
import { renderBrandedToString } from '../render-branded-to-string'
import {
    getTourOptions,
    HAS_SEEN_CODE_MONITOR_FEATURE_TOUR_KEY,
    HAS_SEEN_SEARCH_CONTEXTS_FEATURE_TOUR_KEY,
    useFeatureTour,
} from '../useFeatureTour'

import { ButtonDropdownCta, ButtonDropdownCtaProps } from './ButtonDropdownCta'
import { CreateCodeInsightButton } from './components/CreateCodeInsightButton'
import { CreateSearchContextButton } from './components/CreateSearchContextButton'

import featureTourStyles from '../FeatureTour.module.scss'
import styles from './SearchResultsInfoBar.module.scss'

function getFeatureTourElementFn(isAuthenticatedUser: boolean): (onClose: () => void) => HTMLElement {
    return (onClose: () => void): HTMLElement => {
        const container = document.createElement('div')
        container.className = featureTourStyles.featureTourStep
        container.innerHTML = renderBrandedToString(
            <>
                <div>
                    <strong>New</strong>: Create a code monitor to get notified about new search results for a query.{' '}
                    {isAuthenticatedUser ? (
                        <Link to="https://docs.sourcegraph.com/code_monitoring" target="_blank" rel="noopener">
                            Learn more.
                        </Link>
                    ) : null}
                </div>
                <div className="d-flex justify-content-end text-muted">
                    <Button size="sm">Dismiss</Button>
                </div>
            </>
        )

        const button = container.querySelector('button')
        button?.addEventListener('click', onClose)
        return container
    }
}

export interface SearchResultsInfoBarProps
    extends ExtensionsControllerProps<'executeCommand' | 'extHostAPI'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        TelemetryProps,
        SearchPatternTypeProps,
        Pick<CaseSensitivityProps, 'caseSensitive'> {
    history: H.History
    /** The currently authenticated user or null */
    authenticatedUser: Pick<AuthenticatedUser, 'id'> | null

    /**
     * Whether the code insights feature flag is enabled.
     */
    enableCodeInsights?: boolean
    enableCodeMonitoring: boolean

    /** The search query and if any results were found */
    query?: string
    resultsFound: boolean

    // Expand all feature
    allExpanded: boolean
    onExpandAllResultsToggle: () => void

    // Saved queries
    onSaveQueryClick: () => void

    location: H.Location

    className?: string

    stats: JSX.Element

    onShowFiltersChanged?: (show: boolean) => void
}

interface ExperimentalActionButtonProps extends ButtonDropdownCtaProps {
    showExperimentalVersion: boolean
    nonExperimentalLinkTo?: string
    isNonExperimentalLinkDisabled?: boolean
    onNonExperimentalLinkClick?: () => void
    className?: string
}

const ExperimentalActionButton: React.FunctionComponent<
    React.PropsWithChildren<ExperimentalActionButtonProps>
> = props => {
    if (props.showExperimentalVersion) {
        return <ButtonDropdownCta {...props} />
    }
    return (
        <ButtonLink
            className={classNames('text-decoration-none', props.className)}
            to={props.nonExperimentalLinkTo}
            onSelect={props.onNonExperimentalLinkClick}
            disabled={props.isNonExperimentalLinkDisabled}
            variant="secondary"
            outline={true}
            size="sm"
        >
            {props.button}
        </ButtonLink>
    )
}

/**
 * A notice for when the user is searching literally and has quotes in their
 * query, in which case it is possible that they think their query `"foobar"`
 * will be searching literally for `foobar` (without quotes). This notice
 * informs them that this may be the case to avoid confusion.
 */
const QuotesInterpretedLiterallyNotice: React.FunctionComponent<
    React.PropsWithChildren<SearchResultsInfoBarProps>
> = props =>
    props.patternType === SearchPatternType.literal && props.query && props.query.includes('"') ? (
        <small
            className={styles.notice}
            data-tooltip="Your search query is interpreted literally, including the quotes. Use the .* toggle to switch between literal and regular expression search."
        >
            <span>
                <Icon as={FormatQuoteOpenIcon} />
                Searching literally <strong>(including quotes)</strong>
            </span>
        </small>
    ) : null

/**
 * The info bar shown over the search results list that displays metadata
 * and a few actions like expand all and save query
 */
export const SearchResultsInfoBar: React.FunctionComponent<
    React.PropsWithChildren<SearchResultsInfoBarProps>
> = props => {
    const canCreateMonitorFromQuery = useMemo(() => {
        if (!props.query) {
            return false
        }
        const globalTypeFilterInQuery = findFilter(props.query, 'type', FilterKind.Global)
        const globalTypeFilterValue = globalTypeFilterInQuery?.value ? globalTypeFilterInQuery.value.value : undefined
        return globalTypeFilterValue === 'diff' || globalTypeFilterValue === 'commit'
    }, [props.query])

    const showCreateCodeMonitoringButton = props.enableCodeMonitoring && !!props.query

    const [hasSeenSearchContextsFeatureTour] = useLocalStorage(HAS_SEEN_SEARCH_CONTEXTS_FEATURE_TOUR_KEY, false)

    const tour = useFeatureTour(
        'create-code-monitor-feature-tour',
        showCreateCodeMonitoringButton &&
            canCreateMonitorFromQuery &&
            hasSeenSearchContextsFeatureTour &&
            props.resultsFound,
        getFeatureTourElementFn(!!props.authenticatedUser),
        HAS_SEEN_CODE_MONITOR_FEATURE_TOUR_KEY,
        getTourOptions({
            attachTo: {
                element: '.create-code-monitor-button',
                on: 'bottom',
            },
            popperOptions: {
                modifiers: [...defaultPopperModifiers, { name: 'offset', options: { offset: [-100, 16] } }],
            },
        })
    )

    const onCreateCodeMonitorButtonSelect = useCallback(() => {
        if (tour.isActive()) {
            props.telemetryService.log('SignUpPLGMonitor_0_Tour')
        }
        tour.cancel()
    }, [props.telemetryService, tour])

    const showActionButtonExperimentalVersion = !props.authenticatedUser

    const searchContextButton = useMemo(
        () => <CreateSearchContextButton query={props.query} authenticatedUser={props.authenticatedUser} />,
        [props.authenticatedUser, props.query]
    )

    const codeInsightsButton = useMemo(
        () => (
            <CreateCodeInsightButton
                query={props.query}
                authenticatedUser={props.authenticatedUser}
                patternType={props.patternType}
                enableCodeInsights={props.enableCodeInsights}
            />
        ),
        [props.authenticatedUser, props.enableCodeInsights, props.patternType, props.query]
    )

    const createCodeMonitorButton = useMemo(() => {
        if (!showCreateCodeMonitoringButton) {
            return null
        }
        const searchParameters = new URLSearchParams(props.location.search)
        searchParameters.set('trigger-query', `${props.query ?? ''} patterntype:${props.patternType}`)
        const toURL = `/code-monitoring/new?${searchParameters.toString()}`
        return (
            <li
                className={classNames('mr-2', styles.navItem)}
                data-tooltip={
                    props.authenticatedUser && !canCreateMonitorFromQuery
                        ? 'Code monitors only support type:diff or type:commit searches.'
                        : undefined
                }
            >
                {/*
                    a11y-ignore
                    Rule: "color-contrast" (Elements must have sufficient color contrast)
                    GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/33343
                */}
                <ExperimentalActionButton
                    showExperimentalVersion={showActionButtonExperimentalVersion}
                    nonExperimentalLinkTo={toURL}
                    isNonExperimentalLinkDisabled={!canCreateMonitorFromQuery}
                    onNonExperimentalLinkClick={onCreateCodeMonitorButtonSelect}
                    className="a11y-ignore create-code-monitor-button"
                    button={
                        <>
                            <Icon className="mr-1" as={CodeMonitoringLogo} />
                            Monitor
                        </>
                    }
                    icon={<CodeMonitorRadialGradientIcon />}
                    title="Monitor code for changes"
                    copyText="Create a monitor and get notified when your code changes. Free for registered users."
                    telemetryService={props.telemetryService}
                    source="Monitor"
                    viewEventName="SearchResultMonitorCTAShown"
                    returnTo={toURL}
                    onToggle={onCreateCodeMonitorButtonSelect}
                />
            </li>
        )
    }, [
        showActionButtonExperimentalVersion,
        showCreateCodeMonitoringButton,
        props.authenticatedUser,
        props.location.search,
        props.query,
        props.patternType,
        props.telemetryService,
        canCreateMonitorFromQuery,
        onCreateCodeMonitorButtonSelect,
    ])

    const saveSearchButton = useMemo(
        () => (
            <li className={classNames('mr-2', styles.navItem)}>
                <ExperimentalActionButton
                    showExperimentalVersion={showActionButtonExperimentalVersion}
                    onNonExperimentalLinkClick={props.onSaveQueryClick}
                    className="test-save-search-link"
                    button={
                        <>
                            <Icon className="mr-1" as={BookmarkOutlineIcon} />
                            Save search
                        </>
                    }
                    icon={<BookmarkRadialGradientIcon />}
                    title="Saved searches"
                    copyText="Save your searches and quickly run them again. Free for registered users."
                    source="Saved"
                    viewEventName="SearchResultSavedSeachCTAShown"
                    returnTo={props.location.pathname + props.location.search}
                    telemetryService={props.telemetryService}
                />
            </li>
        ),
        [props.location, showActionButtonExperimentalVersion, props.onSaveQueryClick, props.telemetryService]
    )

    const extraContext = useMemo(
        () => ({
            searchQuery: props.query || null,
            patternType: props.patternType,
            caseSensitive: props.caseSensitive,
        }),
        [props.query, props.patternType, props.caseSensitive]
    )

    const [showFilters, setShowFilters] = useState(false)
    const onShowFiltersClicked = (): void => {
        const newShowFilters = !showFilters
        setShowFilters(newShowFilters)
        props.onShowFiltersChanged?.(newShowFilters)
    }

    return (
        <div className={classNames(props.className, styles.searchResultsInfoBar)} data-testid="results-info-bar">
            <div className={styles.row}>
                <Button
                    className={classNames('d-flex d-lg-none', showFilters && 'active')}
                    aria-pressed={showFilters}
                    onClick={onShowFiltersClicked}
                    outline={true}
                    variant="secondary"
                    size="sm"
                >
                    <Icon className="mr-1" as={MenuIcon} />
                    Filters
                    <Icon as={showFilters ? MenuUpIcon : MenuDownIcon} />
                </Button>

                {props.stats}

                <QuotesInterpretedLiterallyNotice {...props} />

                <div className={styles.expander} />

                <ul className="nav align-items-center">
                    <ActionsContainer
                        {...props}
                        extraContext={extraContext}
                        menu={ContributableMenu.SearchResultsToolbar}
                    >
                        {actionItems => (
                            <>
                                {actionItems.map(actionItem => (
                                    <Button
                                        {...props}
                                        {...actionItem}
                                        key={actionItem.action.id}
                                        showLoadingSpinnerDuringExecution={false}
                                        className="mr-2 text-decoration-none"
                                        variant="secondary"
                                        outline={true}
                                        size="sm"
                                        as={ActionItem}
                                    />
                                ))}
                            </>
                        )}
                    </ActionsContainer>

                    {(searchContextButton || codeInsightsButton || createCodeMonitorButton || saveSearchButton) && (
                        <li className={styles.divider} aria-hidden="true" />
                    )}

                    {searchContextButton}
                    {codeInsightsButton}
                    {createCodeMonitorButton}
                    {saveSearchButton}

                    {props.resultsFound && (
                        <>
                            <li className={styles.divider} aria-hidden="true" />
                            <li className={classNames(styles.navItem)}>
                                <Button
                                    onClick={props.onExpandAllResultsToggle}
                                    className="text-decoration-none"
                                    data-tooltip={`${props.allExpanded ? 'Hide' : 'Show'} more matches on all results`}
                                    aria-label={`${props.allExpanded ? 'Hide' : 'Show'} more matches on all results`}
                                    aria-live="polite"
                                    data-testid="search-result-expand-btn"
                                    outline={true}
                                    variant="secondary"
                                    size="sm"
                                >
                                    <Icon
                                        className="mr-0"
                                        as={props.allExpanded ? ArrowCollapseUpIcon : ArrowExpandDownIcon}
                                    />
                                </Button>
                            </li>
                        </>
                    )}
                </ul>
            </div>
        </div>
    )
}
