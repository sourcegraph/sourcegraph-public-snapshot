import { History, UnregisterCallback } from 'history'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import * as React from 'react'
import { matchPath } from 'react-router'
import { Link } from 'react-router-dom'
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

    private updateToastVisibility(): void {
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
        this.updateToastVisibility()
        this.unlisten = this.props.history.listen(() => {
            this.updateToastVisibility()
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
            if (this.state.siteFlags.needsRepositoryConfiguration) {
                return null
            }
        }

        return (
            <Toast
                icon={<PuzzleIcon className="icon-inline" />}
                title="Use with your code host"
                subtitle="Get Sourcegraph code intelligence while viewing code on GitHub, GitLab, Bitbucket Server, Phabricator, and more."
                cta={
                    <>
                        <Link
                            to="/help/integration/browser_extension"
                            className="btn btn-primary mr-2"
                            onClick={this.onClickConfigure}
                        >
                            Install
                        </Link>
                        <Link to="/help/integration/browser_extension" onClick={this.onClickConfigure}>
                            Learn more
                        </Link>
                    </>
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
