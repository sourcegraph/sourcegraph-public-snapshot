import { type FC, useState } from 'react'

import { mdiChevronUp, mdiChevronDown, mdiFileDocumentOutline } from '@mdi/js'
import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
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
} from '@sourcegraph/wildcard'

import { DiffStatStack } from '../components/diff/DiffStat'
import { usePageSwitcherPagination } from '../components/FilteredConnection/hooks/usePageSwitcherPagination'
import { ConnectionLoading, ConnectionError } from '../components/FilteredConnection/ui'
import type {
    SiteConfigurationChangeNode,
    SiteConfigurationHistoryResult,
    SiteConfigurationHistoryVariables,
} from '../graphql-operations'

import { SITE_CONFIGURATION_CHANGE_CONNECTION_QUERY } from './backend'

import styles from './SiteConfigurationChangeList.module.scss'

export const SiteConfigurationChangeList: FC = () => {
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
            {!!connection?.nodes?.length && (
                <div>
                    <Container className="mb-3">
                        <H3>Change history</H3>
                        {loading && <ConnectionLoading />}
                        {error && <ConnectionError errors={[error.message]} />}
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
function linesChanged(diffString: string): [number, number] {
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

export const SiteConfigurationHistoryItem: FC<SiteConfigurationHistoryItemProps> = ({ node }) => {
    const [open, setOpen] = useState<boolean>(false)
    const icon = open ? mdiChevronUp : mdiChevronDown
    const [removedLines, addedLines] = linesChanged(node.diff)

    const editedBy = node.author ? (
        <Link to={`/users/${node.author.username}`} className="text-truncate">
            {node.author.displayName ?? node.author.username}
        </Link>
    ) : (
        'Site configuration updated'
    )

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
                    <span className={styles.diffmeta}>
                        {node.author ? (
                            <UserAvatar className="ml-2 mr-2" user={node.author} size={32} />
                        ) : (
                            <Icon
                                aria-hidden={true}
                                svgPath={mdiFileDocumentOutline}
                                className={classNames('ml-2 mr-2', styles.fileicon)}
                                color="text-muted"
                            />
                        )}
                        <div className="d-flex flex-column align-items-start">
                            {editedBy}
                            <small className="text-muted">
                                Changed <Timestamp date={node.createdAt} />
                            </small>
                        </div>
                    </span>
                    <span className="ml-auto">
                        <DiffStatStack className="mr-1" added={addedLines} deleted={removedLines} />
                    </span>
                </CollapseHeader>
                <CollapsePanel>
                    <Code className={classNames('p-2', 'mt-2', styles.diffblock)}>{node.diff}</Code>
                </CollapsePanel>
            </Collapse>
            <hr className="mb-3 mt-3" />
        </>
    )
}
