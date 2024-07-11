import type { FC } from 'react'

import classNames from 'classnames'

import styles from './Skeleton.module.scss'

export const Skeleton: FC = props => (
    <div className={styles.lines}>
        <div className={classNames(styles.line, styles.lineTitle)} />
        <div className={styles.line} />
        <div className={styles.line} />
        <div className={styles.line} />
    </div>
)
