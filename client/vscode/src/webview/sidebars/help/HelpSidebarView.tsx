import React, { useEffect, useState } from 'react'
import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'
import classNames from 'classnames'

import { version } from '../../../../package.json'
import {
    VSCE_LINK_FEEDBACK,
    VSCE_LINK_ISSUES,
    VSCE_LINK_TROUBLESHOOT,
    VSCE_SG_LOGOMARK_DARK,
    VSCE_SG_LOGOMARK_LIGHT,
    VSCE_LINK_SIGNUP,
} from '../../../common/links'
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
    const [openAuthPanel, setOpenAuthPanel] = useState(false)
    const [isLightTheme, setIsLightTheme] = useState<boolean | undefined>(undefined)

    useEffect(() => {
        if (isLightTheme === undefined) {
            extensionCoreAPI.getEditorTheme
                .then(theme => {
                    console.log(theme)
                    setIsLightTheme(theme === 'Light')
                })
                .catch(error => {
                    console.log(error)
                    setIsLightTheme(false)
                })
        }
    }, [extensionCoreAPI.getEditorTheme, isLightTheme])
    console.log(isLightTheme)

    const onHelpItemClick = async (url: string, item: string): Promise<void> => {
        platformContext.telemetryService.log(`VSCEHelpSidebar${item}Click`)
        await extensionCoreAPI.openLink(url)
    }

    return (
        <div className={classNames(styles.sidebarContainer)}>
            <VSCodeButton
                onClick={() => onHelpItemClick(VSCE_LINK_FEEDBACK, 'Feedback')}
                className="btn btn-text text-left p-0 m-0"
            >
                <i className="codicon codicon-github" slot="start" />
                Give feedback
            </VSCodeButton>
            <VSCodeButton
                onClick={() => onHelpItemClick(VSCE_LINK_ISSUES, 'Issues')}
                className="btn btn-text text-left p-0 m-0"
            >
                <i className="codicon codicon-bug" slot="start" />
                Report issue
            </VSCodeButton>
            <VSCodeButton
                onClick={() => onHelpItemClick(VSCE_LINK_TROUBLESHOOT, 'Troubleshoot')}
                className="btn btn-text text-left p-0 m-0"
            >
                <i className="codicon codicon-notebook" slot="start" />
                Troubleshooting docs
            </VSCodeButton>
            <VSCodeButton
                onClick={() => onHelpItemClick(VSCE_LINK_SIGNUP, 'Authenticate')}
                className="btn btn-text text-left p-0 m-0"
            >
                <img
                    alt="sg-logo"
                    className={classNames(styles.icon, 'codicon')}
                    slot="start"
                    src={isLightTheme ? VSCE_SG_LOGOMARK_DARK : VSCE_SG_LOGOMARK_LIGHT}
                />
                Create an account
            </VSCodeButton>
            <VSCodeButton
                onClick={() => setOpenAuthPanel(previousOpenAuthPanel => !previousOpenAuthPanel)}
                className="btn btn-text text-left p-0 m-0"
            >
                <i className="codicon codicon-account" slot="start" />
                {authenticatedUser ? `User: ${authenticatedUser.username}` : 'Authenticate account'}
            </VSCodeButton>
            {openAuthPanel && (
                <div className="ml-3 mt-1">
                    {!authenticatedUser ? (
                        <AuthSidebarView
                            instanceURL={instanceURL}
                            extensionCoreAPI={extensionCoreAPI}
                            platformContext={platformContext}
                            authenticatedUser={authenticatedUser}
                        />
                    ) : (
                        <p className="ml-2">Connected to {new URL(instanceURL).hostname}</p>
                    )}
                </div>
            )}
            <VSCodeButton className="btn btn-text text-left p-0 m-0">
                <i className="codicon codicon-calendar" slot="start" />
                Version v{version}
            </VSCodeButton>
        </div>
    )
}
