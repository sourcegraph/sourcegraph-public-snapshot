import React, { useMemo, useState } from 'react'

import classNames from 'classnames'

import { version } from '../../../../package.json'
import { WebviewPageProps } from '../../platform/context'
import { AuthSidebarView } from '../auth/AuthSidebarView'

import styles from './HelpSidebarView.module.scss'

interface HelpSidebarViewProps
    extends Pick<WebviewPageProps, 'extensionCoreAPI' | 'platformContext' | 'authenticatedUser' | 'instanceURL'> {}

/**
 * Rendered by sidebar in search-home state when user doesn't have a valid access token.
 */
export const HelpSidebarView: React.FunctionComponent<React.PropsWithChildren<HelpSidebarViewProps>> = ({
    platformContext,
    extensionCoreAPI,
    authenticatedUser,
    instanceURL,
}) => {
    const [hasAccount, setHasAccount] = useState(false)

    const hostname = useMemo(() => new URL(instanceURL).hostname, [instanceURL])

    const onHelpItemClick = async (url: string, item: string): Promise<void> => {
        platformContext.telemetryService.log(`VSCEHelpSidebar${item}Click`)
        await extensionCoreAPI.openLink(url)
    }

    return (
        <div className={classNames(styles.sidebarContainer)}>
            <button
                type="button"
                onClick={() =>
                    onHelpItemClick('https://github.com/sourcegraph/sourcegraph/discussions/34821', 'Feedback')
                }
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
            >
                <i className="codicon codicon-github" />
                <span>Give feedback</span>
            </button>
            <button
                type="button"
                onClick={() =>
                    onHelpItemClick(
                        'https://github.com/sourcegraph/sourcegraph/issues/new?labels=team/integrations,vscode-extension&title=VSCode+Bug+report:+&projects=Integrations%20Project%20Board',
                        'Issues'
                    )
                }
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
            >
                <i className="codicon codicon-bug" />
                <span>Report issue</span>
            </button>
            <button
                type="button"
                onClick={() =>
                    onHelpItemClick(
                        'https://docs.sourcegraph.com/admin/how-to/troubleshoot-sg-extension#vs-code-extension',
                        'Troubleshoot'
                    )
                }
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
            >
                <i className="codicon codicon-notebook" />
                <span>Troubleshooting docs</span>
            </button>
            <button
                type="button"
                onClick={() =>
                    onHelpItemClick(
                        'https://sourcegraph.com/sign-up?editor=vscode&utm_medium=VSCODE&utm_source=sidebar&utm_campaign=vsce-sign-up&utm_content=sign-up',
                        'Authenticate'
                    )
                }
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
            >
                <img
                    alt="sg-logo"
                    className="codicon"
                    src="https://raw.githubusercontent.com/sourcegraph/sourcegraph/fd431743e811ba756490e5e7bd88aa2362b6453e/client/vscode/images/logomark_light.svg"
                />
                <span>Create an account</span>
            </button>
            <button
                type="button"
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
                onClick={() => setHasAccount(previousHasAccount => !previousHasAccount)}
            >
                <i className="codicon codicon-account" />
                <span>Authenticate account</span>
            </button>
            {hasAccount && (
                <div className="ml-3 mt-1">
                    {!authenticatedUser ? (
                        <AuthSidebarView
                            instanceURL={instanceURL}
                            extensionCoreAPI={extensionCoreAPI}
                            platformContext={platformContext}
                            authenticatedUser={authenticatedUser}
                        />
                    ) : (
                        <p className="ml-2">
                            Connected to {hostname} as {authenticatedUser.displayName}
                        </p>
                    )}
                </div>
            )}
            <button type="button" className={classNames(styles.itemContainer, 'btn btn-text text-left')}>
                <i className="codicon codicon-calendar" />
                <span>Version v{version}</span>
            </button>
        </div>
    )
}
