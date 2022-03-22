import React, { useCallback } from 'react'

import classNames from 'classnames'

import { SyntaxHighlightedSearchQuery, ModalVideo } from '@sourcegraph/search-ui'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Card, Link } from '@sourcegraph/wildcard'

import { communitySearchContextsList } from '../../communitySearchContexts/HomepageConfig'
import { FeatureFlagProps } from '../../featureFlags/featureFlags'
import { GettingStartedTour } from '../../gettingStartedTour/GettingStartedTour'

import { CustomersSection } from './CustomersSection'
import { DynamicWebFonts } from './DynamicWebFonts'
import { HeroSection } from './HeroSection'
import { SearchExample, exampleTripsAndTricks, fonts } from './LoggedOutHomepage.constants'
import { SelfHostInstructions } from './SelfHostInstructions'

import styles from './LoggedOutHomepage.module.scss'

export interface LoggedOutHomepageProps extends TelemetryProps, ThemeProps, FeatureFlagProps {}

interface TipsAndTricksProps extends TelemetryProps {
    title: string
    examples: SearchExample[]
    moreLink: {
        href: string
        label: string
    }
}
const TipsAndTricks: React.FunctionComponent<TipsAndTricksProps> = ({
    title,
    moreLink,
    telemetryService,
    examples,
}) => {
    const searchExampleClicked = useCallback(
        (trackEventName: string) => (): void => telemetryService.log(trackEventName),
        [telemetryService]
    )
    return (
        <div className={classNames(styles.tipsAndTricks)}>
            <div className={classNames('mb-2', styles.title)}>{title}</div>
            <div className={styles.tipsAndTricksExamples}>
                {examples.map(example => (
                    <div key={example.query} className={styles.tipsAndTricksExample}>
                        {example.label}
                        <Card
                            as={Link}
                            to={example.to}
                            className={styles.tipsAndTricksCard}
                            onClick={searchExampleClicked(example.trackEventName)}
                        >
                            <SyntaxHighlightedSearchQuery query={example.query} />
                        </Card>
                    </div>
                ))}
            </div>
            <Link className={styles.tipsAndTricksMore} to={moreLink.href}>
                {moreLink.label}
            </Link>
        </div>
    )
}

export const LoggedOutHomepage: React.FunctionComponent<LoggedOutHomepageProps> = props => (
    <DynamicWebFonts fonts={fonts}>
        <div className={styles.loggedOutHomepage}>
            <div className={styles.content}>
                <GettingStartedTour
                    height={7.5}
                    className={styles.gettingStartedTour}
                    telemetryService={props.telemetryService}
                    featureFlags={props.featureFlags}
                    isSourcegraphDotCom={true}
                />
                <div className={styles.videoCard}>
                    <div className={classNames(styles.title, 'mb-2')}>Watch and learn</div>
                    <ModalVideo
                        id="three-ways-to-search-title"
                        title="Three ways to search"
                        src="https://www.youtube-nocookie.com/embed/XLfE2YuRwvw"
                        showCaption={true}
                        thumbnail={{
                            src: `img/watch-and-learn-${props.isLightTheme ? 'light' : 'dark'}.png`,
                            alt: 'Watch and learn video thumbnail',
                        }}
                        onToggle={isOpen =>
                            props.telemetryService.log(
                                isOpen ? 'HomepageVideoWaysToSearchClicked' : 'HomepageVideoClosed'
                            )
                        }
                        assetsRoot={window.context?.assetsRoot || ''}
                    />
                </div>

                <TipsAndTricks
                    title="Tips and Tricks"
                    examples={exampleTripsAndTricks}
                    moreLink={{
                        label: 'More search features',
                        href: 'https://docs.sourcegraph.com/code_search/explanations/features',
                    }}
                    {...props}
                />
            </div>

            <div className={styles.heroSection}>
                <HeroSection {...props} />
            </div>

            <div className={styles.communitySearchContextsSection}>
                <div className="d-block d-md-flex align-items-baseline mb-3">
                    <div className={classNames(styles.title, 'mr-2')}>Search open source communities</div>
                    <div className="font-weight-normal text-muted">
                        Customized search portals for our open source partners
                    </div>
                </div>
                <div className={styles.loggedOutHomepageCommunitySearchContextListCards}>
                    {communitySearchContextsList.map(communitySearchContext => (
                        <div
                            className={classNames(
                                styles.loggedOutHomepageCommunitySearchContextListCard,
                                'd-flex align-items-center'
                            )}
                            key={communitySearchContext.spec}
                        >
                            <img
                                className={classNames(styles.loggedOutHomepageCommunitySearchContextListIcon, 'mr-2')}
                                src={communitySearchContext.homepageIcon}
                                alt={`${communitySearchContext.spec} icon`}
                            />
                            <Link
                                to={communitySearchContext.url}
                                className={classNames(styles.loggedOutHomepageCommunitySearchContextsListingTitle)}
                            >
                                {communitySearchContext.title}
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
