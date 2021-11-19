import classNames from 'classnames'
import React from 'react'

import { ErrorAlert, ErrorAlertProps } from '@sourcegraph/branded/src/components/alerts'

import styles from './TreeRowAlert.module.scss'

type TreeRowAlertProps = ErrorAlertProps

export const TreeRowAlert: React.FunctionComponent<TreeRowAlertProps> = ({ className, children, ...rest }) => (
    <ErrorAlert className={classNames(styles.rowAlert, className)} {...rest} />
)
