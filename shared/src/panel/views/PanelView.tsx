import { Location } from '@sourcegraph/extension-api-types'
import H from 'history'
import React from 'react'
import { Observable } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { FetchFileCtx } from '../../components/CodeExcerpt'
import { Markdown } from '../../components/Markdown'
import { ExtensionsControllerProps } from '../../extensions/controller'
import { SettingsCascadeProps } from '../../settings/settings'
import { ErrorLike, isErrorLike } from '../../util/errors'
import { createLinkClickHandler } from '../../util/linkClickHandler'
import { renderMarkdown } from '../../util/markdown'
import { EmptyPanelView } from './EmptyPanelView'
import { FileLocationsError } from './FileLocations'
import { HierarchicalLocationsView } from './HierarchicalLocationsView'

interface T extends Pick<sourcegraph.PanelView, 'title' | 'content' | 'priority'> {
    id: string
    locationsOrCustom:
        | { locations: { results?: Location[]; loading: boolean } | ErrorLike }
        | { custom: React.ReactFragment }
}

interface Props extends ExtensionsControllerProps, SettingsCascadeProps {
    panelView: T
    repoName?: string
    history: H.History
    location: H.Location
    isLightTheme: boolean
    fetchHighlightedFileLines: (ctx: FetchFileCtx, force?: boolean) => Observable<string[]>
}

/**
 * A panel view contributed by an extension using {@link sourcegraph.app.createPanelView}.
 */
export class PanelView extends React.PureComponent<Props, {}> {
    public render(): JSX.Element | null {
        return (
            <div
                onClick={createLinkClickHandler(this.props.history)}
                className="panel__tabs-content panel__tabs-content--scroll"
            >
                {this.props.panelView.content && (
                    <div className="px-2 pt-2">
                        <Markdown dangerousInnerHTML={renderMarkdown(this.props.panelView.content)} />
                    </div>
                )}

                {'locations' in this.props.panelView.locationsOrCustom ? (
                    isErrorLike(this.props.panelView.locationsOrCustom.locations) ? (
                        <FileLocationsError error={this.props.panelView.locationsOrCustom.locations} />
                    ) : (
                        this.props.panelView.locationsOrCustom.locations &&
                        this.props.repoName && (
                            <HierarchicalLocationsView
                                location={this.props.location}
                                locations={this.props.panelView.locationsOrCustom.locations}
                                defaultGroup={this.props.repoName}
                                isLightTheme={this.props.isLightTheme}
                                fetchHighlightedFileLines={this.props.fetchHighlightedFileLines}
                                extensionsController={this.props.extensionsController}
                                settingsCascade={this.props.settingsCascade}
                            />
                        )
                    )
                ) : (
                    <>
                        {this.props.panelView.locationsOrCustom.custom}
                        {!this.props.panelView.content && !this.props.panelView.locationsOrCustom.custom && (
                            <EmptyPanelView className="mt-3" />
                        )}
                    </>
                )}
            </div>
        )
    }
}
