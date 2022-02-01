import classNames from 'classnames'
import * as React from 'react'

import styles from './PanelContainer.module.scss'

interface Props {
    title: string
    state: 'loading' | 'populated' | 'empty'
    // the content displayed when state is 'loading'
    loadingContent?: JSX.Element
    // the content displayed when state is 'empty'
    emptyContent?: JSX.Element
    // the content displayed when state is 'populated'
    populatedContent: JSX.Element
    actionButtons?: JSX.Element
    className?: string
}

export const PanelContainer: React.FunctionComponent<Props> = ({
    title,
    state,
    loadingContent = <></>,
    emptyContent = <></>,
    populatedContent,
    actionButtons,
    className,
}) => (
    <div className={classNames(className, styles.panelContainer, 'd-flex', 'flex-column')}>
        <div className={classNames('d-flex border-bottom', styles.header)}>
            <h4 className={styles.headerText}>{title}</h4>
            {actionButtons}
        </div>

        <div className={classNames('h-100', styles.content)}>
            {state === 'loading' && loadingContent}
            {state === 'populated' && populatedContent}
            {state === 'empty' && emptyContent}
        </div>
    </div>
)
