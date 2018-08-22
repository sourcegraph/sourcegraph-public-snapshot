import * as React from 'react'
import {
    Alert,
    Badge,
    Button,
    Card,
    CardBody,
    CardHeader,
    Col,
    FormGroup,
    FormText,
    Input,
    InputGroup,
    InputGroupAddon,
    ListGroupItemHeading,
    Row,
} from 'reactstrap'
import * as permissions from '../../../extension/permissions'
import storage from '../../../extension/storage'
import { StorageItems } from '../../../extension/types'
import { GQL } from '../../../types/gqlschema'
import { fetchSite } from '../../backend/server'
import { DEFAULT_SOURCEGRAPH_URL, isSourcegraphDotCom, setSourcegraphUrl, sourcegraphUrl } from '../../util/context'

interface Props {
    currentUser: GQL.IUser | undefined
    storage: StorageItems
    permissionOrigins: string[]
}

interface State {
    site?: GQL.ISite
    isUpdatingURL: boolean
    error: boolean
}

export class ConnectionCard extends React.Component<Props, State> {
    private urlInput: HTMLInputElement | null
    private contentScriptUrls: string[]

    constructor(props: Props) {
        super(props)
        this.state = {
            isUpdatingURL: false,
            error: false,
        }
    }

    public componentDidMount(): void {
        this.contentScriptUrls = this.props.storage.clientConfiguration.contentScriptUrls
        this.checkConnection()
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.contentScriptUrls = nextProps.storage.clientConfiguration.contentScriptUrls
    }

    private sourcegraphServerAlert = (): JSX.Element => {
        const { permissionOrigins } = this.props
        if (isSourcegraphDotCom()) {
            return (
                <div className="pt-2">
                    <Alert color="warning">Add a Server URL to enable support on private code.</Alert>
                </div>
            )
        }

        const { site } = this.state
        if (!site) {
            return (
                <div className="pt-2">
                    <Alert color="danger">
                        Error connecting to Server. Ensure you are authenticated and that the URL is correct.
                    </Alert>
                </div>
            )
        }
        const hasPermissions = this.contentScriptUrls.every(val => permissionOrigins.indexOf(`${val}/*`) >= 0)
        if (!hasPermissions && !permissionOrigins.includes('<all_urls>')) {
            return (
                <div className="pt-2">
                    <Alert color="warning">
                        {`Missing content script permissions: ${this.contentScriptUrls.join(', ')}.`}
                        <div className="pt-2">
                            <Button
                                onClick={this.requestPermissions}
                                color="primary"
                                className="btn btn-primary btn-sm"
                                size="sm"
                            >
                                Grant permissions
                            </Button>
                        </div>
                    </Alert>
                </div>
            )
        }

        if (!site.hasCodeIntelligence) {
            const isSiteAdmin = this.props.currentUser && this.props.currentUser.siteAdmin
            return (
                <div className="pt-2">
                    <Alert color="info">
                        {!isSiteAdmin &&
                            `Code intelligence is not enabled. Contact your site admin to enable language servers. Code
                        intelligence is available for open source repositories.`}
                        {isSiteAdmin && (
                            <div>
                                Code intelligence is disabled. Enable code intelligence for jump to definition, hover
                                tooltips, and find references.
                                <div className="pt-2">
                                    <Button
                                        href={`${sourcegraphUrl}/site-admin/code-intelligence`}
                                        color="primary"
                                        className="btn btn-primary btn-sm"
                                        size="sm"
                                    >
                                        Enable code intellligence
                                    </Button>
                                </div>
                            </div>
                        )}
                    </Alert>
                </div>
            )
        }

        return (
            <div className="pt-2">
                <Alert color="success">
                    You are connected to your server and code intelligence is fully functional.
                    <div className="pt-2">
                        <Button href={sourcegraphUrl} color="primary" className="btn btn-primary btn-sm" size="sm">
                            Open Sourcegraph
                        </Button>
                    </div>
                </Alert>
            </div>
        )
    }

    private serverStatusText = (): JSX.Element => {
        const { site } = this.state
        if (!site) {
            return <Badge color="danger">Unable to connect</Badge>
        }
        if (isSourcegraphDotCom()) {
            return <Badge color="warning">Limited functionality</Badge>
        }
        return <Badge color="success">Connected</Badge>
    }

