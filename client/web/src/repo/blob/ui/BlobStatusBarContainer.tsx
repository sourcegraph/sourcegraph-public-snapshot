import React from 'react'

import classNames from 'classnames'

import styles from './BlobStatusBarContainer.module.scss'

interface BlobStatusBarContainerProps {
    className?: string
}

export const BlobStatusBarContainer: React.FunctionComponent<React.PropsWithChildren<BlobStatusBarContainerProps>> = ({
    children,
    className,
}) => <div className={classNames(className, styles.blobStatusBarContainer, styles.content)}>{children}</div>
