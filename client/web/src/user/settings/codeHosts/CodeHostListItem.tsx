import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'

import styles from './CodeHostListItem.module.scss'

type CodeHostListItemProps = HTMLAttributes<HTMLLIElement>

export const CodeHostListItem: React.FunctionComponent<CodeHostListItemProps> = ({ children, ...rest }) => (
    <li className={classNames('list-group-item', styles.codeHostItem)} {...rest}>
        {children}
    </li>
)