    private updateButtonClicked = (): void => {
        this.setState(
            () => ({ isUpdatingURL: true }),
            () => {
                if (this.urlInput) {
                    this.urlInput.focus()
                    this.urlInput.select()
                }
            }
        )
    }

    private requestPermissions = (): void => {
        permissions.request(this.contentScriptUrls).then(
            () => {
                /** noop */
            },
            () => {
                /** noop */
            }
        )
    }

    private cancelButtonClicked = (): void => {
        this.setState(() => ({ isUpdatingURL: false }))
        if (!this.urlInput) {
            return
        }
        this.urlInput.value = sourcegraphUrl
        this.urlInput.blur()
    }

    private updateRef = (ref: HTMLInputElement | null): void => {
        this.urlInput = ref
    }

    private saveUrlButtonClicked = (): void => {
        if (!this.urlInput) {
            return
        }
        try {
            // If there is no url in the input use https://sourcegraph.com.
            const url = new URL(this.urlInput.value || DEFAULT_SOURCEGRAPH_URL)
            // (TODO): Remove serverUrl setting after release.
            storage.setSync({ sourcegraphURL: url.origin, serverUrls: [url.origin] })
            setSourcegraphUrl(url.origin)
            this.checkConnection()
            this.urlInput.value = url.origin
            this.setState({ isUpdatingURL: false, error: false })
        } catch {
            this.handleInvalidUrl()
        }
    }

    private handleInvalidUrl = (): void => {
        this.setState(
            () => ({ error: true }),
            () => {
                setTimeout(() => this.setState({ error: false }), 2000)
            }
        )
    }

    private handleKeyPress = (e: React.KeyboardEvent<HTMLElement>): void => {
        if (e.charCode === 13) {
            this.saveUrlButtonClicked()
        }
    }

    private checkConnection = (): void => {
        fetchSite().subscribe(
            site => {
                this.setState(() => ({ site }))
            },
            () => {
                this.setState(() => ({ site: undefined }))
            }
        )
    }

    public render(): JSX.Element | null {
        const { isUpdatingURL } = this.state
        return (
            <Row className="pb-3">
                <Col>
                    <Card>
                        <CardHeader>Sourcegraph configuration</CardHeader>
                        <CardBody>
                            <Col className="px-0">
                                <ListGroupItemHeading>Server connection</ListGroupItemHeading>
                                <FormGroup>
                                    <InputGroup>
                                        <Input
                                            invalid={!!this.state.error}
                                            type="url"
                                            required={true}
                                            innerRef={this.updateRef}
                                            readOnly={!isUpdatingURL}
                                            defaultValue={sourcegraphUrl}
                                            onKeyPress={this.handleKeyPress}
                                        />
                                        {!isUpdatingURL && (
                                            <InputGroupAddon className="input-group-append" addonType="append">
                                                <Button
                                                    onClick={this.updateButtonClicked}
                                                    color="primary"
                                                    className="btn btn-primary btn-sm"
                                                    size="sm"
                                                >
                                                    Update
                                                </Button>
                                            </InputGroupAddon>
                                        )}
                                        {isUpdatingURL && (
                                            <InputGroupAddon className="input-group-append" addonType="append">
                                                <Button
                                                    onClick={this.saveUrlButtonClicked}
                                                    color="primary"
                                                    className="btn btn-primary btn-sm"
                                                    size="sm"
                                                >
                                                    Save
                                                </Button>
                                                <Button
                                                    onClick={this.cancelButtonClicked}
                                                    color="secondary"
                                                    className="btn btn-secondary btn-sm"
                                                    size="sm"
                                                >
                                                    Cancel
                                                </Button>
                                            </InputGroupAddon>
                                        )}
                                    </InputGroup>
                                    {this.state.error && <FormText color="muted">Please enter a valid URL.</FormText>}
                                </FormGroup>
                                <ListGroupItemHeading className="pt-3">
                                    Status: {this.serverStatusText()}
                                    <Button
                                        onClick={this.checkConnection}
                                        size="sm"
                                        color="primary"
                                        className="float-right"
                                    >
                                        Check connection
                                    </Button>
                                </ListGroupItemHeading>
                                {this.sourcegraphServerAlert()}
                            </Col>
                        </CardBody>
                    </Card>
                </Col>
            </Row>
        )
    }
}
