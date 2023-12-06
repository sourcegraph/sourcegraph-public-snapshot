import { useCallback, type FC } from 'react'

import { mdiHelpCircleOutline, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    PopoverTrigger,
    PopoverContent,
    PopoverTail,
    Popover,
    Button,
    Position,
    Link,
    MenuDivider,
    MenuHeader,
    Icon,
    MenuText,
    Code,
} from '@sourcegraph/wildcard'

import styles from './SearchHelpDropdownButton.module.scss'

interface SearchHelpDropdownButtonProps extends TelemetryProps, TelemetryV2Props {
    isSourcegraphDotCom?: boolean
    className?: string
}

/**
 * A dropdown button that shows a menu with reference documentation for Sourcegraph search query
 * syntax.
 */
export const SearchHelpDropdownButton: FC<SearchHelpDropdownButtonProps> = props => {
    const { isSourcegraphDotCom, className, telemetryService, telemetryRecorder } = props

    const onQueryDocumentationLinkClicked = useCallback(() => {
        telemetryService.log('SearchHelpDropdownQueryDocsLinkClicked')
        telemetryRecorder.recordEvent('SearchHelpDropdownQueryDocsLink', 'clicked')
    }, [telemetryService, telemetryRecorder])

    return (
        <Popover>
            <PopoverTrigger
                as={Button}
                variant="link"
                aria-label="Quick help for search"
                className={classNames(className, styles.triggerButton)}
                onClick={onQueryDocumentationLinkClicked}
            >
                <Icon
                    aria-hidden={true}
                    className="test-search-help-dropdown-button-icon"
                    svgPath={mdiHelpCircleOutline}
                />
            </PopoverTrigger>

            <PopoverContent position={Position.bottom} className={styles.content}>
                <MenuHeader>
                    <strong>Search reference</strong>
                </MenuHeader>
                <MenuDivider />
                <MenuHeader>Finding matches:</MenuHeader>
                <ul className="list-unstyled px-2 mb-2">
                    <li>
                        <span className="text-muted small">Structural:</span> <Code weight="bold">if(:[my_match])</Code>
                    </li>
                    <li>
                        <span className="text-muted small">Regexp:</span> <Code weight="bold">(read|write)File</Code>
                    </li>
                    <li>
                        <span className="text-muted small">Exact:</span> <Code weight="bold">fs.open(f)</Code>
                    </li>
                </ul>
                <MenuDivider />
                <MenuHeader>Common search keywords:</MenuHeader>
                <ul className="list-unstyled px-2 mb-2">
                    <li>
                        <Code>
                            repo:<strong>my/repo</strong>
                        </Code>
                    </li>
                    {isSourcegraphDotCom && (
                        <li>
                            <Code>
                                repo:<strong>github.com/myorg/</strong>
                            </Code>
                        </li>
                    )}
                    <li>
                        <Code>
                            file:<strong>my/file</strong>
                        </Code>
                    </li>
                    <li>
                        <Code>
                            lang:<strong>javascript</strong>
                        </Code>
                    </li>
                </ul>
                <MenuDivider />
                <MenuHeader>Diff/commit search keywords:</MenuHeader>
                <ul className="list-unstyled px-2 mb-2">
                    <li>
                        <Code>type:diff</Code> <em className="text-muted small">or</em> <Code>type:commit</Code>
                    </li>
                    <li>
                        <Code>
                            after:<strong>"2 weeks ago"</strong>
                        </Code>
                    </li>
                    <li>
                        <Code>
                            author:<strong>alice@example.com</strong>
                        </Code>
                    </li>
                    <li className="text-nowrap">
                        <Code>
                            repo:<strong>r@*refs/heads/</strong>
                        </Code>{' '}
                        <span className="text-muted small">(all branches)</span>
                    </li>
                </ul>
                <MenuDivider className="mb-0" />
                <MenuText
                    target="_blank"
                    rel="noopener"
                    as={Link}
                    to="/help/code_search/reference/queries"
                    onClick={onQueryDocumentationLinkClicked}
                >
                    <Icon aria-hidden={true} className="small" svgPath={mdiOpenInNew} /> All search keywords
                </MenuText>
            </PopoverContent>
            <PopoverTail size="sm" />
        </Popover>
    )
}
