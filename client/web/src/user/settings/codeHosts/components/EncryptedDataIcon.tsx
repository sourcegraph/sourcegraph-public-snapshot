import classNames from 'classnames'
import ShieldCheckIcon from 'mdi-react/ShieldCheckIcon'
import React from 'react'

import { ForwardReferenceComponent, Icon, IconProps } from '@sourcegraph/wildcard'

import styles from './EncryptedDataIcon.module.scss'

export const EncryptedDataIcon = React.forwardRef<SVGElement, IconProps>(({ className, ...rest }, reference) => (
    <Icon
        className={classNames('text-muted', styles.iconInside, className)}
        data-tooltip="Data will be encrypted and will not be visible again."
        as={ShieldCheckIcon}
        {...rest}
        ref={reference}
    />
)) as ForwardReferenceComponent<'svg', IconProps>
