import React, { useEffect, useRef } from 'react'

import { useLocation } from 'react-router'

import { TabbedPanelContent } from '@sourcegraph/branded/src/components/panel/TabbedPanelContent'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'
import { Panel } from '@sourcegraph/wildcard'

import { LayoutProps } from './Layout'
import { PageRoutes } from './routes.constants'
import { useThemeProps } from './theme'
import { parseBrowserRepoURL } from './util/url'

import styles from './GlobalPanel.module.scss'

export interface GlobalPanelLocationState {
    focus?: boolean
}

export const GlobalPanel: React.FunctionComponent<React.PropsWithChildren<LayoutProps>> = props => {
    const { search, hash, pathname, state } = useLocation<GlobalPanelLocationState | undefined>()
    const panelRef = useRef<HTMLDivElement>(null)
    const themeProps = useThemeProps()

    const { viewState } = parseQueryAndHash(search, hash)

    useEffect(() => {
        const panelElement = panelRef.current
        if (state?.focus && panelElement) {
            console.log(panelElement)
            panelElement.focus()
        }
    }, [state])

    if (!viewState || pathname === PageRoutes.SignIn) {
        return null
    }

    return (
        <Panel ref={panelRef} className={styles.panel} position="bottom" defaultSize={350} storageKey="panel-size">
            <TabbedPanelContent
                {...props}
                {...themeProps}
                repoName={`git://${parseBrowserRepoURL(pathname).repoName}`}
                fetchHighlightedFileLineRanges={props.fetchHighlightedFileLineRanges}
            />
        </Panel>
    )
}
