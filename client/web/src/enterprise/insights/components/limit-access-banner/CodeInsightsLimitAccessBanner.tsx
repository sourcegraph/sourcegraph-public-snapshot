import classNames from 'classnames'
import React from 'react'

import { Link } from '@sourcegraph/wildcard'

import styles from './CodeInsightsLimitAccessBanner.module.scss'

interface CodeInsightsLimitAccessBannerProps extends React.HTMLAttributes<HTMLDivElement> {}

export const CodeInsightsLimitAccessBanner: React.FunctionComponent<CodeInsightsLimitAccessBannerProps> = props => (
    <div {...props} className={classNames(styles.banner, props.className)}>
        <div className={styles.content}>
            <h4>You’re currently viewing a demo version of Code Insights</h4>
            <span>
                Contact your admin or <Link to="mailto:support@sourcegraph.com">reach out to us</Link> to upgrade your
                licence for unlimited insights and dashboards.
            </span>
        </div>
    </div>
)
