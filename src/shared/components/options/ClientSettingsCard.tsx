import * as React from 'react'
import { Card, CardBody, CardHeader, Col, Row } from 'reactstrap'
import { BrowserSettingsEditor } from './BrowserSettingsEditor'

export class ClientSettingsCard extends React.Component {
    public render(): JSX.Element | null {
        return (
            <Row className="pb-3">
                <Col>
                    <Card>
                        <CardHeader>Client settings</CardHeader>
                        <CardBody>
                            <BrowserSettingsEditor />
                        </CardBody>
                    </Card>
                </Col>
            </Row>
        )
    }
}
