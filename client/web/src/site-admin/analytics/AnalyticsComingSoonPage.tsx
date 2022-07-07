import React, { useMemo } from 'react'

import { mdiChartTimelineVariantShimmer } from '@mdi/js'
import classNames from 'classnames'
import { upperFirst } from 'lodash'
import { RouteComponentProps } from 'react-router'

import { H3, Text, Icon } from '@sourcegraph/wildcard'

import { AnalyticsPageTitle } from './components/AnalyticsPageTitle'

import styles from './index.module.scss'

export const AnalyticsComingSoonPage: React.FunctionComponent<RouteComponentProps<{}>> = props => {
    const title = useMemo(() => {
        const title = props.match.path.split('/').filter(Boolean)[2] ?? 'Overview'
        return upperFirst(title.replace('-', ' '))
    }, [props.match.path])
    return (
        <>
            <AnalyticsPageTitle>Analytics / {title}</AnalyticsPageTitle>

            <div className="d-flex flex-column justify-content-center align-items-center p-5">
                <Icon
                    svgPath={mdiChartTimelineVariantShimmer}
                    aria-label="Home analytics icon"
                    className={classNames(styles.largeIcon, 'm-3')}
                />
                <H3>Coming soon</H3>
                <Text>We are working on making this live.</Text>
            </div>
        </>
    )
}
