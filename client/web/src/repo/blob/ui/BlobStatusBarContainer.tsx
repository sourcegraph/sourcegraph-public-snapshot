import classNames from 'classnames'
import React from 'react'

import styles from './BlobStatusBarContainer.module.scss'

interface BlobStatusBarContainerProps {
    className?: string
}

export const BlobStatusBarContainer: React.FunctionComponent<BlobStatusBarContainerProps> = ({
    children,
    className,
}) => <div className={classNames(className, styles.blobStatusBarContainer, styles.content)}>{children}</div>
