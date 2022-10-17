import React from 'react'

import classNames from 'classnames'

import styles from './OrDivider.module.scss'

interface Props {
    className?: string
}

export const OrDivider: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ className }) => (
    <div className={classNames(className, 'd-flex align-items-center')}>
        <div className={classNames('w-100', styles.border)} />
        <small className="px-2 text-muted ">OR</small>
        <div className={classNames('w-100', styles.border)} />
    </div>
)
