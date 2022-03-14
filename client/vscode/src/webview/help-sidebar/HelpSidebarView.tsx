import classNames from 'classnames'
import React from 'react'

import styles from './HelpSidebarView.module.scss'

interface HelpSidebarViewProps {}

/**
 * Rendered by sidebar in search-home state when user doesn't have a valid access token.
 */
export const HelpSidebarView: React.FunctionComponent<HelpSidebarViewProps> = () => (
    // const [state, setState] = useState<'initial' | 'validating' | 'success' | 'failure'>('initial')

    <div className={classNames(styles.sidebarContainer)}>
        <div className="icon">
            <a
                className={classNames(styles.itemContainer)}
                href="https://github.com/sourcegraph/sourcegraph/discussions/categories/feedback"
            >
                <i className="codicon codicon-github" /> Give feedback
            </a>
        </div>
        <div className="icon">
            <a
                className={classNames(styles.itemContainer)}
                href="https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=&template=bug_report.md&title="
            >
                <i className="codicon codicon-issues" /> Report issue
            </a>
        </div>
        <div className="icon">
            <a
                className={classNames(styles.itemContainer)}
                href="https://docs.sourcegraph.com/admin/how-to/troubleshoot-sg-extension#vs-code-extension"
            >
                <i className="codicon codicon-issues" /> Troubleshooting docs
            </a>
        </div>
        <div className="icon">
            <a
                className={classNames(styles.itemContainer)}
                href="https://sourcegraph.com/sign-up?editor=vscode&utm_medium=VSCIDE&utm_source=sidebar&utm_campaign=vsce-sign-up&utm_content=sign-up"
            >
                <i className="codicon codicon-issues" /> Create an account
            </a>
        </div>
        <div className="icon">
            <a
                className={classNames(styles.itemContainer)}
                href="https://docs.sourcegraph.com/admin/how-to/troubleshoot-sg-extension#vs-code-extension"
            >
                <i className="codicon codicon-issues" /> Authenticate account
            </a>
        </div>
    </div>
)
