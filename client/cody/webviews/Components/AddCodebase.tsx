import { useState } from 'react'

import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'
import classNames from 'classnames'

import { AddRepoToAppPanelProps } from '@sourcegraph/cody-ui/src/Chat'

import styles from './AddCodebase.module.css'

export const AddRepoToAppPanel: React.FunctionComponent<AddRepoToAppPanelProps> = ({ repoName, onClick }) => {
    const [isClicked, setIsClicked] = useState(false)
    if (isClicked || !repoName) {
        return <></>
    }
    return (
        <div className={classNames(styles.section, styles.codyGradient)}>
            <p className={styles.title}>Add {repoName}</p>
            <p>To use this repository, you'll need to add it to Cody App.</p>
            <VSCodeButton
                type="button"
                onClick={() => {
                    onClick(repoName || '')
                    setIsClicked(true)
                }}
            >
                Add Repo to Cody App
            </VSCodeButton>
        </div>
    )
}
