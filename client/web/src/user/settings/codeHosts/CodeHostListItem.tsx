import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from './CodeHostListItem.module.scss'

type CodeHostListItemProps = HTMLAttributes<HTMLLIElement>

export const CodeHostListItem: React.FunctionComponent<React.PropsWithChildren<CodeHostListItemProps>> = ({
    children,
    ...rest
}) => (
    <li className={classNames('list-group-item', styles.codeHostItem)} {...rest}>
        {children}
    </li>
)
