import {
    ConfigurationCascadeOrError,
    ConfigurationSubject,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'
import * as React from 'react'
import { Alert, Card, CardBody, CardHeader, CardLink, CardText, Col, Row } from 'reactstrap'
import { Subscription } from 'rxjs'
import { GQL } from '../../../types/gqlschema'
import { isErrorLike } from '../../backend/errors'
import { configurationCascade } from '../../backend/extensions'
import { sourcegraphUrl } from '../../util/context'
import { BrowserSettingsEditor } from './BrowserSettingsEditor'

interface Props {
    currentUser: GQL.IUser | undefined
}

interface State {
    configurationCascadeOrError?: ConfigurationCascadeOrError<ConfigurationSubject, Settings>
}

export class SettingsCard extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            configurationCascade.subscribe(configurationCascadeOrError =>
                this.setState({ configurationCascadeOrError })
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <Row className="pb-3">
                <Col>
                    <Card>
                        <CardHeader>Settings</CardHeader>
                        <CardBody>
                            {this.props.currentUser ? (
                                <>
                                    <CardText>
                                        Browser extension settings are synchronized with user settings for{' '}
                                        <strong>{this.props.currentUser.username}</strong> on{' '}
                                        <a href={sourcegraphUrl}>{sourcegraphUrl}</a>.
                                    </CardText>
                                    <CardLink
                                        className="btn btn-primary"
                                        href={sourcegraphUrl + this.props.currentUser.settingsURL}
                                    >
                                        Edit user settings
                                    </CardLink>
                                    {this.state.configurationCascadeOrError === undefined ? (
                                        ''
                                    ) : isErrorLike(this.state.configurationCascadeOrError.merged) ? (
                                        <Alert color="danger">
                                            Error: {this.state.configurationCascadeOrError.merged.message}
                                        </Alert>
                                    ) : (
                                        <>
                                            <hr />
                                            <pre className="card-text">
                                                <code>
                                                    {JSON.stringify(
                                                        this.state.configurationCascadeOrError.merged,
                                                        null,
                                                        2
                                                    )}
                                                </code>
                                            </pre>
                                        </>
                                    )}
                                </>
                            ) : (
                                <BrowserSettingsEditor />
                            )}
                        </CardBody>
                    </Card>
                </Col>
            </Row>
        )
    }
}
