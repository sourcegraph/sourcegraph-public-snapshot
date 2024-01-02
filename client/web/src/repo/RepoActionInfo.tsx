import type { ReactNode, FC } from 'react'

import classNames from 'classnames'

import { Text } from '@sourcegraph/wildcard'

import styles from './RepoActionInfo.module.scss'

interface RepoActionInfoProps {
    displayName: string
    icon: ReactNode
    className?: string
}

export const RepoActionInfo: FC<RepoActionInfoProps> = ({ displayName, icon, className }) => (
    <>
        {icon}
        <Text className={classNames(styles.repoActionLabel, className)}>{displayName}</Text>
    </>
)
