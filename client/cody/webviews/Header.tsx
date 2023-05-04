import { VSCodeTag } from '@vscode/webview-ui-toolkit/react'

import { CodyBySourcegraphSvg } from '@sourcegraph/cody-ui/src/utils/icons'

import styles from './Header.module.css'

export const Header: React.FunctionComponent = () => (
    <div className={styles.headerContainer}>
        <div className={styles.headerLogo}>
            <CodyBySourcegraphSvg />
        </div>
        <div className={styles.headerTitle}>
            <VSCodeTag className={styles.headerTagBeta}>experimental</VSCodeTag>
        </div>
    </div>
)
