import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/wildcard'

import styles from '../Tree.module.scss'

export const TreeRowAlert: typeof ErrorAlert = ({ className, ...rest }) => (
    <ErrorAlert className={classNames(styles.rowAlert, className)} {...rest} />
)
