import * as React from 'react'
import { Dropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import CursorTextIcon from 'mdi-react/CursorTextIcon'
import ViewQuiltIcon from 'mdi-react/ViewQuiltIcon'

interface Props {
    interactiveSearchMode: boolean
    toggleSearchMode: (event: React.MouseEvent<HTMLAnchorElement>) => void
}

interface State {
    isOpen: boolean
}

export class SearchModeToggle extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)

        this.state = {
            isOpen: false,
        }
    }
    private toggleOpen = (): void => {
        this.setState(state => ({
            isOpen: !state.isOpen,
        }))
    }
    public render(): JSX.Element {
        return (
            <Dropdown isOpen={this.state.isOpen} toggle={this.toggleOpen} className="search-mode-toggle">
                <DropdownToggle
                    caret={true}
                    className="search-mode-toggle__button e2e-search-mode-toggle"
                    data-tooltip="Toggle search mode"
                    aria-label="Toggle search mode"
                >
                    {this.props.interactiveSearchMode ? (
                        <ViewQuiltIcon className="icon-inline" size={8}></ViewQuiltIcon>
                    ) : (
                        <CursorTextIcon className="icon-inline" size={8}></CursorTextIcon>
                    )}
                </DropdownToggle>
                <DropdownMenu>
                    <DropdownItem
                        active={this.props.interactiveSearchMode}
                        onClick={!this.props.interactiveSearchMode ? this.props.toggleSearchMode : undefined}
                        className="e2e-search-mode-toggle__interactive-mode"
                    >
                        Interactive mode
                    </DropdownItem>
                    <DropdownItem
                        active={!this.props.interactiveSearchMode}
                        onClick={this.props.interactiveSearchMode ? this.props.toggleSearchMode : undefined}
                        className="e2e-search-mode-toggle__omni-mode"
                    >
                        Omni mode
                    </DropdownItem>
                </DropdownMenu>
            </Dropdown>
        )
    }
}
