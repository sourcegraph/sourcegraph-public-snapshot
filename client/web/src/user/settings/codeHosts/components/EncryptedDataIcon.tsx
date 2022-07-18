import React from 'react'

import { mdiShieldCheck } from '@mdi/js'
import classNames from 'classnames'

import { Icon } from '@sourcegraph/wildcard'

import styles from './EncryptedDataIcon.module.scss'

export const EncryptedDataIcon: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
    ...rest
}) => (
    <Icon
        className={classNames('text-muted', styles.iconInside, className)}
        aria-label="Encrypted Data"
        svgPath={mdiShieldCheck}
        {...rest}
    />
)
