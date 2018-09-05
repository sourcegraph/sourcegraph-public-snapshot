import { without } from 'lodash'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import { Badge, ListGroup, ListGroupItem } from 'reactstrap'
import storage from '../../../browser/storage'

interface State {
    serverUrls: string[]
    sourcegraphUrl: string
}

interface Props {
    serverUrls: string[]
}

export class ServerURLSelection extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            serverUrls: this.props.serverUrls,
            sourcegraphUrl: '',
        }
    }

    public componentDidMount(): void {
        storage.getSync(({ sourcegraphURL }) => this.setState(() => ({ sourcegraphUrl: sourcegraphURL })))

        storage.onChanged(({ sourcegraphURL, serverUrls }) => {
            const newState = {
                serverUrls: this.state.serverUrls,
                sourcegraphUrl: this.state.sourcegraphUrl,
            }

            if (sourcegraphURL) {
                newState.sourcegraphUrl = sourcegraphURL.newValue!
            }

            if (serverUrls) {
                newState.serverUrls = serverUrls.newValue!
            }

            this.setState(newState)
        })
    }

    public componentWillReceiveProps(nextProps: Props): void {
        if (this.props.serverUrls !== nextProps.serverUrls) {
            this.setState(() => ({ serverUrls: nextProps.serverUrls }))
        }
    }

    private handleClick = (url: string) => (e: React.MouseEvent<HTMLElement>) => {
        storage.setSync({ sourcegraphURL: url }, () => {
            this.setState(() => ({ sourcegraphUrl: url }))
        })
    }

    private handleRemove = (url: string) => (e: React.MouseEvent<HTMLElement>): void => {
        e.preventDefault()
        e.stopPropagation()
        storage.getSync(items => {
            const urls = items.serverUrls || []
            const serverUrls = without(urls, url)
            // If the current primary server url is being removed,
            // use the first url in the list or empty string if none are left
            const sourcegraphURL =
                url !== items.sourcegraphURL ? items.sourcegraphURL : serverUrls.find(u => u !== url) || ''

            storage.setSync(
                {
                    serverUrls,
                    sourcegraphURL,
                },
                () => {
                    this.setState(() => ({ serverUrls: without(urls, url), sourcegraphUrl: sourcegraphURL }))
                }
            )
        })
    }

    public render(): JSX.Element | null {
        return (
            <ListGroup className="options__list-group">
                {this.state.serverUrls.map((url, i) => (
                    <ListGroupItem
                        className={`options__group-item ${
                            url === this.state.sourcegraphUrl ? 'options__group-item-disabled' : ''
                        } justify-content-between`}
                        key={i}
                        disabled={url === this.state.sourcegraphUrl}
                        action={true}
                        onClick={this.handleClick(url)}
                    >
                        <span className="options__group-item-text">{url}</span>
                        {url === this.state.sourcegraphUrl && (
                            <Badge className="options__item-badge" pill={true}>
                                Primary
                            </Badge>
                        )}
                        <button onClick={this.handleRemove(url)} className="options__row-close btn btn-icon">
                            <span style={{ verticalAlign: 'middle' }} className="icon-inline">
                                <CloseIcon size={17} />
                            </span>
                        </button>
                    </ListGroupItem>
                ))}
            </ListGroup>
        )
    }
}
