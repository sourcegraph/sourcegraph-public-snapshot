import { VSCodeTag } from '@vscode/webview-ui-toolkit/react'

import { CodyColoredSvg } from '@sourcegraph/cody-ui/src/utils/icons'

import styles from './Header.module.css'

export const Header: React.FunctionComponent = () => (
    <div className={styles.headerContainer}>
        <div className={styles.headerContainerLeft}>
            <div className={styles.headerLogo}>
                <CodyColoredSvg />
            </div>
            <div className={styles.headerTitle}>
                <span className={styles.headerCody}>Cody</span>
                <VSCodeTag className={styles.tagBeta}>experimental</VSCodeTag>
            </div>
        </div>
        <div className={styles.headerContainerRight} />
    </div>
)
