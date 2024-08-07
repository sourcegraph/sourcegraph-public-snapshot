import type { FC } from 'react'

import classnames from 'classnames'

import { CodyWebPanel } from '@sourcegraph/cody-web'

import '@sourcegraph/cody-web/dist/style.css'

import styles from './ChatUI.module.scss'

export const ChatUi: FC<{ className?: string }> = props => (
    <CodyWebPanel className={classnames(styles.chat, props.className)} />
)
