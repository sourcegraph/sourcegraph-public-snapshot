import classNames from 'classnames'
import CancelIcon from 'mdi-react/CancelIcon'
import React from 'react'

import styles from './EmptyPanelView.module.scss'

interface EmptyPanelViewProps {
    className?: string
}

export const EmptyPanelView: React.FunctionComponent<EmptyPanelViewProps> = props => {
    const { className, children } = props

    return (
        <div className={classNames(styles.emptyPanel, className)}>
            {children || (
                <>
                    <CancelIcon className="icon-inline mr-2" /> Nothing to show here
                </>
            )}
        </div>
    )
}
