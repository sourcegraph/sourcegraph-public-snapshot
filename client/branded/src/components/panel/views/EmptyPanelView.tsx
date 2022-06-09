import React from 'react'

import classNames from 'classnames'
import CancelIcon from 'mdi-react/CancelIcon'

import { Icon } from '@sourcegraph/wildcard'

import styles from './EmptyPanelView.module.scss'

interface EmptyPanelViewProps {
    className?: string
}

export const EmptyPanelView: React.FunctionComponent<React.PropsWithChildren<EmptyPanelViewProps>> = props => {
    const { className, children } = props

    return (
        <div className={classNames(styles.emptyPanel, className)}>
            {children || (
                <>
                    <Icon role="img" className="mr-2" as={CancelIcon} aria-hidden={true} /> Nothing to show here
                </>
            )}
        </div>
    )
}
