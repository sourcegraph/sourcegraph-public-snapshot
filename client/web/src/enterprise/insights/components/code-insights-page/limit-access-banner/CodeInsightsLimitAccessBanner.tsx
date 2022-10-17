import React from 'react'

import classNames from 'classnames'

import { Badge, Link, Text } from '@sourcegraph/wildcard'

import styles from './CodeInsightsLimitAccessBanner.module.scss'

interface CodeInsightsLimitAccessBannerProps extends React.HTMLAttributes<HTMLDivElement> {}

export const CodeInsightsLimitAccessBanner: React.FunctionComponent<
    React.PropsWithChildren<CodeInsightsLimitAccessBannerProps>
> = props => (
    <div {...props} className={classNames(styles.banner, props.className)}>
        <div className={styles.content}>
            <Badge className={classNames('mb-2', styles.badge)}>LIMITED ACCESS</Badge>
            <Text className="m-0">
                Contact your admin or{' '}
                <Link to="mailto:support@sourcegraph.com" target="_blank" rel="noopener noreferrer">
                    reach out to us
                </Link>{' '}
                to upgrade your Sourcegraph license to unlock Code Insights for unlimited insights and dashboards.{' '}
                <Link to="/help/code_insights/references/license" rel="noopener noreferrer" target="_blank">
                    Learn more
                </Link>
            </Text>
        </div>
    </div>
)
