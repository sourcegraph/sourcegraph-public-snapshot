import { History, UnregisterCallback } from 'history'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import * as React from 'react'
import { matchPath } from 'react-router'
import { Subscription } from 'rxjs'
import { SiteFlags } from '../site'
import { siteFlags } from '../site/backend'
import { eventLogger } from '../tracking/eventLogger'
import { showDotComMarketing } from '../util/features'
import { Toast } from './Toast'
import { daysActiveCount } from './util'

interface State {
    visible: boolean
    siteFlags?: SiteFlags
}

interface Props {
    history: History
}

const HAS_DISMISSED_TOAST_KEY = 'has-dismissed-integrations-toast'

/**
 * Renders a toast as long as the query contains toast=integrations. This toast will be rendered after sign-up and sign-in, if the
 * toast has already been dismissed we will not display the toast.
 */
export class IntegrationsToast extends React.Component<Props, State> {
    private unlisten: UnregisterCallback | undefined
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            visible: false,
        }
    }

    private updateToastVisibility(query: string): void {
        const canShow = localStorage.getItem(HAS_DISMISSED_TOAST_KEY) !== 'true' && !showDotComMarketing
        if (!canShow) {
            return
        }
        // Check if we explictily set the toast to be visible.
        const parsedQuery = new URLSearchParams(location.search)
        if (parsedQuery && parsedQuery.get('toast') === 'integrations') {
            this.showToast()
            return
        }

        // Do not show integrations toast on /search or /search?q= routes if it is their first session. Otherwise, show it.
        const match = matchPath<{ repoRev?: string; filePath?: string }>(location.pathname, { path: '/search' })
        if (match && daysActiveCount <= 1) {
            return
        }

        this.showToast()
    }

    public componentDidMount(): void {
        this.subscriptions.add(siteFlags.subscribe(siteFlags => this.setState({ siteFlags })))
        this.updateToastVisibility(this.props.history.location.search)
        this.unlisten = this.props.history.listen(location => {
            this.updateToastVisibility(location.search)
        })
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
        if (this.unlisten) {
            this.unlisten()
        }
    }

    public render(): JSX.Element | null {
        if (!this.state.visible) {
            return null
        }

        if (this.state.siteFlags) {
            if (this.state.siteFlags.needsRepositoryConfiguration || this.state.siteFlags.noRepositoriesEnabled) {
                return null
            }
        }

        return (
            <Toast
                icon={<PuzzleIcon className="icon-inline" />}
                title="Configure Integrations"
                subtitle="Get Sourcegraph code search while reading code on GitHub and more."
                cta={
                    <div>
                        <a target="_blank" href="https://about.sourcegraph.com/docs/integrations/">
                            <button type="button" className="btn btn-primary" onClick={this.onClickConfigure}>
                                Configure
                            </button>
                        </a>
                    </div>
                }
                onDismiss={this.onDismiss}
            />
        )
    }

    private showToast = (): void => {
        this.setState(() => ({ visible: true }))
        eventLogger.log('IntegrationsToastViewed')
    }

    private onClickConfigure = (): void => {
        eventLogger.log('IntegrationsToastClicked')
        this.dismissToast()
    }

    private onDismiss = (): void => {
        eventLogger.log('IntegrationsToastDismissed')
        this.dismissToast()
    }

    private dismissToast = (): void => {
        localStorage.setItem(HAS_DISMISSED_TOAST_KEY, 'true')
        this.setState({ visible: false })
    }
}
