import { FC } from 'react'

import { mdiBitbucket, mdiGithub, mdiGitlab } from '@mdi/js'
import { noop } from 'lodash'

import { Icon, Link, LoadingSpinner, PageSwitcher, Text } from '@sourcegraph/wildcard'

import { ValueLegendList, ValueLegendListProps } from '../../site-admin/analytics/components/ValueLegendList'
import { SetupStep, SetupStepActions } from '../components/SetupTabs'

import styles from './CloningRepositoriesStep.module.scss'

const items: ValueLegendListProps['items'] = [
    {
        value: 10,
        description: 'Repositories',
        color: 'var(--purple)',
        tooltip:
            'Total number of repositories in the Sourcegraph instance. This number might be higher than the total number of repositories in the list below in case repository permissions do not allow you to view some repositories.',
    },
    {
        value: 10,
        description: 'Not cloned',
        color: 'var(--body-color)',
        position: 'right',
        tooltip: 'The number of repositories that have not been cloned yet.',
        filter: { name: 'status', value: 'not-cloned' },
    },
    {
        value: 10,
        description: 'Cloning',
        color: 'var(--body-color)',
        position: 'right',
        tooltip: 'The number of repositories that are currently being cloned.',
        filter: { name: 'status', value: 'cloning' },
    },
    {
        value: 0,
        description: 'Cloned',
        color: 'var(--body-color)',
        position: 'right',
        tooltip: 'The number of repositories that have been cloned.',
        filter: { name: 'status', value: 'cloned' },
    },
    {
        value: 0,
        description: 'Indexed',
        color: 'var(--body-color)',
        position: 'right',
        tooltip: 'The number of repositories that have been indexed for search.',
        filter: { name: 'status', value: 'indexed' },
    },
    {
        value: 0,
        description: 'Failed',
        color: 'var(--body-color)',
        position: 'right',
        tooltip: 'The number of repositories where the last syncing attempt produced an error.',
        filter: { name: 'status', value: 'failed-fetch' },
    },
]

interface Repository {
    name: string
    status: 'cloning' | 'cloned'
    codeHost: 'github' | 'gitlab' | 'bitbucket'
    details: string
}

const REPOSITORIES: Repository[] = [
    {
        name: 'sourcegraph/sourcegraph',
        status: 'cloning',
        codeHost: 'github',
        details: 'Fetched 1.1 mb out of 74mb, Shard: gitserver-0',
    },
    {
        name: 'sourcegraph/about',
        status: 'cloning',
        codeHost: 'github',
        details: 'Fetched 10.1 mb out of 74mb, Shard: gitserver-0',
    },
    {
        name: 'vovakulikov/tokio',
        status: 'cloned',
        codeHost: 'gitlab',
        details: 'Shard: gitserver-0',
    },
    {
        name: 'vovakulikov/pickers',
        status: 'cloned',
        codeHost: 'bitbucket',
        details: 'Shard: gitserver-0',
    },
    {
        name: 'vovakulikov/datepicker',
        status: 'cloned',
        codeHost: 'bitbucket',
        details: 'Shard: gitserver-0',
    },
    {
        name: 'vovakulikov/inputs',
        status: 'cloned',
        codeHost: 'bitbucket',
        details: 'Shard: gitserver-0',
    },
    {
        name: 'vovakulikov/grid',
        status: 'cloned',
        codeHost: 'bitbucket',
        details: 'Shard: gitserver-0',
    },
    {
        name: 'vovakulikov/ui-kit',
        status: 'cloning',
        codeHost: 'gitlab',
        details: 'Fetched 10.1 mb out of 74mb, Shard: gitserver-0',
    },
]

export const CloningRepositoriesStep: FC = props => (
    <SetupStep>
        <Text>
            Syncing repositories. It may take a few moments to clone and index each repository. Repository statuses are
            displayed below.
        </Text>
        <ValueLegendList items={items} />

        <ul className={styles.repoList}>
            {REPOSITORIES.map(repo => (
                <RepositoryItem key={repo.name} item={repo} />
            ))}
        </ul>

        <PageSwitcher
            totalCount={100}
            goToNextPage={noop}
            goToPreviousPage={noop}
            goToFirstPage={noop}
            goToLastPage={noop}
            hasNextPage={false}
            hasPreviousPage={true}
        />

        <SetupStepActions nextAvailable={true} finish={true} />
    </SetupStep>
)

const ICONS = {
    github: mdiGithub,
    gitlab: mdiGitlab,
    bitbucket: mdiBitbucket,
}

interface RepositoryItemProps {
    item: Repository
}

const RepositoryItem: FC<RepositoryItemProps> = props => (
    <li className={styles.repoItem}>
        <Icon svgPath={ICONS[props.item.codeHost]} className={styles.repoItemIcon} />{' '}
        <Link className="mr-2">{props.item.name}</Link>{' '}
        {props.item.status === 'cloning' ? (
            <>
                <LoadingSpinner className="mr-1" /> <small>Cloning</small>
            </>
        ) : (
            <small>Cloned</small>
        )}
        <small className={styles.repoItemDescription}>{props.item.details}</small>
    </li>
)
