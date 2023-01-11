import React from 'react'

import { mdiChartLineVariant } from '@mdi/js'
import classNames from 'classnames'

import { Heading, Icon } from '@sourcegraph/wildcard'

import styles from './AnalyticsPageTitle.module.scss'

export const AnalyticsPageTitle: React.FunctionComponent<React.PropsWithChildren<{}>> = ({ children }) => (
    <div className="d-flex flex-column justify-content-between align-items-start">
        <Heading as="h3" styleAs="h2" className="mb-4 mt-2 d-flex align-items-center">
            <Icon
                className="mr-1"
                color="var(--link-color)"
                svgPath={mdiChartLineVariant}
                size="sm"
                aria-label="Analytics icon"
            />
            Analytics
            <span className={classNames(styles.iconColor, 'mx-2')}>/</span>
            {children}
        </Heading>
    </div>
)
