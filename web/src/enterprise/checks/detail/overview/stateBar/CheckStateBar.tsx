import React from 'react'
import { CheckStateIcon } from '../../../components/CheckStateIcon'
import { themeColorForCheck } from '../../../util'
import { CheckAreaContext } from '../../CheckArea'

interface Props extends Pick<CheckAreaContext, 'check'> {
    className?: string
}

/**
 * A bar that displays the state of a check.
 */
export const CheckStateBar: React.FunctionComponent<Props> = ({ check, className = '' }) => (
    <div className={`d-flex align-items-center border border-${themeColorForCheck(check.check)} ${className}`}>
        <CheckStateIcon check={check.check} className="icon-inline mr-3" />
        {check.check.state.message && (
            <span className={`text-${themeColorForCheck(check.check)}`}>{check.check.state.message}</span>
        )}
        <div className="flex-1" />
    </div>
)
