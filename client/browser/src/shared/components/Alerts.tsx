import * as React from 'react'
import { Subscription } from 'rxjs'
import storage from '../../browser/storage'
import { resolveRev } from '../repo/backend'
import { isSourcegraphDotCom } from '../util/context'
import { NeedsRepositoryConfigurationAlert, REPO_CONFIGURATION_KEY } from './NeedsRepositoryConfigurationAlert'
import { NeedsServerConfigurationAlert, SERVER_CONFIGURATION_KEY } from './ServerAlert'

interface State {
    needsConfig: boolean
    alerts: string[]
}

interface Props {
    repoName: string
}

export class Alerts extends React.Component<Props, State> {
    private subscriptions = new Subscription()
    constructor(props: Props) {
        super(props)
        this.state = {
            needsConfig: false,
            alerts: [],
        }
    }

    public componentDidMount(): void {
        this.updateAlerts()
        this.subscriptions.add(
            resolveRev({ repoName: this.props.repoName }).subscribe({
                error: e => {
                    this.setState(() => ({ ...this.state, needsConfig: true }))
                },
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private updateAlerts = () => {
        storage.getSync(items => {
            const alerts: string[] = []
            if (!items[SERVER_CONFIGURATION_KEY]) {
                alerts.push(SERVER_CONFIGURATION_KEY)
            }
            if (!items[REPO_CONFIGURATION_KEY] || !items[REPO_CONFIGURATION_KEY]![this.props.repoName]) {
                alerts.push(REPO_CONFIGURATION_KEY)
            }
            this.setState(() => ({ ...this.state, alerts }))
        })
    }

    public render(): JSX.Element | null {
        if (this.state.needsConfig && isSourcegraphDotCom() && this.state.alerts.includes(SERVER_CONFIGURATION_KEY)) {
            return <NeedsServerConfigurationAlert onClose={this.updateAlerts} />
        }

        if (this.state.needsConfig && this.state.alerts.includes(REPO_CONFIGURATION_KEY) && !isSourcegraphDotCom()) {
            return <NeedsRepositoryConfigurationAlert repoName={this.props.repoName} onClose={this.updateAlerts} />
        }

        return null
    }
}
