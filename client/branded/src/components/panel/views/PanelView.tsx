import * as H from 'history'
import React from 'react'
import { Observable } from 'rxjs'
import { FetchFileParameters } from '../../../../../shared/src/components/CodeExcerpt'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { renderMarkdown } from '../../../../../shared/src/util/markdown'
import { EmptyPanelView } from './EmptyPanelView'
import { HierarchicalLocationsView } from './HierarchicalLocationsView'
import { VersionContextProps } from '../../../../../shared/src/search/util'
import { PanelViewWithComponent } from '../Panel'

interface Props extends ExtensionsControllerProps, SettingsCascadeProps, VersionContextProps {
    panelView: PanelViewWithComponent
    repoName?: string
    history: H.History
    location: H.Location
    isLightTheme: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

/**
 * A panel view contributed by an extension using {@link sourcegraph.app.createPanelView}.
 */
export const PanelView = React.memo<Props>(props => {
    return (
        <div className="panel__tabs-content panel__tabs-content--scroll">
            {props.panelView.content && (
                <div className="px-2 pt-2">
                    <Markdown dangerousInnerHTML={renderMarkdown(props.panelView.content)} history={props.history} />
                </div>
            )}
            {props.panelView.reactElement}
            {props.panelView.locationProvider && props.repoName && (
                <HierarchicalLocationsView
                    location={props.location}
                    locations={props.panelView.locationProvider}
                    defaultGroup={props.repoName}
                    isLightTheme={props.isLightTheme}
                    fetchHighlightedFileLineRanges={props.fetchHighlightedFileLineRanges}
                    extensionsController={props.extensionsController}
                    settingsCascade={props.settingsCascade}
                    versionContext={props.versionContext}
                />
            )}
            {!props.panelView.content && !props.panelView.reactElement && !props.panelView.locationProvider && (
                <EmptyPanelView className="mt-3" />
            )}
        </div>
    )
})
