import * as H from 'history'
import React from 'react'
import { Observable } from 'rxjs'

import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'

import { PanelViewWithComponent } from '../Panel'

import { EmptyPanelView } from './EmptyPanelView'
import { HierarchicalLocationsView } from './HierarchicalLocationsView'
import styles from './PanelView.module.scss'

interface Props extends ExtensionsControllerProps, SettingsCascadeProps, TelemetryProps {
    panelView: PanelViewWithComponent
    repoName?: string
    location: H.Location
    isLightTheme: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

/**
 * A panel view contributed by an extension using {@link sourcegraph.app.createPanelView}.
 */
export const PanelView = React.memo<Props>(props => (
    <div className={styles.panelView}>
        {props.panelView.content && (
            <div className="px-2 pt-2">
                <Markdown dangerousInnerHTML={renderMarkdown(props.panelView.content)} />
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
                telemetryService={props.telemetryService}
            />
        )}
        {!props.panelView.content && !props.panelView.reactElement && !props.panelView.locationProvider && (
            <EmptyPanelView className="mt-3" />
        )}
    </div>
))
