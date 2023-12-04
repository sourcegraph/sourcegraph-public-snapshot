import React, { type HTMLAttributes } from 'react'

import { ConnectionContainer } from '../../../../components/FilteredConnection/ui'

import styles from './ConnectionPopoverContainer.module.scss'

type ConnectionPopoverContainerProps = HTMLAttributes<HTMLDivElement>

export const ConnectionPopoverContainer: React.FunctionComponent<
    React.PropsWithChildren<ConnectionPopoverContainerProps>
> = ({ className, children, ...rest }) => (
    <ConnectionContainer className={styles.connectionPopoverContent} compact={true} {...rest}>
        {children}
    </ConnectionContainer>
)
