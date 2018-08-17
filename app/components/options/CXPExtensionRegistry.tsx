import { ExtensionsList } from '@sourcegraph/extensions-client-common/lib/extensions/manager/ExtensionsList'
import {
    ConfigurationCascadeProps,
    ConfigurationSubject,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'
import * as React from 'react'
import { RouteComponentProps } from 'react-router-dom'
import { Subscription } from 'rxjs'
import { createExtensionsContextController } from '../../../app/backend/extensions'
import { BrowserSettingsEditor } from '../../../chrome/extension/cxp'

interface OptionsPageProps extends RouteComponentProps<{}> {}
interface OptionsPageState extends ConfigurationCascadeProps<ConfigurationSubject, Settings> {}

const extensionsContextController = createExtensionsContextController()

export class CXPExtensionRegistry extends React.Component<OptionsPageProps, OptionsPageState> {
    public state: OptionsPageState = {
        configurationCascade: { subjects: [], merged: {} },
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            extensionsContextController.context.configurationCascade.subscribe(
                configurationCascade => this.setState({ configurationCascade }),
                err => console.error(err)
            )
        )
    }

    public render(): JSX.Element {
        return (
            <>
                <div>
                    Known issue: the extension links go nowhere. To view details, visit your Sourcegraph instance (e.g.{' '}
                    <a href="https://sourcegraph.com/extensions">sourcegraph.com/extensions</a>)
                </div>
                <ExtensionsList
                    {...this.props}
                    subject={'Client'}
                    configurationCascade={this.state.configurationCascade}
                    extensions={extensionsContextController}
                />
                <BrowserSettingsEditor />
            </>
        )
    }
}
