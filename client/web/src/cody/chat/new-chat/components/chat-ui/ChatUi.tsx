import type { FC } from 'react'

import classnames from 'classnames'
import { CodyWebChat } from 'cody-web-experimental'

import 'cody-web-experimental/dist/style.css'

import styles from './ChatUI.module.scss'

export const ChatUi: FC<{ className?: string }> = props => (
    <CodyWebChat className={classnames(styles.chat, props.className)} />
)
