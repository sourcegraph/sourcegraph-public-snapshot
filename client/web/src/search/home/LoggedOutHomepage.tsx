import React from 'react'

import classNames from 'classnames'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { H4, Text, Link } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'

import { DynamicWebFonts } from './DynamicWebFonts'
import { exampleTripsAndTricks, fonts } from './LoggedOutHomepage.constants'
import { TipsAndTricks } from './TipsAndTricks'

import styles from './LoggedOutHomepage.module.scss'

export interface LoggedOutHomepageProps extends TelemetryProps, ThemeProps {}

export const LoggedOutHomepage: React.FunctionComponent<React.PropsWithChildren<LoggedOutHomepageProps>> = props => (
    <DynamicWebFonts fonts={fonts}>
        <div className={styles.loggedOutHomepage}>
            <div className={styles.content}>
                <TipsAndTricks
                    examples={exampleTripsAndTricks}
                    moreLink={{
                        label: 'More search features',
                        href: 'https://docs.sourcegraph.com/code_search/explanations/features',
                        trackEventName: 'HomepageExampleMoreSearchFeaturesClicked',
                    }}
                    {...props}
                />
            </div>
            
            <div className="d-md-flex d-flex-wrap justify-content-lg-center my-5">
                <div className="d-flex align-items-center">
                    <H4 className={classNames('mr-2 mb-0 pr-2', styles.proTipTitle)}>Pro Tip</H4>
                    <Text className="mb-0">Search on your private repositories.</Text>
                </div>
                <Link
                    to="https://signup.sourcegraph.com/"
                    onClick={() => eventLogger.log('ClickedOnCloudCTA')}
                    className="ml-md-1"
                >
                    Sign up for a 30-day trial on Sourcegraph Cloud.
                </Link>
            </div>
        </div>
    </DynamicWebFonts>
)
