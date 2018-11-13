import AddIcon from 'mdi-react/AddIcon'
import * as React from 'react'
import { Button, FormText, Input, InputGroup, InputGroupAddon } from 'reactstrap'
import { upsertSourcegraphUrl, URLError } from '../../../browser/helpers/storage'
import storage from '../../../browser/storage'
import { ServerURLSelection } from './ServerURLSelection'

interface State {
    customUrl: string
    error: number | null
    serverUrls: string[]
}

// Make safari not be abnoxious <angry face>
const safariInputAttributes = {
    autoComplete: 'off',
    autoCorrect: 'off',
    autoCapitalize: 'off',
    spellCheck: false,
}

export class ServerConnection extends React.Component<{}, State> {
    public state = {
        customUrl: '',
        error: null,
        serverUrls: [],
    }

    public componentDidMount(): void {
        storage.getSync(items => {
            this.setState(() => ({ serverUrls: items.serverUrls || [] }))
        })

        storage.onChanged(({ serverUrls }) => {
            if (serverUrls && serverUrls.newValue) {
                this.setState({ serverUrls: serverUrls.newValue })
            }
        })
    }

    private addSourcegraphServerURL = (): void => {
        // upsertSourcegraphUrl returns null or an error code
        const err = upsertSourcegraphUrl(this.state.customUrl, (urls: string[]) =>
            this.setState({ customUrl: '', serverUrls: urls, error: null })
        )
        if (err) {
            this.handleInvalidUrl(err)
        }
    }

    private handleInvalidUrl = (error: number): void => {
        this.setState(
            () => ({ error }),
            () => {
                setTimeout(() => this.setState({ error: null }), 2000)
            }
        )
    }

    private inputChanged = (e: React.ChangeEvent<HTMLInputElement>) => {
        e.persist()
        this.setState(() => ({ customUrl: e.target.value }))
    }

    private handleKeyPress = (e: React.KeyboardEvent<HTMLElement>) => {
        if (e.charCode === 13) {
            this.addSourcegraphServerURL()
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className="options__section">
                <div className="options__section-header">Sourcegraph URLs</div>
                <div className="options__section-contents">
                    <div className="options__input-container">
                        <InputGroup>
                            <Input
                                invalid={!!this.state.error}
                                onKeyPress={this.handleKeyPress}
                                value={this.state.customUrl}
                                onChange={this.inputChanged}
                                className="options__input-field"
                                {...safariInputAttributes as any}
                            />
                            <InputGroupAddon className="input-group-append" addonType="append">
                                <Button
                                    className="options__button-icon-add"
                                    onClick={this.addSourcegraphServerURL}
                                    size="sm"
                                >
                                    <span className="icon-inline options__button-icon">
                                        <AddIcon size={17} />
                                    </span>
                                    Add
                                </Button>
                            </InputGroupAddon>
                        </InputGroup>
                        {this.state.error === URLError.Invalid && (
                            <FormText color="muted">Please enter a valid URL.</FormText>
                        )}
                        {this.state.error === URLError.Empty && <FormText color="muted">Please enter a URL.</FormText>}
                        {this.state.error === URLError.HTTPNotSupported && (
                            <FormText color="muted">
                                Extensions cannot communicate over HTTPS in your browser. We suggest using a tool like{' '}
                                <a href="https://ngrok.com/">ngrok</a> for trying the extension out with your local
                                instance.
                            </FormText>
                        )}
                    </div>
                </div>
                <ServerURLSelection serverUrls={this.state.serverUrls} />
            </div>
        )
    }
}
