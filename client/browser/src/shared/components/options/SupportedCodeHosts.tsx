import AddIcon from 'mdi-react/AddIcon'
import * as React from 'react'
import { Button, ButtonGroup, FormText, Input, InputGroup, InputGroupAddon } from 'reactstrap'
import * as permissions from '../../../browser/permissions'
import * as runtime from '../../../browser/runtime'
import storage from '../../../browser/storage'
import { EnterpriseURLList } from './EnterpriseURLList'

interface State {
    customCodeHost: string
    invalid: boolean
    enterpriseUrls: string[]
}

export class SupportedCodeHosts extends React.Component<{}, State> {
    public state: State

    constructor(props: any) {
        super(props)
        this.state = {
            customCodeHost: '',
            invalid: false,
            enterpriseUrls: [],
        }
    }

    public componentDidMount(): void {
        storage.getSync(items => {
            this.setState(() => ({ enterpriseUrls: items.enterpriseUrls || [] }))
        })

        storage.onChanged(items => {
            const { enterpriseUrls } = items
            if (enterpriseUrls && enterpriseUrls.newValue) {
                this.setState(() => ({ enterpriseUrls: enterpriseUrls.newValue || [] }))
            }
        })
    }

    private inputChanged = (e: React.ChangeEvent<HTMLInputElement>) => {
        e.persist()
        this.setState(() => ({ customCodeHost: e.target.value }))
    }

    private addEnterpriseUrl = () => {
        try {
            const url = new URL(this.state.customCodeHost)
            if (!url || !url.origin || url.origin === 'null') {
                this.handleInvalidUrl()
                return
            }

            permissions
                .request([this.state.customCodeHost])
                .then(granted => {
                    if (!granted) {
                        console.log('access not granted', granted)
                        return
                    }
                    runtime.sendMessage({ type: 'setEnterpriseUrl', payload: this.state.customCodeHost }, () => {
                        this.setState(() => ({ customCodeHost: '' }))
                    })
                })
                .catch(e => {
                    // TODO: handle error
                    console.log(e)
                })
        } catch {
            this.handleInvalidUrl()
        }
    }

    private handleKeyPress = (e: React.KeyboardEvent<HTMLElement>): void => {
        if (e.charCode === 13) {
            this.addEnterpriseUrl()
        }
    }

    private handleInvalidUrl = (): void => {
        this.setState(
            () => ({ invalid: true }),
            () => {
                setTimeout(() => this.setState(() => ({ invalid: false })), 2000)
            }
        )
    }

    public render(): JSX.Element | null {
        return (
            <div>
                <div className="options__section-header">Supported Code Hosts</div>
                <div className="options__section-contents">
                    <ButtonGroup>
                        <Button disabled={true}>GitHub</Button> <Button disabled={true}>Phabricator</Button>{' '}
                        <Button disabled={true}>Bitbucket Server</Button>{' '}
                    </ButtonGroup>
                    <div className="options__section-subheader">Code host URLs</div>
                    <div className="options__input-container">
                        <InputGroup>
                            <Input
                                invalid={this.state.invalid}
                                value={this.state.customCodeHost}
                                onKeyPress={this.handleKeyPress}
                                onChange={this.inputChanged}
                                className="options__input-field"
                                type="url"
                            />
                            <InputGroupAddon className="input-group-append" addonType="append">
                                <Button className="options__button-icon-add" onClick={this.addEnterpriseUrl} size="sm">
                                    <span className="icon-inline options__button-icon">
                                        <AddIcon size={17} />
                                    </span>
                                    Add
                                </Button>
                            </InputGroupAddon>
                        </InputGroup>
                        {this.state.invalid && <FormText color="muted">Please enter a URL.</FormText>}
                    </div>
                    <EnterpriseURLList enterpriseUrls={this.state.enterpriseUrls} />
                </div>
            </div>
        )
    }
}
