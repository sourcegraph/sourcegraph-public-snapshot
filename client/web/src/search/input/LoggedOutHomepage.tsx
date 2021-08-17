import classNames from 'classnames'
import BookOutlineIcon from 'mdi-react/BookOutlineIcon'
import React, { useCallback } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { SyntaxHighlightedSearchQuery } from '../../components/SyntaxHighlightedSearchQuery'
import { FeatureFlagProps } from '../../featureFlags/featureFlags'
import { repogroupList } from '../../repogroups/HomepageConfig'

import { DynamicWebFonts } from './DynamicWebFonts'
import { HomepageModalVideo } from './HomepageModalVideo'
import { SearchExample, exampleNotebooks, exampleQueries, fonts } from './LoggedOutHomepage.constants'
import styles from './LoggedOutHomepage.module.scss'
import { SignUpCta } from './SignUpCta'

export interface LoggedOutHomepageProps extends TelemetryProps, ThemeProps, FeatureFlagProps {}

interface SearchExamplesProps extends TelemetryProps {
    title: string
    subtitle: string
    examples: SearchExample[]
    icon: JSX.Element
}

const SearchExamples: React.FunctionComponent<SearchExamplesProps> = ({
    title,
    subtitle,
    telemetryService,
    examples,
    icon,
}) => {
    const searchExampleClicked = useCallback(
        (trackEventName: string) => (): void => telemetryService.log(trackEventName),
        [telemetryService]
    )

    return (
        <div className={styles.searchExamplesWrapper}>
            <div className={classNames('d-flex align-items-baseline mb-2', styles.searchExamplesTitleWrapper)}>
                <div className={classNames('mr-2', styles.title, styles.searchExamplesTitle)}>{title}</div>
                <div className="font-weight-normal text-muted">{subtitle}</div>
            </div>
            <div className={styles.searchExamples}>
                {examples.map(example => (
                    <div key={example.query} className={styles.searchExampleCardWrapper}>
                        <Link
                            to={example.to}
                            className={classNames('card', styles.searchExampleCard)}
                            onClick={searchExampleClicked(example.trackEventName)}
                        >
                            <div className={classNames(styles.searchExampleIcon)}>{icon}</div>
                            <div className={styles.searchExampleQueryWrapper}>
                                <div className={styles.searchExampleQuery}>
                                    <SyntaxHighlightedSearchQuery query={example.query} />
                                </div>
                            </div>
                        </Link>
                        <Link to={example.to} onClick={searchExampleClicked(example.trackEventName)}>
                            {example.label}
                        </Link>
                    </div>
                ))}
            </div>
        </div>
    )
}

export const LoggedOutHomepage: React.FunctionComponent<LoggedOutHomepageProps> = props => {
    const onClickInstallSubtext = useCallback(() => {
        props.telemetryService.log(
            'HomepageInstallSourcegraphCTAClicked',
            { name: 'InstallSourcegraphSubtext' },
            { name: 'InstallSourcegraphSubtext' }
        )
    }, [props.telemetryService])

    return (
        <DynamicWebFonts fonts={fonts}>
            <div className={styles.loggedOutHomepage}>
                <div className={styles.helpContent}>
                    {props.featureFlags.get('search-notebook-onboarding') ? (
                        <SearchExamples
                            title="Search notebooks"
                            subtitle="Three ways code search is more efficient than your IDE"
                            examples={exampleNotebooks}
                            icon={<BookOutlineIcon />}
                            {...props}
                        />
                    ) : (
                        <SearchExamples
                            title="Search examples"
                            subtitle="Find answers faster with code search across multiple repos and commits"
                            examples={exampleQueries}
                            icon={<MagnifyingGlassSearchIcon />}
                            {...props}
                        />
                    )}
                    <div className={styles.thumbnail}>
                        <div className={classNames(styles.title, 'mb-2')}>Watch and learn</div>
                        <HomepageModalVideo {...props} />
                    </div>
                </div>

                <div className="mt-5 d-flex justify-content-center">
                    <div className="d-flex align-items-center flex-column">
                        <SignUpCta className={styles.loggedOutHomepageCta} telemetryService={props.telemetryService} />
                        <div className="mt-2 text-center">
                            Search private code by{' '}
                            <a
                                href="https://docs.sourcegraph.com/admin/install"
                                target="_blank"
                                rel="noopener noreferrer"
                                onClick={onClickInstallSubtext}
                            >
                                installing Sourcegraph locally.
                            </a>
                        </div>
                    </div>
                </div>

                <div className="mt-5">
                    <div className="d-flex align-items-baseline mt-5 mb-3">
                        <div className={classNames(styles.title, 'mr-2')}>Repository groups</div>
                        <div className="font-weight-normal text-muted">Search sets of repositories</div>
                    </div>
                    <div className={styles.loggedOutHomepageRepogroupListCards}>
                        {repogroupList.map(repogroup => (
                            <div className="d-flex align-items-center" key={repogroup.name}>
                                <img
                                    className={classNames(styles.loggedOutHomepageRepogroupListIcon, 'mr-2')}
                                    src={repogroup.homepageIcon}
                                    alt={`${repogroup.name} icon`}
                                />
                                <Link
                                    to={repogroup.url}
                                    className={classNames(styles.loggedOutHomepageRepogroupListingTitle)}
                                >
                                    {repogroup.title}
                                </Link>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </DynamicWebFonts>
    )
}

const MagnifyingGlassSearchIcon = React.memo(() => (
    <svg width="18" height="18" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path
            d="M6.686.5a6.672 6.672 0 016.685 6.686 6.438 6.438 0 01-1.645 4.32l.308.308h.823L18 16.957 16.457 18.5l-5.143-5.143v-.823l-.308-.308a6.438 6.438 0 01-4.32 1.645A6.672 6.672 0 010 7.186 6.672 6.672 0 016.686.5zm0 2.057a4.61 4.61 0 00-4.629 4.629 4.61 4.61 0 004.629 4.628 4.61 4.61 0 004.628-4.628 4.61 4.61 0 00-4.628-4.629z"
            fill="currentColor"
        />
    </svg>
))
