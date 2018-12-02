import H from 'history'
import marked from 'marked'
import React from 'react'
import { map } from 'rxjs/operators'
import { PanelViewWithComponent, ViewProviderRegistrationOptions } from '../../api/client/services/view'
import { RepositoryIcon } from '../../components/icons'
import { Markdown } from '../../components/Markdown'
import { createLinkClickHandler } from '../../util/linkClickHandler'
import { makeRepoURI } from '../../util/url'
import { EmptyPanelView } from './EmptyPanelView'
import { FileLocations } from './FileLocations'

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
            <div onClick={createLinkClickHandler(this.props.history)} className="panel__tabs-content">
                {this.props.panelView.content && (
                    <div className="px-2 pt-2">
                        <Markdown dangerousInnerHTML={marked(this.props.panelView.content)} />
                    </div>
                )}
                {this.props.panelView.reactElement}
                {this.props.panelView.locationProvider && (
                    <FileLocations
                        // tslint:disable-next-line:jsx-no-lambda
                        query={this.props.panelView.locationProvider}
                        // TODO!(sqs): add updates
                        icon={RepositoryIcon}
                        pluralNoun="locations"
                        isLightTheme={this.props.isLightTheme}
                    />
                )}
                {!this.props.panelView.content &&
                    !this.props.panelView.reactElement &&
                    !this.props.panelView.locationProvider && <EmptyPanelView />}
            </div>
        )
    }

    private queryLocations = () =>
        this.props.panelView.locationProvider!(
            // TODO!(sqs)
            {
                textDocument: {
                    uri: makeRepoURI(parseBrowserRepoURL(this.props.location.pathname)),
                },
                position: { line: 1, character: 3 },
            }
        ).pipe(
            map(locations => ({
                loading: false,
                locations: locations ? (Array.isArray(locations) ? locations : [locations]).filter(l => !!l) : [],
            }))
        )
}
