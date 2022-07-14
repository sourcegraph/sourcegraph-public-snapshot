import React from 'react'

import { TabbedPanelContent } from '@sourcegraph/branded/src/components/panel/TabbedPanelContent'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'
import { Panel } from '@sourcegraph/wildcard'

import { LayoutProps } from './Layout'
import { PageRoutes } from './routes.constants'
import { useThemeProps } from './theme'
import { parseBrowserRepoURL } from './util/url'

import styles from './GlobalPanel.module.scss'

export const GlobalPanel: React.FunctionComponent<React.PropsWithChildren<LayoutProps>> = props => {
    const themeProps = useThemeProps()

    const { viewState } = parseQueryAndHash(props.location.search, props.location.hash)
    if (!viewState || props.location.pathname === PageRoutes.SignIn) {
        return null
    }

    return (
        <Panel className={styles.panel} position="bottom" defaultSize={350} storageKey="panel-size">
            <TabbedPanelContent
                {...props}
                {...themeProps}
                repoName={`git://${parseBrowserRepoURL(props.location.pathname).repoName}`}
                fetchHighlightedFileLineRanges={props.fetchHighlightedFileLineRanges}
            />
        </Panel>
    )
}
