import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { iconForCheck } from '../util'

interface Props {
    checkInfo: Pick<sourcegraph.CheckInformation, 'state'>
    className?: string
}

/**
 * An icon that conveys the state and result of a check.
 */
export const CheckStateIcon: React.FunctionComponent<Props> = ({ checkInfo, className = '' }) => {
    const { icon: Icon, className: resultClassName } = iconForCheck(checkInfo)
    return <Icon className={`${className} ${resultClassName}`} />
}
