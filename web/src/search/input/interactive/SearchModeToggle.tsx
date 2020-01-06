import { Dropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import CursorTextIcon from 'mdi-react/CursorTextIcon'
import ViewQuiltIcon from 'mdi-react/ViewQuiltIcon'
import React, { useState, useCallback } from 'react'
import { Link } from 'react-router-dom'

interface Props {
    interactiveSearchMode: boolean
    toggleSearchMode: (event: React.MouseEvent<HTMLAnchorElement>) => void
}

export const SearchModeToggle: React.FunctionComponent<Props> = props => {
    const [isOpen, setIsOpen] = useState<boolean>(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen, setIsOpen])

    return (
        <Dropdown isOpen={isOpen} toggle={toggleIsOpen} className="search-mode-toggle">
            <DropdownToggle
                caret={true}
                className="search-mode-toggle__button e2e-search-mode-toggle"
                data-tooltip="Toggle search mode"
                aria-label="Toggle search mode"
            >
                {props.interactiveSearchMode ? (
                    <ViewQuiltIcon className="icon-inline" size={8}></ViewQuiltIcon>
                ) : (
                    <CursorTextIcon className="icon-inline" size={8}></CursorTextIcon>
                )}
            </DropdownToggle>
            <DropdownMenu>
                <DropdownItem
                    active={props.interactiveSearchMode}
                    onClick={!props.interactiveSearchMode ? props.toggleSearchMode : undefined}
                    className="e2e-search-mode-toggle__interactive-mode"
                >
                    Interactive mode
                </DropdownItem>
                <DropdownItem
                    active={!props.interactiveSearchMode}
                    onClick={props.interactiveSearchMode ? props.toggleSearchMode : undefined}
                    className="e2e-search-mode-toggle__omni-mode"
                >
                    Omni mode
                </DropdownItem>
                <DropdownItem tag={Link} to="/search/query-builder">
                    Query builder&hellip;
                </DropdownItem>
            </DropdownMenu>
        </Dropdown>
    )
}
