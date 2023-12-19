import React, { type FC, type ComponentType } from 'react'

import classNames from 'classnames'
import type { MdiReactIconComponentType } from 'mdi-react'

import { Icon, Text } from '@sourcegraph/wildcard'

import styles from './RepoActionInfo.module.scss'

interface RepoActionInfoProps {
    displayName: string
    icon: string | MdiReactIconComponentType | ComponentType<{ className?: string | undefined }>
    hideActionLabel?: boolean
    iconClassName?: string
}

export const RepoActionInfo: FC<RepoActionInfoProps> = ({
    displayName,
    icon,
    hideActionLabel = false,
    iconClassName,
}) => {
    let iconToBeRendered: JSX.Element | undefined

    if (typeof icon === 'string') {
        iconToBeRendered = (
            <Icon svgPath={icon} aria-hidden={true} className={classNames(styles.repoActionIcon, iconClassName)} />
        )
        // } else if (typeof icon === 'function') {
        //     iconToBeRendered = (
        //         <Icon as={icon} aria-hidden={true} className={classNames(styles.repoActionIcon, iconClassName)} />
        //     )
    } else {
        // console.log(icon, { icon })
        // iconToBeRendered = <span>hell</span>
        iconToBeRendered = (
            <Icon as={icon} aria-hidden={true} className={classNames(styles.repoActionIcon, iconClassName)} />
        )
    }

    return (
        <>
            {iconToBeRendered}
            {!hideActionLabel && <Text className={styles.repoActionLabel}>{displayName}</Text>}
        </>
    )
}
