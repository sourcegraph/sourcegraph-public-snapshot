import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React from 'react'
import { DropdownItem } from 'reactstrap'

const SearchContextMenuItem: React.FunctionComponent<{ spec: string; description: string; isDefault?: boolean }> = ({
    spec,
    description,
    isDefault = false,
}) => (
    <DropdownItem className="search-context-menu__item">
        <span className="search-context-menu__item-name" title={spec}>
            {spec}
        </span>
        <span className="search-context-menu__item-description" title={description}>
            {description}
        </span>
        {isDefault && <span className="search-context-menu__item-default">Default</span>}
    </DropdownItem>
)

export const SearchContextMenu: React.FunctionComponent<{}> = () => (
    <div className="search-context-menu">
        <div className="search-context-menu__header d-flex">
            <span aria-hidden="true" className="search-context-menu__header-prompt">
                <ChevronRightIcon className="icon-inline" />
            </span>
            <input type="search" placeholder="Find a context" className="search-context-menu__header-input" />
        </div>
        <div className="search-context-menu__list">
            <SearchContextMenuItem spec="global" description="All repositories on Sourcegraph" isDefault={true} />
            <SearchContextMenuItem spec="@username" description="Your repositories on Sourcegraph" />
            <SearchContextMenuItem
                spec="@username/context1"
                description="A test context with a very very long description lorem ipsum solor sit amet"
            />
            <SearchContextMenuItem spec="@username/contextwithveryverylongname" description="Another test context" />
        </div>
        <div className="search-context-menu__footer">
            <button type="button" className="btn btn-link btn-sm search-context-menu__footer-button">
                Reset
            </button>
            <span className="flex-grow-1" />
            <button type="button" className="btn btn-link btn-sm search-context-menu__footer-button">
                Manage contexts
            </button>
        </div>
    </div>
)
