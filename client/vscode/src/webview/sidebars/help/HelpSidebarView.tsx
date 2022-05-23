import React, { useEffect, useState } from 'react'

import classNames from 'classnames'

import { version } from '../../../../package.json'
import {
    VSCE_LINK_FEEDBACK,
    VSCE_LINK_ISSUES,
    VSCE_LINK_AUTH,
    VSCE_LINK_TROUBLESHOOT,
    VSCE_SG_LOGOMARK_DARK,
    VSCE_SG_LOGOMARK_LIGHT,
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
            <button
                type="button"
                onClick={() => onHelpItemClick(VSCE_LINK_FEEDBACK, 'Feedback')}
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
            >
                <i className="codicon codicon-github" />
                <span>Give feedback</span>
            </button>
            <button
                type="button"
                onClick={() => onHelpItemClick(VSCE_LINK_ISSUES, 'Issues')}
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
            >
                <i className="codicon codicon-bug" />
                <span>Report issue</span>
            </button>
            <button
                type="button"
                onClick={() => onHelpItemClick(VSCE_LINK_TROUBLESHOOT, 'Troubleshoot')}
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
            >
                <i className="codicon codicon-notebook" />
                <span>Troubleshooting docs</span>
            </button>
            <button
                type="button"
                onClick={() => onHelpItemClick(VSCE_LINK_AUTH('sign-up'), 'Authenticate')}
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
            >
                <img
                    alt="sg-logo"
                    className="codicon"
                    src={isLightTheme ? VSCE_SG_LOGOMARK_DARK : VSCE_SG_LOGOMARK_LIGHT}
                />
                <span>Create an account</span>
            </button>
            <button
                type="button"
                className={classNames(styles.itemContainer, 'btn btn-text text-left')}
                onClick={() => setOpenAuthPanel(previousOpenAuthPanel => !previousOpenAuthPanel)}
            >
                <i className="codicon codicon-account" />
                <span>Authenticate account</span>
            </button>
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
                        <p className="ml-2">Authenticated as {authenticatedUser.username}</p>
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
