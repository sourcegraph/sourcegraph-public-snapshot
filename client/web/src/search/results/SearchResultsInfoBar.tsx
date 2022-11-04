import React, { useMemo, useState } from 'react'

import {
    mdiBookmarkOutline,
    mdiArrowExpandDown,
    mdiArrowCollapseUp,
    mdiChevronDoubleUp,
    mdiChevronDoubleDown,
} from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'

import { ContributableMenu } from '@sourcegraph/client-api'
import { SearchPatternTypeProps, CaseSensitivityProps } from '@sourcegraph/search'
import { ActionItem } from '@sourcegraph/shared/src/actions/ActionItem'
import { ActionsContainer } from '@sourcegraph/shared/src/actions/ActionsContainer'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import { useCoreWorkflowImprovementsEnabled } from '@sourcegraph/shared/src/settings/useCoreWorkflowImprovementsEnabled'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, ButtonLink, Icon, Tooltip } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { BookmarkRadialGradientIcon, CodeMonitorRadialGradientIcon } from '../../components/CtaIcons'
import { eventLogger } from '../../tracking/eventLogger'

import { ButtonDropdownCta, ButtonDropdownCtaProps } from './ButtonDropdownCta'
import {
    getCodeMonitoringCreateAction,
    getInsightsCreateAction,
    getSearchContextCreateAction,
    getBatchChangeCreateAction,
    CreateAction,
} from './createActions'
import { CreateActionsMenu } from './CreateActionsMenu'
import { SearchActionsMenu } from './SearchActionsMenu'

import createActionsStyles from './CreateActions.module.scss'
import styles from './SearchResultsInfoBar.module.scss'

export interface SearchResultsInfoBarProps
    extends ExtensionsControllerProps<'executeCommand' | 'extHostAPI'>,
        PlatformContextProps<'settings' | 'sourcegraphURL'>,
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

    batchChangesEnabled?: boolean
    /** Whether running batch changes server-side is enabled */
    batchChangesExecutionEnabled?: boolean

    // Expand all feature
    allExpanded: boolean
    onExpandAllResultsToggle: () => void

    // Saved queries
    onSaveQueryClick: () => void

    location: H.Location

    className?: string

    stats: JSX.Element

    onShowMobileFiltersChanged?: (show: boolean) => void

    sidebarCollapsed: boolean
    setSidebarCollapsed: (collapsed: boolean) => void
}

interface ExperimentalActionButtonProps extends ButtonDropdownCtaProps {
    showExperimentalVersion: boolean
    nonExperimentalLinkTo?: string
    isNonExperimentalLinkDisabled?: boolean
    onNonExperimentalLinkClick?: () => void
    className?: string
    ariaLabel?: string
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
            aria-disabled={props.isNonExperimentalLinkDisabled ? 'true' : undefined}
            aria-label={props.ariaLabel}
            // to make disabled ButtonLink focusable
            tabIndex={0}
        >
            {props.button}
        </ButtonLink>
    )
}

/**
 * The info bar shown over the search results list that displays metadata
 * and a few actions like expand all and save query
 */
export const SearchResultsInfoBar: React.FunctionComponent<
    React.PropsWithChildren<SearchResultsInfoBarProps>
