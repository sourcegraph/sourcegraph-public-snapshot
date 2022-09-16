import React, { useEffect, useState } from 'react'

import { VSCodeButton } from '@vscode/webview-ui-toolkit/react'
import classNames from 'classnames'

import { Button, Text } from '@sourcegraph/wildcard'

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

    const onHelpItemClick = async (url: string, item: string): Promise<void> => {
        platformContext.telemetryService.log(`VSCEHelpSidebar${item}Click`)
        await extensionCoreAPI.openLink(url)
    }

    return (
        <div className={classNames(styles.sidebarContainer)}>
            <Button
                as={VSCodeButton}
                onClick={() => onHelpItemClick(VSCE_LINK_FEEDBACK, 'Feedback')}
                className={classNames('p-0 m-0', styles.sidebarViewButton)}
            >
                <i className="codicon codicon-github" slot="start" />
                Give feedback
            </Button>
            <Button
                as={VSCodeButton}
                onClick={() => onHelpItemClick(VSCE_LINK_ISSUES, 'Issues')}
                className={classNames('p-0 m-0', styles.sidebarViewButton)}
            >
                <i className="codicon codicon-bug" slot="start" />
                Report issue
            </Button>
            <Button
                as={VSCodeButton}
                onClick={() => onHelpItemClick(VSCE_LINK_TROUBLESHOOT, 'Troubleshoot')}
                className={classNames('p-0 m-0', styles.sidebarViewButton)}
            >
                <i className="codicon codicon-notebook" slot="start" />
                Troubleshooting docs
            </Button>
            <Button
                as={VSCodeButton}
                onClick={() => onHelpItemClick(VSCE_LINK_SIGNUP, 'Authenticate')}
                className={classNames('p-0 m-0', styles.sidebarViewButton)}
            >
                <img
                    alt="sg-logo"
                    className={classNames(styles.icon, 'codicon')}
                    slot="start"
                    src={isLightTheme ? VSCE_SG_LOGOMARK_DARK : VSCE_SG_LOGOMARK_LIGHT}
                />
                Create an account
            </Button>
            <Button
                as={VSCodeButton}
                onClick={() => setOpenAuthPanel(previousOpenAuthPanel => !previousOpenAuthPanel)}
                className={classNames('p-0 m-0', styles.sidebarViewButton)}
            >
                <i className="codicon codicon-account" slot="start" />
                {authenticatedUser ? `User: ${authenticatedUser.username}` : 'Authenticate account'}
            </Button>
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
                        <Text className="ml-2">Connected to {new URL(instanceURL).hostname}</Text>
                    )}
                </div>
            )}
            <Button as={VSCodeButton} className={classNames('p-0 m-0', styles.sidebarViewButton)}>
                <i className="codicon codicon-calendar" slot="start" />
                Version v{version}
            </Button>
        </div>
    )
}
