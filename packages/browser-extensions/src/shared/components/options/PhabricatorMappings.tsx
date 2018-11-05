import { filter } from 'lodash'
import AddIcon from 'mdi-react/AddIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import { Button, FormText, Input, ListGroup, ListGroupItem } from 'reactstrap'
import storage from '../../../browser/storage'
import { PhabricatorMapping } from '../../../browser/types'

enum errors {
    Empty,
    CallsignInvalid,
    PathInvalid,
}

interface State {
    error: errors | null
    callsign: string
    path: string
    mappings: PhabricatorMapping[]
}

export class PhabricatorMappings extends React.Component<{}, State> {
    public state: State = {
        error: null,
        callsign: '',
        path: '',
        mappings: [],
    }

    public componentDidMount(): void {
        storage.getSync(items => {
            const mappings = items.phabricatorMappings || []
            this.setState(() => ({ mappings }))
        })
    }

    private addMappingClicked = () => {
        const { path, callsign } = this.state
        if (!path && !callsign) {
            this.setState(() => ({ error: errors.Empty }))
            return
        }
        if (!path || !path.trim().length) {
            this.setState(() => ({ error: errors.PathInvalid }))
            return
        }
        if (!callsign || !callsign.trim().length) {
            this.setState(() => ({ error: errors.CallsignInvalid }))
            return
        }
        storage.getSync(items => {
            const mappings = items.phabricatorMappings || []
            mappings.push({ callsign: callsign!, path: path! })
            storage.setSync({ phabricatorMappings: mappings }, () => {
                this.setState(() => ({
                    mappings,
                    callsign: '',
                    path: '',
                }))
            })
        })
    }

    private onCallsignInputChanged = (e: React.ChangeEvent<HTMLInputElement>) => {
        e.persist()
        this.setState(() => ({ callsign: e.target.value }))
    }

    private onPathInputChanged = (e: React.ChangeEvent<HTMLInputElement>) => {
        e.persist()
        this.setState(() => ({ path: e.target.value }))
    }

    private handleRemove = (mapping: PhabricatorMapping) => (e: React.MouseEvent<HTMLElement>): void => {
        e.persist()
        storage.getSync(items => {
            const mappings = items.phabricatorMappings || []
            const newMappings = filter(
                mappings,
                savedMapping => savedMapping.path !== mapping.path && savedMapping.callsign !== mapping.callsign
            )
            storage.setSync({ phabricatorMappings: newMappings }, () => {
                this.setState(() => ({
                    mappings: newMappings,
                    callsign: '',
                    path: '',
                }))
            })
        })
    }

    public render(): JSX.Element | null {
        return (
            <div>
                <Input
                    className="options__input-field options__input-spacer"
                    type="text"
                    id="callSign"
                    placeholder="Callsign"
                    invalid={this.state.error === errors.CallsignInvalid || this.state.error === errors.Empty}
                    onChange={this.onCallsignInputChanged}
                    value={this.state.callsign}
                />
                <Input
                    className="options__input-field options__input-spacer"
                    type="text"
                    id="repoPath"
                    placeholder="Path"
                    invalid={this.state.error === errors.PathInvalid || this.state.error === errors.Empty}
                    onChange={this.onPathInputChanged}
                    value={this.state.path}
                />
                {this.state.error === errors.Empty && (
                    <FormText color="muted">Please enter a callsign and repository path.</FormText>
                )}
                {this.state.error === errors.CallsignInvalid && (
                    <FormText color="muted">Please enter a callsign.</FormText>
                )}
                {this.state.error === errors.PathInvalid && (
                    <FormText color="muted">Please enter a repository path.</FormText>
                )}
                <div className="options__button-right">
                    <Button className="options__button-icon-add" onClick={this.addMappingClicked} size="sm">
                        <span className="icon-inline options__button-icon">
                            <AddIcon size={17} />
                        </span>
                        Add
                    </Button>
                </div>
                <div className="options__contents">
                    <ListGroup className="options__list-group">
                        {this.state.mappings.map((mapping, i) => (
                            <ListGroupItem className={`options__group-item justify-content-between`} key={i}>
                                <div className="options__row-padding">{mapping.callsign}</div>
                                <div>{mapping.path}</div>
                                <button
                                    onClick={this.handleRemove(mapping)}
                                    className="options__row-close btn btn-icon"
                                >
                                    <span style={{ verticalAlign: 'middle' }} className="icon-inline">
                                        <CloseIcon size={17} />
                                    </span>
                                </button>
                            </ListGroupItem>
                        ))}
                    </ListGroup>
                </div>
            </div>
        )
    }
}
