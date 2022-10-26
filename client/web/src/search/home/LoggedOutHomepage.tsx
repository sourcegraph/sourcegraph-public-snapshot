import React from 'react'

import classNames from 'classnames'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { GettingStartedTour } from '../../tour/GettingStartedTour'

import { DynamicWebFonts } from './DynamicWebFonts'
import { exampleTripsAndTricks, fonts } from './LoggedOutHomepage.constants'
import { TipsAndTricks } from './TipsAndTricks'
import { TrySourcegraphCloudSection } from './TrySourcegraphCloudSection'

import styles from './LoggedOutHomepage.module.scss'

export interface LoggedOutHomepageProps extends TelemetryProps, ThemeProps {}

export const LoggedOutHomepage: React.FunctionComponent<React.PropsWithChildren<LoggedOutHomepageProps>> = props => (
    <DynamicWebFonts fonts={fonts}>
        <div className={styles.loggedOutHomepage}>
            <div className={styles.content}>
                <GettingStartedTour
                    height={8}
                    className={classNames(styles.gettingStartedTour, 'h-100')}
                    telemetryService={props.telemetryService}
                    isSourcegraphDotCom={true}
                />
                <TipsAndTricks
                    title="Tips and Tricks"
                    examples={exampleTripsAndTricks}
                    moreLink={{
                        label: 'More search features',
                        href: 'https://docs.sourcegraph.com/code_search/explanations/features',
                        trackEventName: 'HomepageExampleMoreSearchFeaturesClicked',
                    }}
                    {...props}
                />
            </div>

            <div className={styles.trySourcegraphCloudSection}>
                <TrySourcegraphCloudSection />
            </div>
        </div>
    </DynamicWebFonts>
)
