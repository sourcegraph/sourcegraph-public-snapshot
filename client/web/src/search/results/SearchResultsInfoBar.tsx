import classNames from 'classnames'
import * as H from 'history'
import ArrowCollapseUpIcon from 'mdi-react/ArrowCollapseUpIcon'
import ArrowExpandDownIcon from 'mdi-react/ArrowExpandDownIcon'
import BookmarkOutlineIcon from 'mdi-react/BookmarkOutlineIcon'
import FormatQuoteOpenIcon from 'mdi-react/FormatQuoteOpenIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuIcon from 'mdi-react/MenuIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import React, { useCallback, useMemo, useState } from 'react'

import { ContributableMenu } from '@sourcegraph/shared/src/api/protocol'
import { ButtonLink } from '@sourcegraph/shared/src/components/LinkOrButton'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/validate'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { PatternTypeProps, CaseSensitivityProps } from '..'
import { AuthenticatedUser } from '../../auth'
import { CodeMonitoringProps } from '../../code-monitoring'
import { CodeMonitoringLogo } from '../../code-monitoring/CodeMonitoringLogo'
import { WebActionsNavItems as ActionsNavItems } from '../../components/shared'
import { SearchPatternType } from '../../graphql-operations'
import styles from '../FeatureTour.module.scss'
import { defaultPopperModifiers } from '../input/tour-options'
import {
    getTourOptions,
    HAS_SEEN_CODE_MONITOR_FEATURE_TOUR_KEY,
    HAS_SEEN_SEARCH_CONTEXTS_FEATURE_TOUR_KEY,
    useFeatureTour,
} from '../useFeatureTour'

import { ButtonDropdownCta, ButtonDropdownCtaProps } from './ButtonDropdownCta'
import { CreateCodeInsightButton } from './components/CreateCodeInsightButton'

function getFeatureTourElementFn(isAuthenticatedUser: boolean): (onClose: () => void) => HTMLElement {
    return (onClose: () => void): HTMLElement => {
        const container = document.createElement('div')
        container.className = styles.featureTourStep
        container.innerHTML = `
            <div>
                <strong>New</strong>: Create a code monitor to get notified about new search results for a query.
                ${
                    isAuthenticatedUser
                        ? '<a href="https://docs.sourcegraph.com/code_monitoring" target="_blank">Learn more.</a>'
                        : ''
                }
            </div>
            <div class="d-flex justify-content-end text-muted">
                <button type="button" class="btn btn-sm">
                    Dismiss
                </button>
            </div>
        `
        const button = container.querySelector('button')
        button?.addEventListener('click', onClose)
        return container
    }
}

