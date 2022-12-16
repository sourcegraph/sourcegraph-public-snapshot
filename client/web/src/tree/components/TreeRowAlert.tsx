import React from 'react'

import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/wildcard'

import styles from '../Tree.module.scss'

type TreeRowAlertProps = ErrorAlertProps

export const TreeRowAlert: React.FunctionComponent<React.PropsWithChildren<TreeRowAlertProps>> = ({
    className,
    children,
    ...rest
}) => <ErrorAlert className={classNames(styles.rowAlert, className)} {...rest} />
