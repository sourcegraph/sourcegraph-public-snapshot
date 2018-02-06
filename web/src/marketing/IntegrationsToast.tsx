import PuzzleIcon from '@sourcegraph/icons/lib/Puzzle'
import { History, UnregisterCallback } from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { eventLogger } from '../tracking/eventLogger'
import { Toast } from './Toast'

interface State {
    visible: boolean
}

interface Props {
    history: History
}

/**
 * Renders a toast as long as the query contains toast=integrations. This toast will be rendered after sign-up and sign-in, if the
 * toast has already been dismissed we will not display the toast.
 */
export class IntegrationsToast extends React.Component<Props, State> {
    private unlisten: UnregisterCallback | undefined

    constructor(props: Props) {
        super(props)
        this.state = {
            visible: false,
        }
    }

    private updateToastVisibility(query: string): void {
        if (query.length > 0) {
            const parsedQuery = new URLSearchParams(location.search)
            if (parsedQuery && parsedQuery.get('toast') === 'integrations') {
                this.setState({
                    visible: true,
                })
                eventLogger.log('IntegrationsToastViewed')
                return
            }
        }
        this.setState({
            visible: false,
        })
    }

    public componentDidMount(): void {
        this.updateToastVisibility(this.props.history.location.search)
        this.unlisten = this.props.history.listen(location => {
            this.updateToastVisibility(location.search)
        })
    }

    public componentWillUnmount(): void {
        if (this.unlisten) {
            this.unlisten()
        }
    }

    public render(): JSX.Element | null {
        if (!this.state.visible) {
            return null
        }

        return (
            <Toast
                icon={<PuzzleIcon className="icon-inline" />}
                title="Configure Integrations"
                subtitle="Get Sourcegraph code search while reading code on GitHub and more."
                cta={
                    <div>
                        <Link to="/settings/integrations">
                            <button type="button" className="btn btn-primary" onClick={this.onClickConfigure}>
                                Configure
                            </button>
                        </Link>
                    </div>
                }
                onDismiss={this.onDismiss}
            />
        )
    }

    private onClickConfigure = (): void => {
        eventLogger.log('IntegrationsToastClicked')
        this.props.history.push('/settings/integrations')
        this.dismissToast()
    }

    private onDismiss = (): void => {
        eventLogger.log('IntegrationsToastDismissed')
        this.dismissToast()
    }

    private dismissToast = (): void => {
        this.setState({ visible: false })
    }
}
