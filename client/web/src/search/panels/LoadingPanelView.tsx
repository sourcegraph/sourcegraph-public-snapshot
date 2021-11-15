import classNames from 'classnames'
import * as React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { EmptyPanelContainer } from './EmptyPanelContainer'
import styles from './LoadingPanelView.module.scss'

export const LoadingPanelView: React.FunctionComponent<{ text: string }> = ({ text }) => (
    <EmptyPanelContainer
        className={classNames('d-flex justify-content-center align-items-center', styles.loadingContainer)}
    >
        <LoadingSpinner className="icon-inline" />
        <span className="text-muted">{text}</span>
    </EmptyPanelContainer>
)
