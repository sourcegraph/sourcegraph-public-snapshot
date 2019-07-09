import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { iconForStatus } from '../util'

interface Props {
    status: Pick<sourcegraph.Status, 'state'>
    className?: string
}

/**
 * An icon that conveys the state and result of a status.
 */
export const StatusStateIcon: React.FunctionComponent<Props> = ({ status, className = '' }) => {
    const { icon: Icon, className: resultClassName } = iconForStatus(status)
    return <Icon className={`${className} ${resultClassName}`} />
}
