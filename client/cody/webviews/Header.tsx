import { VSCodeTag } from '@vscode/webview-ui-toolkit/react'

import { CodyColoredSvg } from '@sourcegraph/cody-ui/src/utils/icons'

import styles from './Header.module.css'

export const Header: React.FunctionComponent = () => (
    <div className={styles.container}>
        <div className={styles.logo}>
            <CodyColoredSvg />
        </div>
        <VSCodeTag className={styles.tag}>experimental</VSCodeTag>
    </div>
)
