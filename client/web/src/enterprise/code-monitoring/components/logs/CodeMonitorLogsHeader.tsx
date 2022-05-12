import React from 'react'

import classNames from 'classnames'

import { Typography } from '@sourcegraph/wildcard'

import styles from './CodeMonitorLogsHeader.module.scss'

export const CodeMonitorLogsHeader: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <>
        <Typography.H5 className={classNames(styles.nameColumn, 'text-uppercase text-nowrap')}>
            Monitor name
        </Typography.H5>
        <Typography.H5 className="text-uppercase text-nowrap">Last run</Typography.H5>
    </>
)
