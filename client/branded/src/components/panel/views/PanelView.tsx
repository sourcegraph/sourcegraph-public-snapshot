import React from 'react'

import * as H from 'history'
import { Observable } from 'rxjs'

import { renderMarkdown } from '@sourcegraph/common'
import { FetchFileParameters } from '@sourcegraph/search-ui'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { PanelViewWithComponent } from '../TabbedPanelContent'

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
export const PanelView = React.memo<Props>(props => {
    const panelView = (
        <>
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
                    maxLocationResults={props.panelView.maxLocationResults}
                    defaultGroup={props.repoName}
                    isLightTheme={props.isLightTheme}
                    fetchHighlightedFileLineRanges={props.fetchHighlightedFileLineRanges}
                    extensionsController={props.extensionsController}
                    settingsCascade={props.settingsCascade}
                    telemetryService={props.telemetryService}
                    onSelectLocation={(): void =>
                        props.telemetryService.log('ReferencePanelResultsClicked', { action: 'click' })
                    }
                />
            )}
            {!props.panelView.content && !props.panelView.reactElement && !props.panelView.locationProvider && (
                <EmptyPanelView className="mt-3" />
            )}
        </>
    )

    if (props.panelView.noWrapper) {
        return <>{panelView}</>
    }

    return <div className={styles.panelView}>{panelView}</div>
})
