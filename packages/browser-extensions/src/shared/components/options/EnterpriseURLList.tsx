import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import { ListGroup, ListGroupItem } from 'reactstrap'
import * as runtime from '../../../browser/runtime'
import storage from '../../../browser/storage'

interface State {
    enterpriseUrls: string[]
}

interface Props {
    enterpriseUrls: string[]
}

export class EnterpriseURLList extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            enterpriseUrls: [],
        }
    }

    public componentDidMount(): void {
        storage.onChanged(({ enterpriseUrls }) => {
            if (enterpriseUrls && enterpriseUrls.newValue) {
                this.setState({ enterpriseUrls: enterpriseUrls.newValue })
            }
        })
    }

    public componentWillReceiveProps(nextProps: Props): void {
        if (this.props.enterpriseUrls !== nextProps.enterpriseUrls) {
            this.setState(() => ({ enterpriseUrls: nextProps.enterpriseUrls }))
        }
    }

    private handleRemove = (url: string) => (e: React.MouseEvent<HTMLElement>): void => {
        e.preventDefault()
        e.stopPropagation()
        runtime.sendMessage({ type: 'removeEnterpriseUrl', payload: url })
    }

    public render(): JSX.Element | null {
        return (
            <ListGroup className="options__list-group">
                {this.state.enterpriseUrls.map((url, i) => (
                    <ListGroupItem className={`options__group-item justify-content-between`} key={i} disabled={true}>
                        <span className="options__group-item-text">{url}</span>
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
