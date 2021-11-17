import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'

import styles from './UserSettingReposContainer.module.scss'

type UserSettingReposContainerProps = HTMLAttributes<HTMLElement>

export const UserSettingReposContainer: React.FunctionComponent<UserSettingReposContainerProps> = ({
    children,
    className,
    ...rest
}) => (
    <div className={classNames(className, styles.userSettingsRepos)} {...rest}>
        {children}
    </div>
)
