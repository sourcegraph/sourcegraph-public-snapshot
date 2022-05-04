import React from 'react'

import classNames from 'classnames'

import styles from './ExecutionMetaInformation.module.scss'

export interface ExecutionMetaInformationProps {
    image: string
    commands: string[]
    root: string
}

export const ExecutionMetaInformation: React.FunctionComponent<
    React.PropsWithChildren<ExecutionMetaInformationProps>
> = ({ image, commands, root }) => (
    <div className="pt-3">
        <div className={classNames(styles.dockerCommandSpec, 'py-2 border-top pl-2')}>
            <strong className={styles.header}>Image</strong>
            <div>{image}</div>
        </div>
        <div className={classNames(styles.dockerCommandSpec, 'py-2 border-top pl-2')}>
            <strong className={styles.header}>Commands</strong>
            <div>
                <code>{commands.join(' ')}</code>
            </div>
        </div>
        <div className={classNames(styles.dockerCommandSpec, 'py-2 border-top pl-2')}>
            <strong className={styles.header}>Root</strong>
            <div>/{root}</div>
        </div>
    </div>
)
