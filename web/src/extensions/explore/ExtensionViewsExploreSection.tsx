import H from 'history'
import marked from 'marked'
import React from 'react'
import { Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { ContributableViewContainer } from '../../../../shared/src/api/protocol'
import { PanelView } from '../../../../shared/src/api/protocol/plainTypes'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { Markdown } from '../../components/Markdown'
import { createLinkClickHandler } from '../../util/linkClickHandler'

interface Props extends ExtensionsControllerProps {
    history: H.History
}

interface State {
    /** Views contributed by extensions. */
    views?: (PanelView & { id: string })[] | null
}

/**
 * An explore section that shows views from extensions.
 *
 * TODO(sqs): This reuses panels displayed in the blob panel, which is hacky. The sourcegraph.app.createPanelView
 * API should let you specify where the panel should live.
 */
export class ExtensionViewsExploreSection extends React.PureComponent<Props, State> {
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.props.extensionsController.registries.views
                .getViews(ContributableViewContainer.Panel)
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
                {this.state.views.map((view, i) => (
                    <div key={i} className="mt-5">
                        <h2>{view.title}</h2>
                        <div onClick={createLinkClickHandler(this.props.history)}>
                            <Markdown dangerousInnerHTML={marked(view.content)} />
                        </div>
                    </div>
                ))}
            </div>
        )
    }
}