> = props => {
    const [coreWorkflowImprovementsEnabled] = useCoreWorkflowImprovementsEnabled()

    const canCreateMonitorFromQuery = useMemo(() => {
        if (!props.query) {
            return false
        }
        const globalTypeFilterInQuery = findFilter(props.query, 'type', FilterKind.Global)
        const globalTypeFilterValue = globalTypeFilterInQuery?.value ? globalTypeFilterInQuery.value.value : undefined
        return globalTypeFilterValue === 'diff' || globalTypeFilterValue === 'commit'
    }, [props.query])

    const showActionButtonExperimentalVersion = !props.authenticatedUser

    // When adding a new create action check and update the $collapse-breakpoint in CreateActions.module.scss.
    // The collapse breakpoint indicates at which window size we hide the buttons and show the collapsed menu instead.
    const createActions = useMemo(
        () =>
            [
                getBatchChangeCreateAction(
                    props.query,
                    props.patternType,
                    props.authenticatedUser,
                    props.batchChangesEnabled && props.batchChangesExecutionEnabled
                ),
                getSearchContextCreateAction(props.query, props.authenticatedUser),
                getInsightsCreateAction(
                    props.query,
                    props.patternType,
                    props.authenticatedUser,
                    props.enableCodeInsights
                ),
            ].filter((button): button is CreateAction => button !== null),
        [
            props.authenticatedUser,
            props.enableCodeInsights,
            props.patternType,
            props.query,
            props.batchChangesEnabled,
            props.batchChangesExecutionEnabled,
        ]
    )

    // The create code monitor action is separated from the rest of the actions, because we use the
    // <ExperimentalActionButton /> component instead of a regular (button) link, and it has a tour attached.
    const createCodeMonitorAction = useMemo(
        () => getCodeMonitoringCreateAction(props.query, props.patternType, props.enableCodeMonitoring),
        [props.enableCodeMonitoring, props.patternType, props.query]
    )

    const createCodeMonitorButton = useMemo(() => {
        if (!createCodeMonitorAction) {
            return null
        }

        return (
            <Tooltip
                content={
                    props.authenticatedUser && !canCreateMonitorFromQuery
                        ? 'Code monitors only support type:diff or type:commit searches.'
                        : undefined
                }
                placement="bottom"
            >
                <li className={classNames('mr-2', createActionsStyles.button, styles.navItem)}>
                    {/*
                    a11y-ignore
                    Rule: "color-contrast" (Elements must have sufficient color contrast)
                    GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/33343
                */}
                    <ExperimentalActionButton
                        showExperimentalVersion={showActionButtonExperimentalVersion}
                        nonExperimentalLinkTo={createCodeMonitorAction.url}
                        isNonExperimentalLinkDisabled={!canCreateMonitorFromQuery}
                        className="a11y-ignore create-code-monitor-button"
                        button={
                            <>
                                <Icon
                                    aria-hidden={true}
                                    className="mr-1"
                                    {...(typeof createCodeMonitorAction.icon === 'string'
                                        ? { svgPath: createCodeMonitorAction.icon }
                                        : { as: createCodeMonitorAction.icon })}
                                />
                                {createCodeMonitorAction.label}
                            </>
                        }
                        icon={<CodeMonitorRadialGradientIcon />}
                        title="Monitor code for changes"
                        copyText="Create a monitor and get notified when your code changes. Free for registered users."
                        telemetryService={props.telemetryService}
                        source="Monitor"
                        viewEventName="SearchResultMonitorCTAShown"
                        returnTo={createCodeMonitorAction.url}
                        ariaLabel={
                            props.authenticatedUser && !canCreateMonitorFromQuery
                                ? 'Code monitors only support type:diff or type:commit searches.'
                                : undefined
                        }
                    />
                </li>
            </Tooltip>
        )
    }, [
        createCodeMonitorAction,
        props.telemetryService,
        props.authenticatedUser,
        canCreateMonitorFromQuery,
        showActionButtonExperimentalVersion,
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
                            <Icon aria-hidden={true} className="mr-1" svgPath={mdiBookmarkOutline} />
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

    // Show/hide mobile filters menu
    const [showMobileFilters, setShowMobileFilters] = useState(false)
    const onShowMobileFiltersClicked = (): void => {
        const newShowFilters = !showMobileFilters
        setShowMobileFilters(newShowFilters)
        props.onShowMobileFiltersChanged?.(newShowFilters)
    }

    const { extensionsController } = props

    return (
        <aside
            role="region"
            aria-label="Search results information"
            className={classNames(props.className, styles.searchResultsInfoBar)}
            data-testid="results-info-bar"
        >
            <div className={styles.row}>
                {props.stats}

                <div className={styles.expander} />

                <ul className="nav align-items-center">
                    {extensionsController !== null && window.context.enableLegacyExtensions ? (
                        <ActionsContainer
                            {...props}
                            extensionsController={extensionsController}
                            extraContext={extraContext}
                            menu={ContributableMenu.SearchResultsToolbar}
                        >
                            {actionItems => (
                                <>
                                    {actionItems.map(actionItem => (
                                        <ActionItem
                                            {...props}
                                            {...actionItem}
                                            extensionsController={extensionsController}
                                            key={actionItem.action.id}
                                            showLoadingSpinnerDuringExecution={false}
                                            className="mr-2 text-decoration-none"
                                            actionItemStyleProps={{
                                                actionItemVariant: 'secondary',
                                                actionItemSize: 'sm',
                                                actionItemOutline: true,
                                            }}
                                        />
                                    ))}
                                </>
                            )}
                        </ActionsContainer>
                    ) : null}

                    {(createActions.length > 0 ||
                        createCodeMonitorButton ||
                        saveSearchButton ||
                        coreWorkflowImprovementsEnabled) && <li className={styles.divider} aria-hidden="true" />}

                    {coreWorkflowImprovementsEnabled ? (
                        <SearchActionsMenu
                            query={props.query}
                            patternType={props.patternType}
                            sourcegraphURL={props.platformContext.sourcegraphURL}
                            authenticatedUser={props.authenticatedUser}
                            createActions={createActions}
                            createCodeMonitorAction={createCodeMonitorAction}
                            canCreateMonitor={canCreateMonitorFromQuery}
                            resultsFound={props.resultsFound}
                            allExpanded={props.allExpanded}
                            onExpandAllResultsToggle={props.onExpandAllResultsToggle}
                            onSaveQueryClick={props.onSaveQueryClick}
                        />
                    ) : (
                        <>
                            {createActions.map(createActionButton => (
                                <Tooltip
                                    key={createActionButton.label}
                                    content={createActionButton.tooltip}
                                    placement="bottom"
                                >
                                    <li className={classNames('nav-item mr-2', createActionsStyles.button)}>
                                        <ButtonLink
                                            to={createActionButton.url}
                                            className="text-decoration-none"
                                            variant="secondary"
                                            outline={true}
                                            size="sm"
                                            onClick={
                                                createActionButton.eventToLog
                                                    ? () => {
                                                          eventLogger.log(createActionButton.eventToLog!)
                                                      }
                                                    : undefined
                                            }
                                        >
                                            <Icon
                                                aria-hidden={true}
                                                className="mr-1"
                                                {...(typeof createActionButton.icon === 'string'
                                                    ? { svgPath: createActionButton.icon }
                                                    : { as: createActionButton.icon })}
                                            />
                                            {createActionButton.label}
                                        </ButtonLink>
                                    </li>
                                </Tooltip>
                            ))}

                            {createCodeMonitorButton}

                            {(createActions.length > 0 || createCodeMonitorAction) && (
                                <CreateActionsMenu
                                    createActions={createActions}
                                    createCodeMonitorAction={createCodeMonitorAction}
                                    canCreateMonitor={canCreateMonitorFromQuery}
                                    authenticatedUser={props.authenticatedUser}
                                />
                            )}

                            {saveSearchButton}

                            {props.resultsFound && (
                                <>
                                    <li className={styles.divider} aria-hidden="true" />
                                    <li className={classNames(styles.navItem)}>
                                        <Tooltip
                                            content={`${
                                                props.allExpanded ? 'Hide' : 'Show'
                                            } more matches on all results`}
                                            placement="bottom"
                                        >
                                            <Button
                                                aria-label={props.allExpanded ? 'Collapse' : 'Expand'}
                                                onClick={props.onExpandAllResultsToggle}
                                                className="text-decoration-none"
                                                aria-live="polite"
                                                data-testid="search-result-expand-btn"
                                                data-test-tooltip-content={`${
                                                    props.allExpanded ? 'Hide' : 'Show'
                                                } more matches on all results`}
                                                outline={true}
                                                variant="secondary"
                                                size="sm"
                                            >
                                                <Icon
                                                    aria-hidden={true}
                                                    className="mr-0"
                                                    svgPath={
                                                        props.allExpanded ? mdiArrowCollapseUp : mdiArrowExpandDown
                                                    }
                                                />
                                            </Button>
                                        </Tooltip>
                                    </li>
                                </>
                            )}
                        </>
                    )}
                </ul>

                <Button
                    className={classNames(
                        'd-flex align-items-center d-lg-none',
                        styles.filtersButton,
                        showMobileFilters && 'active'
                    )}
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

                {props.sidebarCollapsed && (
                    <Button
                        className={classNames('align-items-center d-none d-lg-flex', styles.filtersButton)}
                        onClick={() => props.setSidebarCollapsed(false)}
                        outline={true}
                        variant="secondary"
                        size="sm"
                        aria-label="Show filters sidebar"
                    >
                        Filters
                        <Icon aria-hidden={true} className="ml-2" svgPath={mdiChevronDoubleDown} />
                    </Button>
                )}
            </div>
        </aside>
    )
}
