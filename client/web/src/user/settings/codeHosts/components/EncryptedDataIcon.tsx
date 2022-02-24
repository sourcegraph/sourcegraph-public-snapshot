import classNames from 'classnames'
import { MdiReactIconProps } from 'mdi-react'
import ShieldCheckIcon from 'mdi-react/ShieldCheckIcon'
import React from 'react'

import { Icon } from '@sourcegraph/wildcard'

import styles from './EncryptedDataIcon.module.scss'

type EncryptedDataIconProps = MdiReactIconProps

export const EncryptedDataIcon: React.FunctionComponent<EncryptedDataIconProps> = ({ className, ref, ...rest }) => (
    <Icon
        className={classNames('text-muted', styles.iconInside, className)}
        data-tooltip="Data will be encrypted and will not be visible again."
        {...rest}
        as={ShieldCheckIcon}
    />
)
