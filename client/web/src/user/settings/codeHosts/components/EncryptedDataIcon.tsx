import React from 'react'

import classNames from 'classnames'
import ShieldCheckIcon from 'mdi-react/ShieldCheckIcon'

import { Icon } from '@sourcegraph/wildcard'

import styles from './EncryptedDataIcon.module.scss'

export const EncryptedDataIcon: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
    ...rest
}) => (
    <Icon
        as={ShieldCheckIcon}
        className={classNames('text-muted', styles.iconInside, className)}
        aria-label="Encrypted Data"
        {...rest}
    />
)
