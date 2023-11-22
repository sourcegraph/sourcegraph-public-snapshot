import type { FC, ReactElement } from 'react'

import type { QueryResult } from '@apollo/client'
import { mdiDelete, mdiInformationOutline, mdiPlus, mdiAlertCircle } from '@mdi/js'
import classNames from 'classnames'

import { pluralize } from '@sourcegraph/common'
import {
    Button,
    ErrorAlert,
    Icon,
    Link,
    LoadingSpinner,
    Popover,
    PopoverTrigger,
    PopoverContent,
    PopoverTail,
    Tooltip,
} from '@sourcegraph/wildcard'

import { type CodeHost, ExternalServiceKind, type GetCodeHostsResult } from '../../../../../graphql-operations'
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

    // Filter out all other external services since we don't properly support them
    // in the wizard, "Other" and "LocalGit" external services are used as local repositories setup in
    // Cody App for which we have a special setup step
    const nonOtherExternalServices = data.externalServices.nodes.filter(
        service => service.kind !== ExternalServiceKind.OTHER && service.kind !== ExternalServiceKind.LOCALGIT
    )

    if (nonOtherExternalServices.length === 0) {
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
                <span>Choose at least one of the code host providers from the list.</span>
            </small>
        )
    }

    return (
        <ul className={styles.list}>
            {createConnectionType && <CreateCodeHostConnectionCard codeHostType={createConnectionType} />}
            {nonOtherExternalServices.map(codeHost => (
                <li
                    key={codeHost.id}
                    className={classNames(styles.item, { [styles.itemActive]: codeHost.id === activeConnectionId })}
                >
                    <Button as={Link} to={`${codeHost.id}/edit`} className={styles.itemButton}>
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
                                {codeHost.lastSyncAt !== null && codeHost.lastSyncError === null && (
                                    <>Synced, {codeHost.repoCount} repositories found</>
                                )}
                                {codeHost.lastSyncAt === null && codeHost.lastSyncError === null && (
                                    <>
                                        Syncing
                                        {codeHost.repoCount > 0 && (
                                            <>
                                                , so far {codeHost.repoCount}{' '}
                                                {pluralize('repository', codeHost.repoCount ?? 0, 'repositories')} found
                                            </>
                                        )}
                                    </>
                                )}
                                {codeHost.lastSyncError !== null && (
                                    <Popover>
                                        <PopoverTrigger as="span" className={styles.errorButton}>
                                            Sync error appeared{' '}
                                            <Icon svgPath={mdiAlertCircle} aria-label="Sync error icon" />
                                        </PopoverTrigger>

                                        <PopoverContent position="right" className={styles.errorPopover}>
                                            <ErrorAlert
                                                error={codeHost.lastSyncError}
                                                variant="danger"
                                                className="m-3"
                                            />
                                        </PopoverContent>

                                        <PopoverTail size="sm" />
                                    </Popover>
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
