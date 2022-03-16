import React, { useState } from 'react'

import classNames from 'classnames'

import { WebviewPageProps } from '../../platform/context'
import { AuthSidebarView } from '../auth/AuthSidebarView'

import styles from './HelpSidebarView.module.scss'

interface HelpSidebarViewProps
    extends Pick<WebviewPageProps, 'extensionCoreAPI' | 'platformContext' | 'authenticatedUser' | 'instanceURL'> {}

/**
 * Rendered by sidebar in search-home state when user doesn't have a valid access token.
 */
export const HelpSidebarView: React.FunctionComponent<HelpSidebarViewProps> = ({
    platformContext,
    extensionCoreAPI,
    authenticatedUser,
    instanceURL,
}) => {
    const [hasAccount, setHasAccount] = useState(false)

    const onHelpItemClick = async (url: string): Promise<void> => {
        platformContext.telemetryService.log('VSCESidebarCreateAccount')
        await extensionCoreAPI.openLink(url)
    }

    return (
        // const [state, setState] = useState<'initial' | 'validating' | 'success' | 'failure'>('initial')

        <div className={classNames(styles.sidebarContainer)}>
            <button
                type="button"
                onClick={() =>
                    onHelpItemClick('https://github.com/sourcegraph/sourcegraph/discussions/categories/feedback')
                }
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
            >
                <i className="codicon codicon-github mr-1" />
                <span>Give feedback</span>
            </button>
            <button
                type="button"
                onClick={() =>
                    onHelpItemClick(
                        'https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=&template=bug_report.md&title='
                    )
                }
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
            >
                <i className="codicon codicon-bug mr-1" />
                <span>Report issue</span>
            </button>
            <button
                type="button"
                onClick={() =>
                    onHelpItemClick(
                        'https://docs.sourcegraph.com/admin/how-to/troubleshoot-sg-extension#vs-code-extension'
                    )
                }
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
            >
                <i className="codicon codicon-notebook mr-1" />
                <span>Troubleshooting docs</span>
            </button>
            <button
                type="button"
                onClick={() =>
                    onHelpItemClick(
                        'https://sourcegraph.com/sign-up?editor=vscode&utm_medium=VSCIDE&utm_source=sidebar&utm_campaign=vsce-sign-up&utm_content=sign-up'
                    )
                }
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
            >
                <i className="codicon codicon-reactions mr-1" />
                <span>Create an account</span>
            </button>
            <button
                type="button"
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
                onClick={() => setHasAccount(previousHasAccount => !previousHasAccount)}
            >
                <i className="codicon codicon-sign-in mr-1" />
                <span>Authenticate account</span>
            </button>

            {hasAccount && (
                <div className="ml-3 mt-1">
                    {!authenticatedUser ? (
                        <AuthSidebarView
                            instanceURL={instanceURL}
                            extensionCoreAPI={extensionCoreAPI}
                            platformContext={platformContext}
                            stateStatus=""
                        />
                    ) : (
                        <p className="ml-2">Logged in as {authenticatedUser.displayName}</p>
                    )}
                </div>
            )}
        </div>
    )
}
