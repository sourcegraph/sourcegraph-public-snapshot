import * as React from 'react'
import { NavLink } from 'react-router-dom'
import { Subscription } from 'rxjs'
import storage from '../../../browser/storage'
import { SourcegraphIcon } from '../Icons'

export const SIDEBAR_CARD_CLASS = 'card mb-3'
export const SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS = 'list-group-item list-group-item-action py-2'
export const SIDEBAR_BUTTON_CLASS = 'btn btn-secondary d-block w-100 my-2'

interface Props {
    className: string
}

interface State {
    useExtensions: boolean
}

/**
 * Sidebar for the options page.
 */
export class OptionsPageSidebar extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            useExtensions: false,
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            storage.observeSync('useExtensions').subscribe(useExtensions => this.setState(() => ({ useExtensions })))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className={`site-admin-sidebar ${this.props.className}`}>
                <div className={SIDEBAR_CARD_CLASS}>
                    <div className="card-header">
                        <SourcegraphIcon className="mr-1" />
                        Sourcegraph Extension
                    </div>
                    <div className="list-group list-group-flush">
                        <NavLink to="/" className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS} exact={true}>
                            Configuration
                        </NavLink>
                        {this.state.useExtensions && (
                            <NavLink to="/extensions" className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS} exact={true}>
                                Extensions
                            </NavLink>
                        )}
                    </div>
                </div>
            </div>
        )
    }
}
