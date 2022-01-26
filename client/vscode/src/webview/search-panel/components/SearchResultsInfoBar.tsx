import classNames from 'classnames'
import BookmarkOutlineIcon from 'mdi-react/BookmarkOutlineIcon'
import FormatQuoteOpenIcon from 'mdi-react/FormatQuoteOpenIcon'
import LinkIcon from 'mdi-react/LinkIcon'
import React, { useMemo } from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/schema'

import { WebviewPageProps } from '../../platform/context'

import { ButtonDropdownCta, ButtonDropdownCtaProps } from './ButtonDropdownCta'
import { BookmarkRadialGradientIcon } from './icons'
import styles from './SearchResultsInfoBar.module.scss'

// Debt: this is a fork of the web <SearchResultsInfobar>.
export interface SearchResultsInfoBarProps
    extends Pick<WebviewPageProps, 'extensionCoreAPI' | 'platformContext' | 'authenticatedUser'> {
    stats: JSX.Element

    onShareResultsClick: () => void
    setShowSavedSearchForm: (status: boolean) => void
    showSavedSearchForm: boolean
    query?: string
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

const ExperimentalActionButton: React.FunctionComponent<ExperimentalActionButtonProps> = props => {
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
const QuotesInterpretedLiterallyNotice: React.FunctionComponent<SearchResultsInfoBarProps> = props =>
    props.patternType === SearchPatternType.literal && props.query && props.query.includes('"') ? (
        <small
            className={styles.notice}
            data-tooltip="Your search query is interpreted literally, including the quotes. Use the .* toggle to switch between literal and regular expression search."
        >
            <span>
                <FormatQuoteOpenIcon className="icon-inline" />
                Searching literally <strong>(including quotes)</strong>
            </span>
        </small>
    ) : null

export const SearchResultsInfoBar: React.FunctionComponent<SearchResultsInfoBarProps> = props => {
    const {
        extensionCoreAPI,
        platformContext,
        authenticatedUser,
        showSavedSearchForm,
        setShowSavedSearchForm,
        onShareResultsClick,
        stats,
    } = props

    const showActionButtonExperimentalVersion = !authenticatedUser

    const saveSearchButton = useMemo(
        () => (
            <li className={classNames('mr-2', styles.navItem)}>
                <ExperimentalActionButton
                    showExperimentalVersion={showActionButtonExperimentalVersion}
                    onNonExperimentalLinkClick={() => setShowSavedSearchForm(!showSavedSearchForm)}
                    className="test-save-search-link"
                    button={
                        <>
                            <BookmarkOutlineIcon className="icon-inline mr-1" />
                            Save search
                        </>
                    }
                    icon={<BookmarkRadialGradientIcon />}
                    title="Saved searches"
                    copyText="Save your searches and quickly run them again. Free for registered users."
                    source="Saved"
                    viewEventName="SearchResultSavedSeachCTAShown"
                    returnTo=""
                    telemetryService={platformContext.telemetryService}
                    isNonExperimentalLinkDisabled={showActionButtonExperimentalVersion}
                />
            </li>
        ),
        [showSavedSearchForm, setShowSavedSearchForm, showActionButtonExperimentalVersion, platformContext]
    )

    return (
        <div className={classNames('flex-grow-1 my-2', styles.searchResultsInfoBar)} data-testid="results-info-bar">
            <div className={styles.row}>
                {stats}

                <QuotesInterpretedLiterallyNotice {...props} />

                <div className={styles.expander} />

                <ul className="nav align-items-center">
                    <li className={classNames('mr-2', styles.navItem)} data-tooltip="Feedback">
                        <button
                            type="button"
                            className="btn btn-sm btn-primary border-0 text-decoration-none"
                            onClick={() =>
                                extensionCoreAPI.openLink(
                                    'https://github.com/sourcegraph/sourcegraph/discussions/categories/feedback'
                                )
                            }
                        >
                            Feedback
                        </button>
                    </li>
                    {saveSearchButton}
                    <li className={classNames('mr-2', styles.navItem)} data-tooltip="Share results link">
                        <button
                            type="button"
                            className="btn btn-sm btn-outline-secondary text-decoration-none"
                            onClick={onShareResultsClick}
                        >
                            <LinkIcon className="icon-inline mr-1" />
                            Share
                        </button>
                    </li>
                </ul>
            </div>
        </div>
    )
}
