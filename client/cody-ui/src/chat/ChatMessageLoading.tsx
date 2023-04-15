import React from 'react'

import classNames from 'classnames'

import styles from './ChatMessageLoading.module.css'

export const ChatMessageLoading: React.FunctionComponent<{ bubbleLoaderDotClassName?: string }> = ({
    bubbleLoaderDotClassName,
}) => (
    <div className={styles.bubbleLoader}>
        <div className={classNames(styles.bubbleLoaderDot, bubbleLoaderDotClassName)} />
        <div className={classNames(styles.bubbleLoaderDot, bubbleLoaderDotClassName)} />
        <div className={classNames(styles.bubbleLoaderDot, bubbleLoaderDotClassName)} />
    </div>
)
