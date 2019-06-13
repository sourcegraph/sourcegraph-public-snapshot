import React from 'react'
import { CheckStateIcon } from '../../../components/CheckStateIcon'
import { themeColorForCheck } from '../../../util'
import { CheckAreaContext } from '../../CheckArea'

interface Props extends Pick<CheckAreaContext, 'checkInfo'> {
    className?: string
}

/**
 * A bar that displays the state of a check.
 */
export const CheckStateBar: React.FunctionComponent<Props> = ({ checkInfo, className = '' }) => (
    <div className={`d-flex align-items-center border border-${themeColorForCheck(checkInfo)} ${className}`}>
        <CheckStateIcon checkInfo={checkInfo} className="icon-inline mr-3" />
        {checkInfo.state.message && (
            <span className={`text-${themeColorForCheck(checkInfo)}`}>{checkInfo.state.message}</span>
        )}
        <div className="flex-1" />
    </div>
)
