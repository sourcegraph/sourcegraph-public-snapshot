import * as React from 'react'
import storage from '../../browser/storage'
import { resolveRev } from '../repo/backend'
import { isSourcegraphDotCom } from '../util/context'
import { NeedsRepositoryConfigurationAlert } from './NeedsRepositoryConfigurationAlert'
import { NeedsServerConfigurationAlert } from './ServerAlert'

interface State {
    needsConfig: boolean
    alerts: string[]
}

interface Props {
    repoName: string
}

const SERVER_CONFIGURATION_KEY = 'NeedsServerConfigurationAlertDismissed'
const REPO_CONFIGURATION_KEY = 'NeedsRepoConfigurationAlertDismissed'

export class Alerts extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            needsConfig: false,
            alerts: [],
        }
    }

    public componentDidMount(): void {
        this.updateAlerts()
        resolveRev({ repoName: this.props.repoName })
            .toPromise()
            .catch(e => {
                this.setState(() => ({ ...this.state, needsConfig: true }))
            })
    }

    private updateAlerts = () => {
        storage.getSync(items => {
            const alerts: string[] = []
            if (!items[SERVER_CONFIGURATION_KEY]) {
                alerts.push(SERVER_CONFIGURATION_KEY)
            }
            if (!items[REPO_CONFIGURATION_KEY] || !items[REPO_CONFIGURATION_KEY][this.props.repoName]) {
                alerts.push(REPO_CONFIGURATION_KEY)
            }
            this.setState(() => ({ ...this.state, alerts }))
        })
    }

    public render(): JSX.Element | null {
        if (this.state.needsConfig && isSourcegraphDotCom() && this.state.alerts.includes(SERVER_CONFIGURATION_KEY)) {
            return <NeedsServerConfigurationAlert alertKey={SERVER_CONFIGURATION_KEY} onClose={this.updateAlerts} />
        }

        if (this.state.needsConfig && this.state.alerts.includes(REPO_CONFIGURATION_KEY) && !isSourcegraphDotCom()) {
            return (
                <NeedsRepositoryConfigurationAlert
                    repoName={this.props.repoName}
                    alertKey={REPO_CONFIGURATION_KEY}
                    onClose={this.updateAlerts}
                />
            )
        }

        return null
    }
}
