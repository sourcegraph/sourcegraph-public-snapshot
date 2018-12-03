import H from 'history'
import marked from 'marked'
import React from 'react'
import { fetchHighlightedFileLines } from '../../../../web/src/repo/backend'
import { PanelViewWithComponent, ViewProviderRegistrationOptions } from '../../api/client/services/view'
import { RepositoryIcon } from '../../components/icons'
import { Markdown } from '../../components/Markdown'
import { createLinkClickHandler } from '../../util/linkClickHandler'
import { EmptyPanelView } from './EmptyPanelView'
import { FileLocations } from './FileLocations'
import { HierarchicalLocationsView } from './HierarchicalLocationsView'

interface Props {
    panelView: PanelViewWithComponent & Pick<ViewProviderRegistrationOptions, 'id'>
    history: H.History
    location: H.Location
    isLightTheme: boolean
}

interface State {}

/**
 * A panel view contributed by an extension using {@link sourcegraph.app.createPanelView}.
 */
export class PanelView extends React.PureComponent<Props, State> {
    public render(): JSX.Element | null {
        return (
            <div
                onClick={createLinkClickHandler(this.props.history)}
                className="panel__tabs-content panel__tabs-content--scroll"
            >
                {this.props.panelView.content && (
                    <div className="px-2 pt-2">
                        <Markdown dangerousInnerHTML={marked(this.props.panelView.content)} />
                    </div>
                )}
                {this.props.panelView.reactElement}
                {this.props.panelView.locationProvider && (
                    <HierarchicalLocationsView
                        locations={this.props.panelView.locationProvider}
                        icon={RepositoryIcon}
                        isLightTheme={this.props.isLightTheme}
                        fetchHighlightedFileLines={fetchHighlightedFileLines}
                    />
                )}
                {!this.props.panelView.content &&
                    !this.props.panelView.reactElement &&
                    !this.props.panelView.locationProvider && <EmptyPanelView className="mt-3" />}
            </div>
        )
    }
}
