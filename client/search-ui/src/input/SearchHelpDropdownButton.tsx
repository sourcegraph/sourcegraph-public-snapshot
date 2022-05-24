import React, { useCallback, useState } from 'react'

import classNames from 'classnames'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    PopoverTrigger,
    PopoverContent,
    Popover,
    Button,
    Alert,
    Position,
    Link,
    MenuDivider,
    MenuHeader,
    Icon,
    Typography,
} from '@sourcegraph/wildcard'

import styles from './SearchHelpDropdownButton.module.scss'

interface SearchHelpDropdownButtonProps extends TelemetryProps {
    isSourcegraphDotCom?: boolean
}

/**
 * A dropdown button that shows a menu with reference documentation for Sourcegraph search query
 * syntax.
 */
export const SearchHelpDropdownButton: React.FunctionComponent<
    React.PropsWithChildren<SearchHelpDropdownButtonProps>
> = ({ isSourcegraphDotCom, telemetryService }) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])
    const onQueryDocumentationLinkClicked = useCallback(() => {
        telemetryService.log('SearchHelpDropdownQueryDocsLinkClicked')
        toggleIsOpen()
    }, [toggleIsOpen, telemetryService])
    const documentationUrlPrefix = isSourcegraphDotCom ? 'https://docs.sourcegraph.com' : '/help'

    return (
        <Popover isOpen={isOpen} onOpenChange={event => setIsOpen(event.isOpen)}>
            <PopoverTrigger
                as={Button}
                variant="link"
                className={classNames('px-2 d-flex align-items-center cursor-pointer', styles.triggerButton)}
                aria-label="Quick help for search"
            >
                <Icon
                    role="img"
                    aria-hidden={true}
                    className="test-search-help-dropdown-button-icon"
                    as={HelpCircleOutlineIcon}
                />
            </PopoverTrigger>
            <PopoverContent position={Position.bottomEnd} className={classNames('pb-0', styles.content)}>
                <MenuHeader>
                    <strong>Search reference</strong>
                </MenuHeader>
                <MenuDivider />
                <MenuHeader>Finding matches:</MenuHeader>
                <ul className="list-unstyled px-2 mb-2">
                    <li>
                        <span className="text-muted small">Structural:</span>{' '}
                        <Typography.Code weight="bold">if(:[my_match])</Typography.Code>
                    </li>
                    <li>
                        <span className="text-muted small">Regexp:</span>{' '}
                        <Typography.Code weight="bold">(read|write)File</Typography.Code>
                    </li>
                    <li>
                        <span className="text-muted small">Exact:</span>{' '}
                        <Typography.Code weight="bold">"fs.open(f)"</Typography.Code>
                    </li>
                </ul>
                <MenuDivider />
                <MenuHeader>Common search keywords:</MenuHeader>
                <ul className="list-unstyled px-2 mb-2">
                    <li>
                        <Typography.Code>
                            repo:<strong>my/repo</strong>
                        </Typography.Code>
                    </li>
                    {isSourcegraphDotCom && (
                        <li>
                            <Typography.Code>
                                repo:<strong>github.com/myorg/</strong>
                            </Typography.Code>
                        </li>
                    )}
                    <li>
                        <Typography.Code>
                            file:<strong>my/file</strong>
                        </Typography.Code>
                    </li>
                    <li>
                        <Typography.Code>
                            lang:<strong>javascript</strong>
                        </Typography.Code>
                    </li>
                </ul>
                <MenuDivider />
                <MenuHeader>Diff/commit search keywords:</MenuHeader>
                <ul className="list-unstyled px-2 mb-2">
                    <li>
                        <Typography.Code>type:diff</Typography.Code> <em className="text-muted small">or</em>{' '}
                        <Typography.Code>type:commit</Typography.Code>
                    </li>
                    <li>
                        <Typography.Code>
                            after:<strong>"2 weeks ago"</strong>
                        </Typography.Code>
                    </li>
                    <li>
                        <Typography.Code>
                            author:<strong>alice@example.com</strong>
                        </Typography.Code>
                    </li>
                    <li className="text-nowrap">
                        <Typography.Code>
                            repo:<strong>r@*refs/heads/</strong>
                        </Typography.Code>{' '}
                        <span className="text-muted small">(all branches)</span>
                    </li>
                </ul>
                <MenuDivider className="mb-0" />
                <Link
                    target="_blank"
                    rel="noopener"
                    to={`${documentationUrlPrefix}/code_search/reference/queries`}
                    className="dropdown-item"
                    onClick={onQueryDocumentationLinkClicked}
                >
                    <Icon role="img" aria-hidden={true} className="small" as={ExternalLinkIcon} /> All search keywords
                </Link>
                {isSourcegraphDotCom && (
                    <Alert className="small rounded-0 mb-0 mt-1" variant="info">
                        On Sourcegraph.com, use a <Typography.Code>repo:</Typography.Code> filter to narrow your search
                        to &le;500 repositories.
                    </Alert>
                )}
            </PopoverContent>
        </Popover>
    )
}
