import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import * as React from 'react'
import { Dropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { ActivationStatus } from '../../../../shared/src/components/activation/Activation'

interface Props {
    /** Hide the "help" icon and dropdown. */
    noHelp?: boolean

    /** Never show the "Search" button label. */
    noLabel?: boolean

    activation?: ActivationStatus
}

interface State {
    isOpen: boolean

    didClick: boolean
}

/**
 * A search button with a dropdown with related links.
 */
export class SearchButton extends React.Component<Props, State> {
    public state: State = { isOpen: false, didClick: false }

    private onClick = (e: React.MouseEvent<HTMLElement, MouseEvent>) => {
        if (this.props.activation) {
            this.props.activation.update({ didSearch: true })
        }
    }

    public render(): JSX.Element | null {
        const docsURLPrefix = window.context.sourcegraphDotComMode ? 'https://docs.sourcegraph.com' : '/help'
        return (
            <div className="search-button d-flex">
                <button className="btn btn-primary search-button__btn" type="submit" onClick={this.onClick}>
                    <SearchIcon className="icon-inline" />
                    {!this.props.noLabel && <span className="search-button__label">Search</span>}
                </button>
                <Dropdown isOpen={this.state.isOpen} toggle={this.toggleIsOpen} className="d-flex">
                    {!this.props.noHelp && (
                        <>
                            <DropdownToggle
                                tag="span"
                                caret={false}
                                className="px-2 btn btn-link d-flex align-items-center"
                            >
                                <HelpCircleOutlineIcon className="icon-inline small" />
                            </DropdownToggle>
                            <DropdownMenu right={true} className="pb-0">
                                <DropdownItem header={true}>
                                    <strong>Search reference</strong>
                                </DropdownItem>
                                <DropdownItem divider={true} />
                                <DropdownItem header={true}>Finding matches:</DropdownItem>
                                <ul className="list-unstyled px-2 mb-2">
                                    <li>
                                        <span className="text-muted small">Regexp:</span>{' '}
                                        <code>
                                            <strong>(read|write)File</strong>
                                        </code>
                                    </li>
                                    <li>
                                        <span className="text-muted small">Exact:</span>{' '}
                                        <code>
                                            "<strong>fs.open(f)</strong>"
                                        </code>
                                    </li>
                                </ul>
                                <DropdownItem divider={true} />
                                <DropdownItem header={true}>Common search keywords:</DropdownItem>
                                <ul className="list-unstyled px-2 mb-2">
                                    <li>
                                        <code>
                                            repo:<strong>my/repo</strong>
                                        </code>
                                    </li>
                                    {window.context.sourcegraphDotComMode && (
                                        <li>
                                            <code>
                                                repo:<strong>github.com/myorg/</strong>
                                            </code>
                                        </li>
                                    )}
                                    <li>
                                        <code>
                                            file:<strong>my/file</strong>
                                        </code>
                                    </li>
                                    <li>
                                        <code>
                                            lang:<strong>javascript</strong>
                                        </code>
                                    </li>
                                </ul>
                                <DropdownItem divider={true} />
                                <DropdownItem header={true}>Diff/commit search keywords:</DropdownItem>
                                <ul className="list-unstyled px-2 mb-2">
                                    <li>
                                        <code>type:diff</code> <em className="text-muted small">or</em>{' '}
                                        <code>type:commit</code>
                                    </li>
                                    <li>
                                        <code>
                                            after:<strong>"2 weeks ago"</strong>
                                        </code>
                                    </li>
                                    <li>
                                        <code>
                                            author:<strong>alice@example.com</strong>
                                        </code>
                                    </li>
                                    <li className="text-nowrap">
                                        <code>
                                            repo:<strong>r@*refs/heads/</strong>
                                        </code>{' '}
                                        <span className="text-muted small">(all branches)</span>
                                    </li>
                                </ul>
                                <DropdownItem divider={true} />
                                <a
                                    href={`${docsURLPrefix}/user/search/queries`}
                                    className="dropdown-item d-flex align-items-center"
                                    target="_blank"
                                    onClick={this.toggleIsOpen}
                                >
                                    <ExternalLinkIcon className="icon-inline small mr-1 mb-1" /> All search keywords
                                </a>
                                {window.context.sourcegraphDotComMode && (
                                    <div className="p-2 alert alert-info small rounded-0 mb-0 mt-1">
                                        On Sourcegraph.com, use a <code>repo:</code> filter to narrow your search to
                                        &le;500 repositories.
                                    </div>
                                )}
                            </DropdownMenu>
                        </>
                    )}
                </Dropdown>
            </div>
        )
    }

    private toggleIsOpen = () => this.setState(prevState => ({ isOpen: !prevState.isOpen }))
}
