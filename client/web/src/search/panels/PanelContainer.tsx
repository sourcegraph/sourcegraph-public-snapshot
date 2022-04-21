import * as React from 'react'

import classNames from 'classnames'

import { H2, H4 } from '@sourcegraph/wildcard'

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
    insideTabPanel?: boolean
}

export const PanelContainer: React.FunctionComponent<Props> = ({
    title,
    state,
    loadingContent = <></>,
    emptyContent = <></>,
    populatedContent,
    actionButtons,
    className,
    insideTabPanel,
}) => (
    <div className={classNames(className, styles.panelContainer, 'd-flex', 'flex-column')}>
        {insideTabPanel !== true ? (
            <div className={classNames('d-flex border-bottom', styles.header)}>
                <H4 as={H2} className={styles.headerText}>
                    {title}
                </H4>
                {actionButtons}
            </div>
        ) : (
            <div className={classNames(styles.header, styles.headerInsideTabPanel)}>{actionButtons}</div>
        )}

        <div className={classNames('h-100', styles.content)}>
            {state === 'loading' && loadingContent}
            {state === 'populated' && populatedContent}
            {state === 'empty' && emptyContent}
        </div>
    </div>
)
