import classNames from 'classnames'
import { MdiReactIconProps } from 'mdi-react'
import ShieldCheckIcon from 'mdi-react/ShieldCheckIcon'
import React from 'react'

import { ForwardReferenceComponent, Icon } from '@sourcegraph/wildcard'

import styles from './EncryptedDataIcon.module.scss'

type EncryptedDataIconProps = Omit<MdiReactIconProps, 'size'>

export const EncryptedDataIcon = React.forwardRef(({ className, ...rest }, reference) => (
    <Icon
        as={ShieldCheckIcon}
        className={classNames('icon-inline text-muted', styles.iconInside, className)}
        {...rest}
        ref={reference}
    />
)) as ForwardReferenceComponent<'svg', EncryptedDataIconProps>
