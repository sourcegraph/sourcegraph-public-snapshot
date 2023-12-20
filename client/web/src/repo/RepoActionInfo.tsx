import React, { type FC, type ComponentType, type ReactNode } from 'react'

import classNames from 'classnames'
import type { MdiReactIconComponentType } from 'mdi-react'

import { Icon, Text } from '@sourcegraph/wildcard'

import styles from './RepoActionInfo.module.scss'

interface RepoActionInfoProps {
    displayName: string
    icon: ReactNode | MdiReactIconComponentType | ComponentType<{ className?: string | undefined }>
    hideActionLabel?: boolean
    iconClassName?: string
    textClassName?: string
}

export const RepoActionInfo: FC<RepoActionInfoProps> = ({
    displayName,
    icon,
    hideActionLabel = false,
    iconClassName,
    textClassName,
}) => {
    let iconToBeRendered: JSX.Element | undefined | null

    if (typeof icon === 'string') {
        iconToBeRendered = (
            <Icon svgPath={icon} aria-hidden={true} className={classNames(styles.repoActionIcon, iconClassName)} />
        )
    } else if (React.isValidElement(icon) || icon === null) {
        iconToBeRendered = icon
    } else {
        // we cast the icon type here because there's no valid way to check that `icon` is explicitly of type
        // MdiReactIconComponentType | ComponentType<{ className?: string | undefined }>
        // especially considering `React.ReactNode` is a union of all primitive types and normal react
        // components.
        // This is a hack and should be reworked later.
        iconToBeRendered = (
            <Icon
                as={icon as ComponentType}
                aria-hidden={true}
                className={classNames(styles.repoActionIcon, iconClassName)}
            />
        )
    }

    return (
        <>
            {iconToBeRendered}
            {!hideActionLabel && (
                <Text className={classNames(styles.repoActionLabel, textClassName)}>{displayName}</Text>
            )}
        </>
    )
}
