import React, { Suspense } from 'react'

import type { TabbedPanelContentProps } from '@sourcegraph/branded/src/components/panel/TabbedPanelContent'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'
import { LoadingSpinner, Panel } from '@sourcegraph/wildcard'

import { LayoutProps } from './Layout'
import { PageRoutes } from './routes.constants'
import { useThemeProps } from './theme'
import { parseBrowserRepoURL } from './util/url'

import styles from './GlobalPanel.module.scss'

const TabbedPanelContent = lazyComponent<TabbedPanelContentProps, 'TabbedPanelContent'>(
    () => import('@sourcegraph/branded/src/components/panel/TabbedPanelContent'),
    'TabbedPanelContent'
)

export const GlobalPanel: React.FunctionComponent<React.PropsWithChildren<LayoutProps>> = props => {
    const themeProps = useThemeProps()

    const { viewState } = parseQueryAndHash(props.location.search, props.location.hash)
    if (!viewState || props.location.pathname === PageRoutes.SignIn) {
        return null
    }

    return (
        <Panel className={styles.panel} position="bottom" defaultSize={350} storageKey="panel-size">
            <Suspense
                fallback={
                    <div className="d-flex justify-content-center">
                        <LoadingSpinner className="m-3" />
                    </div>
                }
            >
                <TabbedPanelContent
                    {...props}
                    {...themeProps}
                    repoName={`git://${parseBrowserRepoURL(props.location.pathname).repoName}`}
                    fetchHighlightedFileLineRanges={props.fetchHighlightedFileLineRanges}
                />
            </Suspense>
        </Panel>
    )
}
