import { VSCodeTag } from '@vscode/webview-ui-toolkit/react'

import { CodyColoredSvg } from '@sourcegraph/cody-ui/src/utils/icons'

import styles from './Header.module.css'

export const Header: React.FunctionComponent<{ endpoint: string | null }> = ({ endpoint }) => (
    <div className={styles.container}>
        <div className={styles.logo} title={endpoint || 'Cody'}>
            <CodyColoredSvg />
        </div>
        <VSCodeTag className={styles.tag}>beta</VSCodeTag>
    </div>
)
