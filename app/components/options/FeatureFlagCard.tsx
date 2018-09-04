import * as React from 'react'
import { Card, CardBody, CardHeader, Col, Form, FormGroup, Input, Label, Row } from 'reactstrap'
import storage from '../../../extension/storage'
import { StorageItems } from '../../../extension/types'

interface Props {
    storage: StorageItems
}

export class FeatureFlagCard extends React.Component<Props, {}> {
    private onExecuteSearchToggled = () => {
        storage.setSync({ executeSearchEnabled: !this.props.storage.executeSearchEnabled })
    }

    private onFileTreeToggled = () => {
        storage.setSync({ repositoryFileTreeEnabled: !this.props.storage.repositoryFileTreeEnabled })
    }

    private onMermaidToggled = () => {
        const renderMermaidGraphsEnabled = !this.props.storage.renderMermaidGraphsEnabled
        storage.setSync({ renderMermaidGraphsEnabled })
    }

    private onInlineSymbolSearchToggled = () => {
        storage.setSync({ inlineSymbolSearchEnabled: !this.props.storage.inlineSymbolSearchEnabled })
    }

    private onUseCXPToggled = () => {
        storage.setSync({ useCXP: !this.props.storage.useCXP })
    }

    public render(): JSX.Element | null {
        const {
            inlineSymbolSearchEnabled,
            renderMermaidGraphsEnabled,
            repositoryFileTreeEnabled,
            executeSearchEnabled,
            useCXP,
        } = this.props.storage
        return (
            <Row className="pb-3">
                <Col>
                    <Card>
                        <CardHeader>Feature Flags</CardHeader>
                        <CardBody>
                            <Form>
                                <FormGroup check={true}>
                                    <Label check={true}>
                                        <Input
                                            onClick={this.onExecuteSearchToggled}
                                            defaultChecked={executeSearchEnabled}
                                            type="checkbox"
                                        />{' '}
                                        Open a new window with Sourcegraph search results when you perform a search on
                                        your code host.
                                    </Label>
                                </FormGroup>
                                <FormGroup check={true}>
                                    <Label check={true}>
                                        <Input
                                            onClick={this.onFileTreeToggled}
                                            defaultChecked={repositoryFileTreeEnabled}
                                            type="checkbox"
                                        />{' '}
                                        GitHub file tree navigation.
                                    </Label>
                                </FormGroup>
                                <FormGroup check={true}>
                                    <Label check={true}>
                                        <Input
                                            onClick={this.onMermaidToggled}
                                            defaultChecked={renderMermaidGraphsEnabled}
                                            type="checkbox"
                                        />{' '}
                                        <div className="options__input-label">
                                            Render{' '}
                                            <a
                                                href="https://mermaidjs.github.io/"
                                                target="_blank"
                                                // tslint:disable-next-line jsx-no-lambda
                                                onClick={e => e.stopPropagation()}
                                                rel="noopener"
                                                className="options__alert-link"
                                            >
                                                mermaid.js
                                            </a>{' '}
                                            diagrams on GitHub markdown files
                                        </div>
                                    </Label>
                                </FormGroup>
                                <FormGroup check={true}>
                                    <Label check={true}>
                                        <Input onClick={this.onUseCXPToggled} defaultChecked={useCXP} type="checkbox" />{' '}
                                        Use Sourcegraph extensions
                                    </Label>
                                </FormGroup>
                                <FormGroup check={true}>
                                    <Label check={true}>
                                        <Input
                                            onClick={this.onInlineSymbolSearchToggled}
                                            defaultChecked={inlineSymbolSearchEnabled}
                                            type="checkbox"
                                        />{' '}
                                        Enable inline symbol search by typing <code>!symbolQueryText</code> inside of
                                        GitHub PR comments (requires reload after toggling)
                                    </Label>
                                </FormGroup>
                            </Form>
                        </CardBody>
                    </Card>
                </Col>
            </Row>
        )
    }
}
