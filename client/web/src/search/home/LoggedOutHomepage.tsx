import classNames from 'classnames'
import BookOutlineIcon from 'mdi-react/BookOutlineIcon'
import React, { useCallback } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { SyntaxHighlightedSearchQuery } from '../../components/SyntaxHighlightedSearchQuery'
import { FeatureFlagProps } from '../../featureFlags/featureFlags'
import { repogroupList } from '../../repogroups/HomepageConfig'

import { CustomersSection } from './CustomersSection'
import { DynamicWebFonts } from './DynamicWebFonts'
import { HeroSection } from './HeroSection'
import { HomepageModalVideo } from './HomepageModalVideo'
import { SearchExample, exampleNotebooks, exampleQueries, fonts } from './LoggedOutHomepage.constants'
import styles from './LoggedOutHomepage.module.scss'
import { SelfHostInstructions } from './SelfHostInstructions'

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
    const onSignUpClick = useCallback(() => {
        props.telemetryService.log('HomepageCTAClicked', { campaign: 'Sign up link' }, { campaign: 'Sign up link' })
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

                <div className="d-flex justify-content-center">
                    <div className={classNames('card', styles.ctaCard)}>
                        <div className="d-flex align-items-center">
                            <span className="badge badge-merged text-uppercase mr-2">Beta</span>
                        </div>
                        <span>
                            Search your public and private code.{' '}
                            <Link to="/sign-up?src=HomepageCTA" onClick={onSignUpClick}>
                                Sign up
                            </Link>{' '}
                            to get started, or{' '}
                            <a
                                href="https://about.sourcegraph.com/blog/why-index-the-oss-universe/"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                read our blog post
                            </a>{' '}
                            to learn more.
                        </span>
                    </div>
                </div>

                <div className={styles.heroSection}>
                    <HeroSection {...props} />
                </div>

                <div className={styles.repogroupSection}>
                    <div className="d-block d-md-flex align-items-baseline mb-3">
                        <div className={classNames(styles.title, 'mr-2')}>Search open source communities</div>
                        <div className="font-weight-normal text-muted">
                            Customized search portals for our open source partners
                        </div>
                    </div>
                    <div className={styles.loggedOutHomepageRepogroupListCards}>
                        {repogroupList.map(repogroup => (
                            <div
                                className={classNames(
                                    styles.loggedOutHomepageRepogroupListCard,
                                    'd-flex align-items-center'
                                )}
                                key={repogroup.name}
                            >
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

                <div className={styles.selfHostSection}>
                    <SelfHostInstructions {...props} />
                </div>

                <div className={styles.customerSection}>
                    <CustomersSection {...props} />
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
