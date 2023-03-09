import { FC, ReactElement } from 'react'

import { QueryResult } from '@apollo/client'
import { mdiInformationOutline, mdiDelete, mdiPlus } from '@mdi/js'
import classNames from 'classnames'

import { pluralize } from '@sourcegraph/common'
import { ErrorAlert, Icon, LoadingSpinner, Button, Tooltip, Link } from '@sourcegraph/wildcard'

import { CodeHost, GetCodeHostsResult } from '../../../../../graphql-operations'
import { CodeHostIcon, getCodeHostKindFromURLParam, getCodeHostName } from '../../helpers'

import styles from './CodeHostsNavigation.module.scss'

interface CodeHostsNavigationProps {
    codeHostQueryResult: QueryResult<GetCodeHostsResult>
    activeConnectionId: string | undefined
    createConnectionType: string | undefined
    className?: string
    onCodeHostDelete: (codeHost: CodeHost) => void
}

export const CodeHostsNavigation: FC<CodeHostsNavigationProps> = props => {
    const { codeHostQueryResult, activeConnectionId, createConnectionType, className, onCodeHostDelete } = props
    const { data, loading, error, refetch } = codeHostQueryResult

    if (error && !loading) {
        return (
            <div className={className}>
                <ErrorAlert error={error} />
                <Button variant="secondary" outline={true} size="sm" onClick={() => refetch()}>
                    Try fetch again
                </Button>
            </div>
        )
    }

    if (!data || (!data && loading)) {
        return (
            <small className={classNames(className, styles.loadingState)}>
                <LoadingSpinner /> Fetching connected code host...
            </small>
        )
    }

    if (data.externalServices.nodes.length === 0) {
        return (
            <small className={classNames(className, styles.emptyState)}>
                <span>
                    <Icon
                        width={24}
                        height={24}
                        aria-hidden={true}
                        svgPath={mdiInformationOutline}
                        className={styles.emptyStateIcon}
                    />
                </span>
                <span>Choose at least one of the code host providers from the list on the right.</span>
            </small>
        )
    }

    return (
        <ul className={styles.list}>
            {createConnectionType && <CreateCodeHostConnectionCard codeHostType={createConnectionType} />}
            {data.externalServices.nodes.map(codeHost => (
                <li
                    key={codeHost.id}
                    className={classNames(styles.item, { [styles.itemActive]: codeHost.id === activeConnectionId })}
                >
                    <Button
                        as={Link}
                        to={`/setup/remote-repositories/${codeHost.id}/edit`}
                        className={styles.itemButton}
                    >
                        <span>
                            <CodeHostIcon codeHostType={codeHost.kind} aria-hidden={true} />
                        </span>
                        <span className={styles.itemDescription}>
                            <span className={styles.itemTitle}>
                                {codeHost.displayName}
                                {codeHost.lastSyncAt === null && (
                                    <small>
                                        <LoadingSpinner />
                                    </small>
                                )}
                            </span>
                            <small className={styles.itemDescriptionStatus}>
                                {codeHost.lastSyncAt !== null && <>Synced, {codeHost.repoCount} repositories found</>}
                                {codeHost.lastSyncAt === null && (
                                    <>
                                        Syncing{' '}
                                        {codeHost.repoCount > 0 && (
                                            <>
                                                , so far {codeHost.repoCount}{' '}
                                                {pluralize('repository', codeHost.repoCount ?? 0, 'repositories')} found
                                            </>
                                        )}
                                    </>
                                )}
                            </small>
                        </span>
                    </Button>

                    <Tooltip content="Delete code host connection" placement="right" debounce={0}>
                        <Button className={styles.deleteButton} onClick={() => onCodeHostDelete(codeHost)}>
                            <Icon svgPath={mdiDelete} aria-label="Delete code host connection" />
                        </Button>
                    </Tooltip>
                </li>
            ))}
        </ul>
    )
}

interface CreateCodeHostConnectionCardProps {
    codeHostType: string
}

function CreateCodeHostConnectionCard(props: CreateCodeHostConnectionCardProps): ReactElement {
    const { codeHostType } = props
    const codeHostKind = getCodeHostKindFromURLParam(codeHostType)

    return (
        <li className={classNames(styles.item, styles.itemCreation, styles.itemActive)}>
            <span>
                <Icon svgPath={mdiPlus} aria-hidden={true} />
            </span>
            <span className={styles.itemDescription}>
                <span>
                    Connect <CodeHostIcon codeHostType={codeHostKind} aria-hidden={true} />{' '}
                    {getCodeHostName(codeHostKind)}
                </span>
                <small className={styles.itemDescriptionStatus}>
                    New code host will appear in the list as soon as you connect it
                </small>
            </span>
        </li>
    )
}
