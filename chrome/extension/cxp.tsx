// We want to polyfill first.
import '../../app/util/polyfill'

import { ExtensionsList } from '@sourcegraph/extensions-client-common/lib/extensions/manager/ExtensionsList'
import {
    ConfigurationCascadeProps,
    ConfigurationSubject,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'
import * as React from 'react'
import { render } from 'react-dom'
import { Route } from 'react-router'
import { BrowserRouter, RouteComponentProps } from 'react-router-dom'
import { Button, FormGroup, Input, Label } from 'reactstrap'
import { Subscription } from 'rxjs'
import { createExtensionsContextController } from '../../app/backend/extensions'
import { useCXP } from '../../app/util/context'
import storage from '../../extension/storage'

const extensionsContextController = createExtensionsContextController()

interface State {
    clientSettings: string
}

class BrowserSettingsEditor extends React.Component<{}, State> {
    private subscriptions = new Subscription()

    constructor(props) {
        super(props)
        this.state = {
            clientSettings: 'Loading...',
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            storage.observeSync('clientSettings').subscribe(clientSettings => {
                this.setState(() => ({ clientSettings }))
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private saveLocalSettings = () => {
        storage.setSync({ clientSettings: this.state.clientSettings })
    }

    private onSettingsChanged = event => {
        const value = event.target.value
        this.setState(() => ({ clientSettings: value }))
    }

    public render(): JSX.Element | null {
        return (
            <div className="options__section">
                <div className="options__section-header">Client settings</div>
                <div className="options__section-contents">
                    <FormGroup>
                        <Label className="options__input">
                            <Input
                                className="options__input-textarea"
                                type="textarea"
                                value={this.state.clientSettings}
                                onChange={this.onSettingsChanged}
                                autoComplete="off"
                                autoCorrect="off"
                                autoCapitalize="off"
                                spellCheck={false}
                            />
                        </Label>
                        <Button className="options__cta" color="primary" onClick={this.saveLocalSettings}>
                            Save
                        </Button>
                    </FormGroup>
                </div>
            </div>
        )
    }
}

interface OptionsPageProps extends RouteComponentProps<{}> {}

interface OptionsPageState extends ConfigurationCascadeProps<ConfigurationSubject, Settings> {}

class OptionsPage extends React.PureComponent<OptionsPageProps, OptionsPageState> {
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

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return useCXP ? (
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
        ) : (
            <div>Enable CXP in the options popup to view the extensions registry.</div>
        )
    }
}

const inject = () => {
    const injectDOM = document.createElement('div')
    injectDOM.id = 'sourcegraph-options-menu'
    document.body.appendChild(injectDOM)
    render(
        <BrowserRouter key={0}>
            <Route path={'/cxp.html'} component={OptionsPage} />
        </BrowserRouter>,
        injectDOM
    )
}

document.addEventListener('DOMContentLoaded', () => {
    inject()
})
