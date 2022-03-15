import React from 'react'

import classNames from 'classnames'
import ShieldCheckIcon from 'mdi-react/ShieldCheckIcon'

import { ForwardReferenceComponent, Icon, IconProps } from '@sourcegraph/wildcard'

import styles from './EncryptedDataIcon.module.scss'

export const EncryptedDataIcon = React.forwardRef(({ className, ...rest }, reference) => (
    <Icon
        as={ShieldCheckIcon}
        className={classNames('text-muted', styles.iconInside, className)}
        {...rest}
        ref={reference}
    />
)) as ForwardReferenceComponent<'svg', IconProps>
