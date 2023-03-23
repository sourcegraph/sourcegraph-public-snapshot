import { FC, useState } from 'react'

import { mdiChevronUp, mdiChevronDown, mdiInformationOutline } from '@mdi/js'
import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import {
    Button,
    Link,
    Code,
    Container,
    Collapse,
    CollapseHeader,
    CollapsePanel,
    H3,
    Icon,
    PageSwitcher,
    Tooltip,
} from '@sourcegraph/wildcard'

import { DiffStatStack } from '../components/diff/DiffStat'
import { usePageSwitcherPagination } from '../components/FilteredConnection/hooks/usePageSwitcherPagination'
import { ConnectionLoading, ConnectionError } from '../components/FilteredConnection/ui'
import {
    SiteConfigurationChangeNode,
    SiteConfigurationHistoryResult,
    SiteConfigurationHistoryVariables,
} from '../graphql-operations'

import { SITE_CONFIGURATION_CHANGE_CONNECTION_QUERY } from './backend'

import styles from './SiteAdminConfigurationPage.module.scss'

export const SiteConfigurationChangeListPage: FC = () => {
    const { connection, loading, error, ...paginationProps } = usePageSwitcherPagination<
        SiteConfigurationHistoryResult,
        SiteConfigurationHistoryVariables,
        SiteConfigurationChangeNode
    >({
        query: SITE_CONFIGURATION_CHANGE_CONNECTION_QUERY,
        variables: {},
        getConnection: ({ data }) => data?.site?.configuration?.history || undefined,
    })

    return (
        <>
            {loading && <ConnectionLoading />}
            {error && <ConnectionError errors={[error.message]} />}
            {connection?.nodes?.length === 0 ? (
                <></>
            ) : (
                <div>
                    <Container className="mb-3">
                        <H3>History</H3>
                        <div className="mt-4">
                            {connection?.nodes
                                .filter(node => node.diff)
                                .map(node => (
                                    <SiteConfigurationHistoryItem key={node.id} node={node} />
                                ))}
                        </div>
                        <PageSwitcher
                            {...paginationProps}
                            className="mt-4"
                            totalCount={connection?.totalCount || 0}
                            totalLabel="changes"
                        />
                    </Container>
                </div>
            )}
        </>
    )
}
interface SiteConfigurationHistoryItemProps {
    node: SiteConfigurationChangeNode
}
function linesChanged(diffString: string | null): [number, number] {
    if (diffString === null) {
        return [0, 0]
    }

    return diffString
        .split('\n')
        .slice(3)
        .reduce(
            (summary, line) => {
                if (line.startsWith('-')) {
                    summary[0]++
                }
                if (line.startsWith('+')) {
                    summary[1]++
                }
                return summary
            },
            [0, 0]
        )
}
const SiteConfigurationHistoryItem: FC<SiteConfigurationHistoryItemProps> = ({ node }) => {
    const [open, setOpen] = useState<boolean>(false)
    const icon = open ? mdiChevronUp : mdiChevronDown

    if (node.reproducedDiff) {
        const [removedLines, addedLines] = linesChanged(node.diff)

        return (
            <>
                <Collapse key={node.id} isOpen={open} onOpenChange={setOpen}>
                    <CollapseHeader
                        as={Button}
                        aria-expanded={open}
                        type="button"
                        className="d-flex p-0 justify-content-start w-100"
                    >
                        <Icon aria-hidden={true} svgPath={icon} />
                        <span>
                            Changed <Timestamp date={node.createdAt} />
                            {node.author ? (
                                <>
                                    {' '}
                                    <span className="ml-1">
                                        by{' '}
                                        <Link to={`/users/${node.author.username}`} className="text-truncate">
                                            {node.author.displayName}
                                        </Link>
                                    </span>
                                </>
                            ) : (
                                <Tooltip content="Author information is not available because this change was made directly by editing the SITE_CONFIG_FILE">
                                    <Icon
                                        className="ml-1"
                                        svgPath={mdiInformationOutline}
                                        aria-label="Author information is not available because this change was made directly by editing the SITE_CONFIG_FILE"
                                    />
                                </Tooltip>
                            )}
                        </span>
                        {node.diff && (
                            <span className="ml-auto">
                                <DiffStatStack className="mr-1" added={addedLines} deleted={removedLines} />
                            </span>
                        )}
                    </CollapseHeader>
                    <CollapsePanel>
                        <Code className={classNames('p-2', 'mt-2', styles.diffblock)}>{node.diff}</Code>
                    </CollapsePanel>
                </Collapse>
                <hr className="mb-3 mt-3" />{' '}
            </>
        )
    }
    return (
        <CollapseHeader className="d-block mb-3">
            <>
                {node.author} {node.createdAt}
            </>
        </CollapseHeader>
    )
}
