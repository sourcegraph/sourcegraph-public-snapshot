import React from 'react'

import { ListboxButton } from '@reach/listbox'
import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'

import { TruncatedText } from '../../../../../../../components'
import { InsightDashboard, isCustomDashboard } from '../../../../../../../core'
import { getDashboardOwnerName, getDashboardTitle } from '../../helpers/get-dashboard-title'
import { InsightsBadge } from '../insights-badge/InsightsBadge'

import styles from './MenuButton.module.scss'

interface MenuButtonProps {
    dashboards: InsightDashboard[]
    className?: string
}

/**
 * Renders ListBox menu button for dashboard select component.
 */
export const MenuButton: React.FunctionComponent<React.PropsWithChildren<MenuButtonProps>> = props => {
    const { dashboards, className } = props

    return (
        <ListboxButton className={classNames(styles.button, className)}>
            {({ value, isExpanded }) => {
                const dashboard = dashboards.find(dashboard => dashboard.id === value)

                if (!dashboard) {
                    return <MenuButtonContent title="Unknown dashboard" isExpanded={isExpanded} />
                }

                return (
                    <MenuButtonContent
                        title={getDashboardTitle(dashboard)}
                        badge={isCustomDashboard(dashboard) ? getDashboardOwnerName(dashboard) : undefined}
                        isExpanded={isExpanded}
                    />
                )
            }}
        </ListboxButton>
    )
}

interface MenuButtonContentProps {
    title: string
    isExpanded: boolean
    badge?: string
}

const MenuButtonContent: React.FunctionComponent<React.PropsWithChildren<MenuButtonContentProps>> = props => {
    const { title, isExpanded, badge } = props
    const ListboxButtonIcon = isExpanded ? ChevronUpIcon : ChevronDownIcon

    return (
        <>
            <span className={styles.text}>
                <TruncatedText title={title}>{title}</TruncatedText>
                {badge && <InsightsBadge value={badge} className={classNames('ml-1 mr-1', styles.badge)} />}
            </span>

            <ListboxButtonIcon className={styles.expandedIcon} />
        </>
    )
}
