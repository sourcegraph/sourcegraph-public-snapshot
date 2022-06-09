import React from 'react'

import classNames from 'classnames'
import ShieldCheckIcon from 'mdi-react/ShieldCheckIcon'

import { ForwardReferenceComponent, Icon, IconProps } from '@sourcegraph/wildcard'

import styles from './EncryptedDataIcon.module.scss'

export const EncryptedDataIcon = React.forwardRef(({ className, ...rest }, reference) => (
    <Icon
        role="img"
        as={ShieldCheckIcon}
        className={classNames('text-muted', styles.iconInside, className)}
        {...rest}
        ref={reference}
        aria-label="Encrypted Data"
    />
)) as ForwardReferenceComponent<'svg', IconProps>
