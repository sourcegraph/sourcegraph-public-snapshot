import { ListboxOption } from '@reach/listbox'
import classnames from 'classnames'
import React from 'react'

import { InsightDashboard } from '../../../../../../core/types'
import { getDashboardOwnerName, getDashboardTitle } from '../../helpers/get-dashboard-title'
import { Badge } from '../badge/Badge'
import { TruncatedText } from '../trancated-text/TrancatedText'

import styles from './SelectOption.module.scss'

interface SelectOptionProps {
    dashboard: InsightDashboard
    className?: string
}

export const SelectOption: React.FunctionComponent<SelectOptionProps> = props => {
    const { dashboard, className } = props

    const optionText = getDashboardTitle(dashboard)

    return (
        <ListboxOption className={classnames(styles.listboxOption, className)} value={dashboard.id}>
            <TruncatedText title={optionText} className={styles.listboxOptionText}>
                {optionText}
            </TruncatedText>
            <Badge value={getDashboardOwnerName(dashboard)} className={styles.listboxOptionBadge} />
        </ListboxOption>
    )
}
