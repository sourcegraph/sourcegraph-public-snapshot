import classNames from 'classnames'
import CancelIcon from 'mdi-react/CancelIcon'
import React from 'react'

import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

interface EmptyPanelViewProps {
    className?: string
}

export const EmptyPanelView: React.FunctionComponent<EmptyPanelViewProps> = props => {
    const { className } = props
    const [isRedesignEnabled] = useRedesignToggle()
    const EmptyPanelWrapper = isRedesignEnabled ? React.Fragment : 'div'

    return (
        <EmptyPanelWrapper {...(!isRedesignEnabled && { className: 'panel' })}>
            <div className={classNames('panel__empty', className)}>
                <CancelIcon className="icon-inline" /> Nothing to show here
            </div>
        </EmptyPanelWrapper>
    )
}
