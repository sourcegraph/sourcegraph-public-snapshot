import { SourcegraphLogo } from '@sourcegraph/branded/src/components/SourcegraphLogo'
import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'
import { ConsoleUserData } from '../model'
import styles from './ConsoleLayout.module.scss'

export const ConsoleLayout: React.FunctionComponent<{ data: ConsoleUserData; children: React.ReactNode }> = ({
    data,
    children,
}) => (
    <div className={styles.layout}>
        <header className={classNames(styles.header, 'py-2', 'border-bottom')}>
            <div className="container">
                <h1 className="mb-1">
                    <Link to="/">
                        <SourcegraphLogo className={classNames(styles.logo, 'mr-2')} />
                        <span className={classNames(styles.logoText)}>Console</span>
                    </Link>
                </h1>
            </div>
        </header>
        <main className="container py-3">{children}</main>
    </div>
)
