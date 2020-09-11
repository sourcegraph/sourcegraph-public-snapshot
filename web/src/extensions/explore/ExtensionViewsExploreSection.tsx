import * as H from 'history'
import React from 'react'
import { Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { PanelViewWithComponent } from '../../../../shared/src/api/client/services/panelViews'
import { ContributableViewContainer } from '../../../../shared/src/api/protocol'
import { Markdown } from '../../../../shared/src/components/Markdown'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { renderMarkdown } from '../../../../shared/src/util/markdown'

interface Props extends ExtensionsControllerProps {
    history: H.History
}

interface State {
    /** Views contributed by extensions. */
    views?: PanelViewWithComponent[] | null
}

/**
 * An explore section that shows views from extensions.
 *
 * TODO(sqs): This reuses panels displayed in the blob panel, which is hacky. The sourcegraph.app.createPanelView
 * API should let you specify where the panel should live. This also does not support panel views with a component
 * (e.g., a location provider).
 */
export class ExtensionViewsExploreSection extends React.PureComponent<Props, State> {
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.props.extensionsController.services.panelViews
                .getPanelViews(ContributableViewContainer.Panel)
                .pipe(map(views => ({ views })))
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state || !this.state.views) {
            return null
        }

        return (
            <div className="extension-views-explore-section">
                {this.state.views.map((view, index) => (
                    <div key={index} className="mt-5">
                        <h2>{view.title}</h2>
                        <Markdown dangerousInnerHTML={renderMarkdown(view.content)} history={this.props.history} />
                    </div>
                ))}
            </div>
        )
    }
}
