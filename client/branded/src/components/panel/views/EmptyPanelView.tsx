import React from 'react'

import { mdiCancel } from '@mdi/js'
import classNames from 'classnames'

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
                    <Icon className="mr-2" aria-hidden={true} svgPath={mdiCancel} /> Nothing to show here
                </>
            )}
        </div>
    )
}
