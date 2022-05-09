import React, { useCallback, useMemo } from 'react'

import classNames from 'classnames'
import BookmarkOutlineIcon from 'mdi-react/BookmarkOutlineIcon'
import FormatQuoteOpenIcon from 'mdi-react/FormatQuoteOpenIcon'
import LinkIcon from 'mdi-react/LinkIcon'

import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import { Icon } from '@sourcegraph/wildcard'

import { WebviewPageProps } from '../../platform/context'

import { ButtonDropdownCta, ButtonDropdownCtaProps } from './ButtonDropdownCta'
import { BookmarkRadialGradientIcon, CodeMonitoringLogo } from './icons'

import styles from './SearchResultsInfoBar.module.scss'

// Debt: this is a fork of the web <SearchResultsInfobar>.
export interface SearchResultsInfoBarProps
    extends Pick<WebviewPageProps, 'extensionCoreAPI' | 'platformContext' | 'authenticatedUser' | 'instanceURL'> {
    stats: JSX.Element

    onShareResultsClick: () => void
    setShowSavedSearchForm: (status: boolean) => void
    showSavedSearchForm: boolean
    fullQuery: string
    patternType: SearchPatternType

    // Expand all feature
    allExpanded: boolean
    onExpandAllResultsToggle: () => void
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
        <button
            type="button"
            className="btn btn-sm btn-outline-secondary text-decoration-none"
            onClick={props.onNonExperimentalLinkClick}
            disabled={props.isNonExperimentalLinkDisabled}
        >
            {props.button}
        </button>
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
    props.patternType === SearchPatternType.literal && props.fullQuery && props.fullQuery.includes('"') ? (
        <small
            className={styles.notice}
            data-tooltip="Your search query is interpreted literally, including the quotes. Use the .* toggle to switch between literal and regular expression search."
        >
            <span>
                <Icon role="img" aria-hidden={true} className="mr-1" as={FormatQuoteOpenIcon} />
                Searching literally <strong>(including quotes)</strong>
            </span>
        </small>
    ) : null

export const SearchResultsInfoBar: React.FunctionComponent<
    React.PropsWithChildren<SearchResultsInfoBarProps>
> = props => {
    const {
        extensionCoreAPI,
        platformContext,
        authenticatedUser,
        showSavedSearchForm,
        setShowSavedSearchForm,
        onShareResultsClick,
        stats,
        instanceURL,
        fullQuery,
        patternType,
    } = props

    const showActionButtonExperimentalVersion = !authenticatedUser

    const onSaveSearchButtonClick = useCallback(
        (event?: React.FormEvent): void => {
            event?.preventDefault()
            setShowSavedSearchForm(!showSavedSearchForm)
            platformContext.telemetryService.log('VSCESaveSearchClick')
        },
        [platformContext.telemetryService, setShowSavedSearchForm, showSavedSearchForm]
    )

    const onCreateCodeMonitorButtonClick = useCallback(
        (event?: React.FormEvent): void => {
            event?.preventDefault()
            platformContext.telemetryService.log('VSCECreateCodeMonitorClick')

            const searchParameters = new URLSearchParams()
            searchParameters.set('q', fullQuery)
            searchParameters.set('trigger-query', `${fullQuery} patternType:${patternType}`)
            const createMonitorURL = new URL(`/code-monitoring/new?${searchParameters.toString()}`, instanceURL)
            extensionCoreAPI.openLink(createMonitorURL.href).catch(() => {
                console.error('Error opening create code monitor link')
            })
        },
        [platformContext.telemetryService, extensionCoreAPI, fullQuery, instanceURL, patternType]
    )

    const canCreateMonitorFromQuery = useMemo(() => {
        if (!fullQuery) {
            return false
        }
        const globalTypeFilterInQuery = findFilter(fullQuery, 'type', FilterKind.Global)
        const globalTypeFilterValue = globalTypeFilterInQuery?.value ? globalTypeFilterInQuery.value.value : undefined
        return globalTypeFilterValue === 'diff' || globalTypeFilterValue === 'commit'
    }, [fullQuery])

    const createCodeMonitorButton = useMemo(() => {
        const searchParameters = new URLSearchParams()
        searchParameters.set('q', fullQuery)
        searchParameters.set('trigger-query', `${fullQuery} patternType:${patternType}`)
        return (
            <li className={classNames('mr-2', styles.navItem)}>
                <ExperimentalActionButton
                    extensionCoreAPI={extensionCoreAPI}
                    showExperimentalVersion={showActionButtonExperimentalVersion}
                    onNonExperimentalLinkClick={onCreateCodeMonitorButtonClick}
                    className="test-save-search-link"
                    data-tooltip={
                        !canCreateMonitorFromQuery
                            ? 'Code monitors only support type:diff or type:commit searches.'
                            : undefined
                    }
                    button={
                        <>
                            <Icon role="img" aria-hidden={true} className="mr-1" as={CodeMonitoringLogo} />
                            Monitor
                        </>
                    }
                    icon={<BookmarkRadialGradientIcon />}
                    title="Monitor code for changes"
                    copyText="Create a monitor and get notified when your code changes. Free for registered users."
                    source="CodeMonitor"
                    viewEventName="VSCECodeMonitorCTAShown"
                    returnTo={`/code-monitoring/new?${searchParameters.toString()}`}
                    telemetryService={platformContext.telemetryService}
                    isNonExperimentalLinkDisabled={!canCreateMonitorFromQuery}
                    instanceURL={instanceURL}
                />
            </li>
        )
    }, [
        fullQuery,
        patternType,
        extensionCoreAPI,
        showActionButtonExperimentalVersion,
        onCreateCodeMonitorButtonClick,
        canCreateMonitorFromQuery,
        platformContext.telemetryService,
        instanceURL,
    ])

    const saveSearchButton = useMemo(
        () => (
            <li className={classNames('mr-2', styles.navItem)}>
                <ExperimentalActionButton
                    extensionCoreAPI={extensionCoreAPI}
                    showExperimentalVersion={showActionButtonExperimentalVersion}
                    onNonExperimentalLinkClick={onSaveSearchButtonClick}
                    className="test-save-search-link"
                    button={
                        <>
                            <Icon role="img" aria-hidden={true} className="mr-1" as={BookmarkOutlineIcon} />
                            Save search
                        </>
                    }
                    icon={<BookmarkRadialGradientIcon />}
                    title="Saved searches"
                    copyText="Save your searches and quickly run them again. Free for registered users."
                    source="SavedSearches"
                    viewEventName="VSCESaveSearchCTAShown"
                    returnTo=""
                    telemetryService={platformContext.telemetryService}
                    isNonExperimentalLinkDisabled={showActionButtonExperimentalVersion}
                    instanceURL={instanceURL}
                    onToggle={() => setShowSavedSearchForm(!showSavedSearchForm)}
                />
            </li>
        ),
        [
            extensionCoreAPI,
            showActionButtonExperimentalVersion,
            onSaveSearchButtonClick,
            platformContext.telemetryService,
            instanceURL,
            setShowSavedSearchForm,
            showSavedSearchForm,
        ]
    )

    const ShareLinkButton = useMemo(
        () => (
            <li className={classNames('mr-2', styles.navItem)} data-tooltip="Share results link">
                <button
                    type="button"
                    className="btn btn-sm btn-outline-secondary text-decoration-none"
                    onClick={onShareResultsClick}
                >
                    <Icon role="img" aria-hidden={true} className="mr-1" as={LinkIcon} />
                    Share
                </button>
            </li>
        ),
        [onShareResultsClick]
    )

    return (
        <div className={classNames('flex-grow-1 my-2', styles.searchResultsInfoBar)} data-testid="results-info-bar">
            <div className={styles.row}>
                {stats}
                <QuotesInterpretedLiterallyNotice {...props} />
                <div className={styles.expander} />
                <ul className="nav align-items-center">
                    {createCodeMonitorButton}
                    {saveSearchButton}
                    {ShareLinkButton}
                </ul>
            </div>
        </div>
    )
}
