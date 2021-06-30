import { ListboxOption } from '@reach/listbox'
import React from 'react'

import { InsightDashboard } from '../../../../../../core/types'
import { getDashboardOwnerName, getDashboardTitle } from '../../helpers/get-dashboard-title'
import { Badge } from '../badge/Badge'
import { TruncatedText } from '../trancated-text/TrancatedText'

import styles from './SelectOption.module.scss'

interface SelectOptionProps {
    dashboard: InsightDashboard
}

export const SelectOption: React.FunctionComponent<SelectOptionProps> = props => {
    const { dashboard } = props

    return (
        <ListboxOption className={styles.listboxOption} value={dashboard.id}>
            <TruncatedText className={styles.listboxOptionText}>{getDashboardTitle(dashboard)}</TruncatedText>
            <Badge value={getDashboardOwnerName(dashboard)} className={styles.listboxOptionBadge} />
        </ListboxOption>
    )
}
