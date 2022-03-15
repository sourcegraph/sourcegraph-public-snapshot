import React from 'react'

import classNames from 'classnames'

import { Badge, Link } from '@sourcegraph/wildcard'

import styles from './CodeInsightsLimitAccessBanner.module.scss'

interface CodeInsightsLimitAccessBannerProps extends React.HTMLAttributes<HTMLDivElement> {}

export const CodeInsightsLimitAccessBanner: React.FunctionComponent<CodeInsightsLimitAccessBannerProps> = props => (
    <div {...props} className={classNames(styles.banner, props.className)}>
        <div className={styles.content}>
            <Badge variant="merged" className="mb-2">
                LIMITED ACCESS
            </Badge>
            <p className="m-0">
                Contact your admin or{' '}
                <Link
                    to="https://about.sourcegraph.com/contact/request-code-insights-demo?utm_medium=direct-traffic&utm_source=in-product&utm_campaign=code-insights-getting-started"
                    target="_blank"
                    rel="noopener"
                >
                    reach out to us
                </Link>{' '}
                to upgrade your Sourcegraph license to unlock Code Insights for unlimited insights and dashboards.{' '}
                <Link to="/help/code_insights" rel="noopener noreferrer" target="_blank">
                    Learn more
                </Link>
            </p>
        </div>
    </div>
)
