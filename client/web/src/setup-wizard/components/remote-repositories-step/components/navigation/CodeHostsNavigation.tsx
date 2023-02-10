import { FC } from 'react';

import { gql, useQuery } from '@apollo/client';
import { mdiInformationOutline } from '@mdi/js';
import classNames from 'classnames';

import { ErrorAlert, Icon, LoadingSpinner, Button } from '@sourcegraph/wildcard';

import { GetCodeHostsResult } from '../../../../../graphql-operations';
import { getCodeHostIcon } from '../../helpers'

import styles from './CodeHostsNavigation.module.scss'

const GET_CODE_HOSTS = gql`
    query GetCodeHosts {
        externalServices {
            nodes {
                id
                kind
                repoCount
                displayName
                lastSyncAt
                nextSyncAt
            }
        }
    }
`

interface CodeHostsNavigationProps {
    className?: string
}

export const CodeHostsNavigation: FC<CodeHostsNavigationProps> = props => {
    const { className } = props

    const { data, loading, error, refetch } = useQuery<GetCodeHostsResult>(GET_CODE_HOSTS, {
        fetchPolicy: 'cache-and-network'
    })

    if (error && !loading) {
        return (
            <div className={classNames(className)}>
                <ErrorAlert error={error}/>
                <Button
                    variant='secondary'
                    outline={true}
                    size='sm'
                    onClick={() => refetch()}
                >
                    Try fetch again
                </Button>
            </div>
        )
    }

    if (!data || (!data && loading)) {
        return (
            <small className={classNames(className, styles.loadingState)}>
                <LoadingSpinner/>{' '}
                Fetching connected code host...
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
            { data.externalServices.nodes.map(codeHost =>
                <li key={codeHost.id} className={styles.item}>
                    <Button className={styles.itemButton}>
                        <span>
                             <Icon svgPath={getCodeHostIcon(codeHost.kind)} aria-hidden={true}/>
                        </span>
                        <span className={styles.itemDescription}>
                            <span>{codeHost.displayName}</span>
                            <small className='text-muted'>
                                {codeHost.lastSyncAt !== null && <>Synced, {codeHost.repoCount} repositories found</>}
                                {codeHost.lastSyncAt === null && <><LoadingSpinner/>, Syncing {codeHost.repoCount > 0 && <>, so far {codeHost.repoCount} repositories found</>}</>}
                            </small>
                        </span>
                    </Button>
                </li>
            )}
        </ul>
    )
}
