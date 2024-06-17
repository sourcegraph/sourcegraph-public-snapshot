import type { FC } from 'react'

import classnames from 'classnames'

import { CodyWebChat } from '@sourcegraph/cody-web'

import styles from './ChatUI.module.scss'

export const ChatUi: FC<{ className?: string }> = props => {
    return <CodyWebChat className={classnames(styles.chat, props.className)} />
}
