import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from './UserSettingReposContainer.module.scss'

type UserSettingReposContainerProps = HTMLAttributes<HTMLElement>

export const UserSettingReposContainer: React.FunctionComponent<
    React.PropsWithChildren<UserSettingReposContainerProps>
> = ({ children, className, ...rest }) => (
    <div className={classNames(className, styles.userSettingsRepos)} {...rest}>
        {children}
    </div>
)
