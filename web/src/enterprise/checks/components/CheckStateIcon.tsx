import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { iconForCheck } from '../util'

interface Props {
    check: Pick<sourcegraph.CheckInformation, 'state'>
    className?: string
}

/**
 * An icon that conveys the state and result of a check.
 */
export const CheckStateIcon: React.FunctionComponent<Props> = ({ check, className = '' }) => {
    const { icon: Icon, className: resultClassName } = iconForCheck(check)
    return <Icon className={`${className} ${resultClassName}`} />
}
