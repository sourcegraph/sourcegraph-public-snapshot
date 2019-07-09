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
export const CheckStateBar: React.FunctionComponent<Props> = ({ checkProvider, className = '' }) => (
    <div className={`d-flex align-items-center border border-${themeColorForCheck(checkProvider.check)} ${className}`}>
        <CheckStateIcon checkInfo={checkProvider.check} className="icon-inline mr-3" />
        {checkProvider.check.state.message && (
            <span className={`text-${themeColorForCheck(checkProvider.check)}`}>
                {checkProvider.check.state.message}
            </span>
        )}
        <div className="flex-1" />
    </div>
)
