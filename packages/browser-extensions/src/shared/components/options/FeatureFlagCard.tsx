import * as React from 'react'
import { Card, CardBody, CardHeader, Col, Form, FormGroup, Input, Label, Row } from 'reactstrap'
import storage from '../../../browser/storage'
import { StorageItems } from '../../../browser/types'
import { sourcegraphUrl } from '../../util/context'

interface Props {
    storage: StorageItems
}

export class FeatureFlagCard extends React.Component<Props, {}> {
    private onMermaidToggled = () => {
        const renderMermaidGraphsEnabled = !this.props.storage.renderMermaidGraphsEnabled
        storage.setSync({ renderMermaidGraphsEnabled })
    }

    private onInlineSymbolSearchToggled = () => {
        storage.setSync({ inlineSymbolSearchEnabled: !this.props.storage.inlineSymbolSearchEnabled })
    }

    private onUseExtensionsToggled = () => {
        storage.setSync({ useExtensions: !this.props.storage.useExtensions })
    }

    public render(): JSX.Element | null {
        const { inlineSymbolSearchEnabled, renderMermaidGraphsEnabled, useExtensions } = this.props.storage
        return (
            <Row className="pb-3">
                <Col>
                    <Card>
                        <CardHeader>Feature flags</CardHeader>
                        <CardBody>
                            <Form>
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
                                        <Input
                                            onClick={this.onUseExtensionsToggled}
                                            defaultChecked={useExtensions}
                                            type="checkbox"
                                        />{' '}
                                        Use Sourcegraph extensions
                                        {useExtensions && (
                                            <>
                                                {' '}
                                                and{' '}
                                                <a href={sourcegraphUrl + '/extensions'} target="_blank">
                                                    enable extensions on the registry
                                                </a>
                                            </>
                                        )}
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
