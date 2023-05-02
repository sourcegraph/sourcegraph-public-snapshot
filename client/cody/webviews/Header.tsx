import { VSCodeTag } from '@vscode/webview-ui-toolkit/react'

import { CodyColoredSvg, SourcegraphLogoMuted } from '@sourcegraph/cody-ui/src/utils/icons'

import styles from './Header.module.css'

export const Header: React.FunctionComponent = () => (
    <div className={styles.headerContainer}>
        <div className={styles.headerContainerLeft}>
            <div className={styles.headerLogo}>
                <CodyColoredSvg />
            </div>
            <div className={styles.headerTitle}>
                <VSCodeTag className={styles.headerTagBeta}>experimental</VSCodeTag>
            </div>
        </div>
        <div className={styles.headerContainerRight}>
            <SourcegraphLogoMuted />
        </div>
    </div>
)
