import * as React from 'react'
import { Card, CardBody, CardHeader, Col, Form, FormGroup, Input, Label, Row } from 'reactstrap'
import { StorageItems } from '../../../browser/types'
import { sourcegraphUrl } from '../../util/context'
import { featureFlags } from '../../util/featureFlags'

interface Props {
    storage: StorageItems
}

export class FeatureFlagCard extends React.Component<Props, {}> {
    private onMermaidToggled = async () => {
        await featureFlags.toggle('renderMermaidGraphsEnabled')
    }

    private onInlineSymbolSearchToggled = async () => {
        await featureFlags.toggle('inlineSymbolSearchEnabled')
    }

    private onUseExtensionsToggled = async () => {
        await featureFlags.toggle('useExtensions')
    }

    private onSimpleOptionsMenu = async () => {
        await featureFlags.toggle('simpleOptionsMenu')
    }

    public render(): JSX.Element | null {
        const {
            simpleOptionsMenu,
            inlineSymbolSearchEnabled,
            renderMermaidGraphsEnabled,
            useExtensions,
        } = this.props.storage.featureFlags
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
                                <FormGroup check={true}>
                                    <Label check={true}>
                                        <Input
                                            onClick={this.onSimpleOptionsMenu}
                                            defaultChecked={simpleOptionsMenu}
                                            type="checkbox"
                                        />{' '}
                                        Enable the simpler options menu.
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
