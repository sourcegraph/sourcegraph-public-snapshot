import type { FC } from 'react'

import classnames from 'classnames'

import { CodyWebChat } from '@sourcegraph/cody-web'

import '@sourcegraph/cody-web/dist/style.css'

import styles from './ChatUI.module.scss'

export const ChatUi: FC<{ className?: string }> = props => (
    <CodyWebChat className={classnames(styles.chat, props.className)} />
)
