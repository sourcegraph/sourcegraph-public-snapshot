import React from 'react'

import ArrowRightIcon from 'mdi-react/ArrowRightIcon'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Link, Typography } from '@sourcegraph/wildcard'

import styles from './HeroSection.module.scss'

export const HeroSection: React.FunctionComponent<React.PropsWithChildren<ThemeProps & TelemetryProps>> = ({
    isLightTheme,
    telemetryService,
}) => {
    const assetsRoot = window.context?.assetsRoot || ''
    const theme = isLightTheme ? 'light' : 'dark'
    return (
        <div className={styles.hero}>
            <div className={styles.column}>
                <img src={`${assetsRoot}/img/homepage-hero-${theme}.svg`} alt="" className={styles.image} />
            </div>
            <div className={styles.column}>
                <Typography.H2 className={styles.header}>
                    Great code search helps you <br className="d-lg-inline d-none" />
                    <strong className={styles.headerBold}>write, reference, and fix, faster.</strong>
                </Typography.H2>
                <div className={styles.text}>
                    Over 800,000 developers use Sourcegraph to:
                    <ul className="mt-2">
                        <li>Find anything in multiple repositories, fast</li>
                        <li>Navigate with definitions and references</li>
                        <li>Make large-scale code changes</li>
                        <li>Integrate code with other services</li>
                    </ul>
                </div>
                <Link
                    to="https://about.sourcegraph.com/"
                    className={styles.link}
                    onClick={() => telemetryService.log('HomepageAboutSiteLinkClicked')}
                >
                    Learn more about Sourcegraph <ArrowRightIcon className="ml-2" />
                </Link>
            </div>
        </div>
    )
}
