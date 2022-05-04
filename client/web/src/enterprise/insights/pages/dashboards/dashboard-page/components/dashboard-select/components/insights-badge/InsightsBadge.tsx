import React from 'react'

import classNames from 'classnames'

import { Badge } from '@sourcegraph/wildcard'

import truncatedStyles from '../../../../../../../components/trancated-text/TruncatedText.module.scss'
import styles from './InsightsBadge.module.scss'

interface BadgeProps {
    value: string
    className?: string
}

/**
 * A wrapper around the Wildcard badge component with some slightly different styling.
 * We can't use the standard "secondary" variant here because the selected select option
 * already uses this color for the background. We use this style variant to avoid visual merging/overlapping.
 */
export const InsightsBadge: React.FunctionComponent<React.PropsWithChildren<BadgeProps>> = props => {
    const { value, className } = props

    return (
        <Badge
            title={value}
            variant="secondary"
            className={classNames(styles.badge, truncatedStyles.truncatedText, className)}
        >
            {value}
        </Badge>
    )
}
