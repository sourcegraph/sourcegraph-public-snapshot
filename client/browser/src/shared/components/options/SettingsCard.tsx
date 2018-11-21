import * as React from 'react'
import { Alert, Card, CardBody, CardHeader, CardLink, CardText, Col, Row } from 'reactstrap'
import { Subscription } from 'rxjs'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { SettingsCascadeOrError } from '../../../../../../shared/src/settings/settings'
import { isErrorLike } from '../../backend/errors'
import { settingsCascade } from '../../backend/extensions'
import { sourcegraphUrl } from '../../util/context'
import { BrowserSettingsEditor } from './BrowserSettingsEditor'

interface Props {
    currentUser: GQL.IUser | undefined
}

interface State {
    settingsCascadeOrError?: SettingsCascadeOrError
}

export class SettingsCard extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            settingsCascade.subscribe(settingsCascadeOrError => this.setState({ settingsCascadeOrError }))
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
                                    {this.state.settingsCascadeOrError === undefined ? (
                                        ''
                                    ) : isErrorLike(this.state.settingsCascadeOrError.final) ? (
                                        <Alert color="danger">
                                            Error: {this.state.settingsCascadeOrError.final.message}
                                        </Alert>
                                    ) : (
                                        <>
                                            <hr />
                                            <pre className="card-text">
                                                <code>
                                                    {JSON.stringify(this.state.settingsCascadeOrError.final, null, 2)}
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