export interface SearchResultsInfoBarProps
    extends ExtensionsControllerProps<'executeCommand' | 'extHostAPI'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        TelemetryProps,
        Pick<PatternTypeProps, 'patternType'>,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        CodeMonitoringProps {
    history: H.History
    /** The currently authenticated user or null */
    authenticatedUser: AuthenticatedUser | null

    /**
     * Whether the code insights feature flag is enabled.
     */
    enableCodeInsights?: boolean

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

interface ButtonWithUnauthenticatedUserCtaPromptProps extends ButtonDropdownCtaProps {
    authenticatedUser: AuthenticatedUser | null
    authenticatedLinkTo?: string
    className?: string
    isAuthenticatedLinkDisabled?: boolean
    onAuthenticatedLinkClick?: () => void
}

const ButtonWithUnauthenticatedUserCtaPrompt: React.FunctionComponent<ButtonWithUnauthenticatedUserCtaPromptProps> = props => {
    if (props.authenticatedUser) {
        return (
            <ButtonLink
                className={classNames(
                    'btn btn-sm btn-outline-secondary mr-2 nav-link text-decoration-none',
                    props.className
                )}
                to={props.authenticatedLinkTo}
                onSelect={props.onAuthenticatedLinkClick}
                disabled={props.isAuthenticatedLinkDisabled}
            >
                {props.button}
            </ButtonLink>
        )
    }
    return <ButtonDropdownCta {...props} />
}

/**
 * A notice for when the user is searching literally and has quotes in their
 * query, in which case it is possible that they think their query `"foobar"`
 * will be searching literally for `foobar` (without quotes). This notice
 * informs them that this may be the case to avoid confusion.
 */
const QuotesInterpretedLiterallyNotice: React.FunctionComponent<SearchResultsInfoBarProps> = props =>
    props.patternType === SearchPatternType.literal && props.query && props.query.includes('"') ? (
        <small
            className="search-results-info-bar__notice"
            data-tooltip="Your search query is interpreted literally, including the quotes. Use the .* toggle to switch between literal and regular expression search."
        >
            <span>
                <FormatQuoteOpenIcon className="icon-inline" />
                Searching literally <strong>(including quotes)</strong>
            </span>
        </small>
    ) : null

/**
 * The info bar shown over the search results list that displays metadata
 * and a few actions like expand all and save query
 */
export const SearchResultsInfoBar: React.FunctionComponent<SearchResultsInfoBarProps> = props => {
    const canCreateMonitorFromQuery = useMemo(() => {
        if (!props.query) {
            return false
        }
        const globalTypeFilterInQuery = findFilter(props.query, 'type', FilterKind.Global)
        const globalTypeFilterValue = globalTypeFilterInQuery?.value ? globalTypeFilterInQuery.value.value : undefined
        return globalTypeFilterValue === 'diff' || globalTypeFilterValue === 'commit'
    }, [props.query])

    const showCreateCodeMonitoringButton = props.enableCodeMonitoring && props.query
    const [hasSeenSearchContextsFeatureTour] = useLocalStorage(HAS_SEEN_SEARCH_CONTEXTS_FEATURE_TOUR_KEY, false)
    const tour = useFeatureTour(
        'create-code-monitor-feature-tour',
        !!showCreateCodeMonitoringButton &&
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

    const createCodeMonitorButton = useMemo(() => {
        if (!showCreateCodeMonitoringButton) {
            return null
        }
        const searchParameters = new URLSearchParams(props.location.search)
        searchParameters.set('trigger-query', `${props.query ?? ''} patterntype:${props.patternType}`)
        const toURL = `/code-monitoring/new?${searchParameters.toString()}`
        return (
            <li
                className="nav-item"
                data-tooltip={
                    props.authenticatedUser && !canCreateMonitorFromQuery
                        ? 'Code monitors only support type:diff or type:commit searches.'
                        : undefined
                }
            >
                <ButtonWithUnauthenticatedUserCtaPrompt
                    authenticatedUser={props.authenticatedUser}
                    authenticatedLinkTo={toURL}
                    isAuthenticatedLinkDisabled={!canCreateMonitorFromQuery}
                    onAuthenticatedLinkClick={onCreateCodeMonitorButtonSelect}
                    className="create-code-monitor-button"
                    button={
                        <>
                            <CodeMonitoringLogo className="icon-inline mr-1" />
                            Monitor
                        </>
                    }
                    icon={<CodeMonitoringLogo />}
                    title="Monitor code for changes"
                    copyText="Create a monitor and get notified when your code changes. Free for registered users."
                    telemetryService={props.telemetryService}
                    eventLogName="SignUpPLGMonitor_1_Search"
                    onToggle={onCreateCodeMonitorButtonSelect}
                />
            </li>
        )
    }, [
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
            <li className="nav-item">
                <ButtonWithUnauthenticatedUserCtaPrompt
                    authenticatedUser={props.authenticatedUser}
                    onAuthenticatedLinkClick={props.onSaveQueryClick}
                    className="test-save-search-link"
                    button={
                        <>
                            <BookmarkOutlineIcon className="icon-inline mr-1" />
                            Save search
                        </>
                    }
                    icon={<BookmarkOutlineIcon />}
                    title="Saved searches"
                    copyText="Save your searches and quickly run them again. Free for registered users."
                    telemetryService={props.telemetryService}
                    eventLogName="SignUpPLGSaved_1_Search"
                />
            </li>
        ),
        [props.authenticatedUser, props.onSaveQueryClick, props.telemetryService]
    )

    const extendButton = useMemo(
        () => (
            <li className="nav-item">
                <ButtonWithUnauthenticatedUserCtaPrompt
                    authenticatedUser={props.authenticatedUser}
                    authenticatedLinkTo="/extensions"
                    button={
                        <>
                            <PuzzleOutlineIcon className="icon-inline mr-1" />
                            Extend
                        </>
                    }
                    icon={<PuzzleOutlineIcon />}
                    title="Extend your search experience"
                    copyText="Customize workflows, display data alongside your code, and extend the UI via Sourcegraph extensions."
                    telemetryService={props.telemetryService}
                    eventLogName="SignUpPLGExtend_1_Search"
                />
            </li>
        ),
        [props.authenticatedUser, props.telemetryService]
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
        <div className={classNames(props.className, 'search-results-info-bar')} data-testid="results-info-bar">
            <div className="search-results-info-bar__row">
                <button
                    type="button"
                    className={classNames('btn btn-sm btn-outline-secondary d-flex d-lg-none', showFilters && 'active')}
                    aria-pressed={showFilters}
                    onClick={onShowFiltersClicked}
                >
                    <MenuIcon className="icon-inline mr-1" />
                    Filters
                    {showFilters ? <MenuUpIcon className="icon-inline" /> : <MenuDownIcon className="icon-inline" />}
                </button>

                {props.stats}

                <QuotesInterpretedLiterallyNotice {...props} />

                <div className="search-results-info-bar__expander" />

                <ul className="nav align-items-center">
                    <ActionsNavItems
                        {...props}
                        extraContext={extraContext}
                        menu={ContributableMenu.SearchResultsToolbar}
                        wrapInList={false}
                        showLoadingSpinnerDuringExecution={true}
                        actionItemClass="btn nav-link btn-outline-secondary mr-2 text-decoration-none btn-sm"
                    />

                    {extendButton}
                    {(createCodeMonitorButton || saveSearchButton) && (
                        <li className="search-results-info-bar__divider" aria-hidden="true" />
                    )}
                    <CreateCodeInsightButton
                        query={props.query}
                        authenticatedUser={props.authenticatedUser}
                        patternType={props.patternType}
                        enableCodeInsights={props.enableCodeInsights}
                    />
                    {createCodeMonitorButton}
                    {saveSearchButton}

                    {props.resultsFound && (
                        <>
                            <li className="search-results-info-bar__divider" aria-hidden="true" />
                            <li className="nav-item">
                                <button
                                    type="button"
                                    onClick={props.onExpandAllResultsToggle}
                                    className="btn btn-sm btn-outline-secondary nav-link text-decoration-none"
                                    data-tooltip={`${props.allExpanded ? 'Hide' : 'Show'} more matches on all results`}
                                >
                                    {props.allExpanded ? (
                                        <ArrowCollapseUpIcon className="icon-inline mr-0" />
                                    ) : (
                                        <ArrowExpandDownIcon className="icon-inline mr-0" />
                                    )}
                                </button>
                            </li>
                        </>
                    )}
                </ul>
            </div>
        </div>
    )
}
