import classNames from 'classnames'
import BookmarkOutlineIcon from 'mdi-react/BookmarkOutlineIcon'
import LinkIcon from 'mdi-react/LinkIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

import { StreamingProgress } from '@sourcegraph/branded/src/search/results/progress/StreamingProgress'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { BookmarkRadialGradientIcon } from './icons'
import styles from './SearchResults.module.scss'
interface VsceSearchResultsInfoBarProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    onShareResultsClick: () => void
    setOpenSavedSearchCreateForm: (status: boolean) => void
    openSavedSearchCreateForm: boolean
    results: AggregateStreamingSearchResults
    onFeedbackClick: () => void
}

interface ExperimentalActionButtonProps extends ButtonDropdownCtaProps {
    showExperimentalVersion: boolean
    nonExperimentalLinkTo?: string
    isNonExperimentalLinkDisabled?: boolean
    onNonExperimentalLinkClick?: () => void
    className?: string
}

interface ButtonDropdownCtaProps extends TelemetryProps {
    button: JSX.Element
    icon: JSX.Element
    title: string
    copyText: string
    source: string
    viewEventName: string
    returnTo: string
    onToggle?: () => void
    className?: string
}

export const VsceSearchResultsInfoBar: React.FunctionComponent<VsceSearchResultsInfoBarProps> = ({
    authenticatedUser,
    onShareResultsClick,
    telemetryService,
    results,
    setOpenSavedSearchCreateForm,
    openSavedSearchCreateForm,
    onFeedbackClick,
}) => {
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

    const showActionButtonExperimentalVersion = !authenticatedUser

    const saveSearchButton = useMemo(
        () => (
            <li className={classNames('mr-2', styles.navItem)}>
                <ExperimentalActionButton
                    showExperimentalVersion={showActionButtonExperimentalVersion}
                    onNonExperimentalLinkClick={() => setOpenSavedSearchCreateForm(!openSavedSearchCreateForm)}
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
                    telemetryService={telemetryService}
                    isNonExperimentalLinkDisabled={showActionButtonExperimentalVersion}
                />
            </li>
        ),
        [openSavedSearchCreateForm, setOpenSavedSearchCreateForm, showActionButtonExperimentalVersion, telemetryService]
    )

    return (
        <div className={classNames('flex-grow-1', styles.searchResultsInfoBar)}>
            <div className={styles.row}>
                <StreamingProgress
                    progress={results?.progress || { durationMs: 0, matchCount: 0, skipped: [] }}
                    state={results?.state || 'loading'}
                    // TODO IMPLEMENT ONSEARCHAGAIN
                    onSearchAgain={() => console.log('Search Again')}
                    showTrace={false}
                />

                <div className={styles.expander} />
                <ul className="nav align-items-center">
                    <li className={styles.divider} aria-hidden="true" />
                    {/* Feedback Button */}
                    <li className={classNames('mr-2', styles.navItem)} data-tooltip="Feedback">
                        <button
                            type="button"
                            className="btn btn-sm btn-primary border-0 text-decoration-none"
                            onClick={onFeedbackClick}
                        >
                            Feedback
                        </button>
                    </li>
                    {saveSearchButton}
                    {/* Share Link Button */}
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

export const ButtonDropdownCta: React.FunctionComponent<ButtonDropdownCtaProps> = ({
    button,
    icon,
    title,
    copyText,
    source,
    telemetryService,
    viewEventName,
    returnTo,
    onToggle,
    className,
}) => {
    const [isDropdownOpen, setIsDropdownOpen] = useState(false)

    const toggleDropdownOpen = useCallback(() => {
        setIsDropdownOpen(isOpen => !isOpen)
        onToggle?.()
    }, [onToggle])

    const onClick = (): void => {
        telemetryService.log('SignUpPLG_VSCE_1_Search')
    }

    // Whenever dropdown opens, log view event
    useEffect(() => {
        if (isDropdownOpen) {
            telemetryService.log(viewEventName)
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [isDropdownOpen])

    return (
        <ButtonDropdown className="menu-nav-item" direction="down" isOpen={isDropdownOpen} toggle={toggleDropdownOpen}>
            <DropdownToggle
                className={classNames(
                    'btn btn-sm btn-outline-secondary text-decoration-none',
                    className,
                    styles.buttonDropdownCtaToggle
                )}
                nav={true}
                caret={false}
            >
                {button}
            </DropdownToggle>
            <DropdownMenu right={true} className={styles.buttonDropdownCtaContainer}>
                <div className={styles.buttonDropdownCtaWrapper}>
                    <div className="d-flex align-items-center mr-3">
                        <div className={styles.buttonDropdownCtaIcon}>{icon}</div>
                    </div>
                    <div>
                        <div className={styles.buttonDropdownCtaTitle}>
                            <strong>{title}</strong>
                        </div>
                        <div className={classNames('text-muted', styles.buttonDropdownCtaCopyText)}>{copyText}</div>
                    </div>
                </div>
                <div className={styles.buttonDropdownCtaWrapper}>
                    <Link
                        className={classNames('btn btn-primary', styles.buttonDropdownCtaButton)}
                        to={`/sign-up?src=${source}&returnTo=${encodeURIComponent(returnTo)}`}
                        onClick={onClick}
                    >
                        Sign up for Sourcegraph
                    </Link>
                </div>
            </DropdownMenu>
        </ButtonDropdown>
    )
}
