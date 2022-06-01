import React from 'react'

import classNames from 'classnames'

import { H5 } from '@sourcegraph/wildcard'

import styles from './CodeMonitorLogsHeader.module.scss'

export const CodeMonitorLogsHeader: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <>
        <H5 className={classNames(styles.nameColumn, 'text-uppercase text-nowrap')}>Monitor name</H5>
        <H5 className="text-uppercase text-nowrap">Last run</H5>
    </>
)
