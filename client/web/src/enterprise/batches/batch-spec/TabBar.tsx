import classNames from 'classnames'
import { upperFirst } from 'lodash'
import { NavLink as RouterLink } from 'react-router-dom'

import { BatchSpecExecutionFields, EditBatchChangeFields } from '../../../graphql-operations'

import styles from './TabBar.module.scss'

// We only use a subset of the batch change and batch spec fields to render the right state of the `TabBar`.
type BatchChangeFields = Pick<EditBatchChangeFields, 'name' | 'namespace'>
type BatchSpecFields = Pick<BatchSpecExecutionFields, 'id' | 'startedAt' | 'applyURL' | 'state'>

type TabName = 'configuration' | 'batch spec' | 'execution' | 'preview'

interface TabConfig {
    name: TabName
    buildLink: (batchChange: BatchChangeFields, batchSpec: BatchSpecFields) => string
    isLinkEnabled: (batchChange: BatchChangeFields | null, batchSpec: BatchSpecFields | null) => boolean
    buildDisabledTooltip?: (
        batchChange: BatchChangeFields | null,
        batchSpec: BatchSpecFields | null
    ) => string | undefined
}

const EXECUTION_TABS: TabConfig[] = [
    {
        name: 'configuration',
        buildLink: (batchChange: BatchChangeFields, batchSpec: BatchSpecFields) =>
            `${batchChange.namespace.url}/batch-changes/${batchChange.name}/executions/${batchSpec.id}/configuration`,
        isLinkEnabled: () => true,
    },
    {
        name: 'batch spec',
        buildLink: (batchChange: BatchChangeFields, batchSpec: BatchSpecFields) =>
            `${batchChange.namespace.url}/batch-changes/${batchChange.name}/executions/${batchSpec.id}/${
                batchSpec.state === 'PENDING' ? 'edit' : 'spec'
            }`,
        isLinkEnabled: (batchChange: BatchChangeFields | null) => !!batchChange,
    },
    {
        name: 'execution',
        buildLink: (batchChange: BatchChangeFields, batchSpec: BatchSpecFields) =>
            `${batchChange.namespace.url}/batch-changes/${batchChange.name}/executions/${batchSpec.id}/execution`,
        isLinkEnabled: (batchChange: BatchChangeFields | null, batchSpec: BatchSpecFields | null) =>
            !!batchChange && !!batchSpec?.startedAt,
    },
    {
        name: 'preview',
        buildLink: (batchChange: BatchChangeFields, batchSpec: BatchSpecFields) =>
            `${batchChange.namespace.url}/batch-changes/${batchChange.name}/executions/${batchSpec.id}/preview`,
        isLinkEnabled: (batchChange: BatchChangeFields | null, batchSpec: BatchSpecFields | null) =>
            !!batchChange && !!batchSpec?.applyURL,
        buildDisabledTooltip: (batchChange: BatchChangeFields | null, batchSpec: BatchSpecFields | null) =>
            !!batchChange && batchSpec?.startedAt ? 'Wait for execution to finish.' : undefined,
    },
]

type TabBarProps =
    | {
          batchChange: null
          batchSpec: null
          activeTabName: TabName
      }
    | {
          batchChange: BatchChangeFields
          batchSpec: BatchSpecFields
          activeTabName: TabName
      }

export const TabBar: React.FunctionComponent<TabBarProps> = ({ batchChange, batchSpec, activeTabName }) => (
    <ul className="nav nav-tabs d-inline-flex d-sm-flex flex-nowrap text-nowrap">
        {EXECUTION_TABS.map(({ name, buildLink, isLinkEnabled, buildDisabledTooltip }, index) => {
            const isEnabled = isLinkEnabled(batchChange, batchSpec)
            const tabName = `${index + 1}. ${upperFirst(name)}`

            return (
                <li className="nav-item" key={name}>
                    {activeTabName === name ? (
                        <span aria-disabled="true" className="nav-link active">
                            {tabName}
                        </span>
                    ) : isEnabled && !!batchChange ? (
                        <RouterLink
                            to={buildLink(batchChange, batchSpec)}
                            role="button"
                            className={classNames('nav-link', { active: activeTabName === name })}
                        >
                            {tabName}
                        </RouterLink>
                    ) : (
                        <span
                            aria-disabled="true"
                            className={classNames('nav-link text-muted', styles.navLinkDisabled)}
                            data-tooltip={
                                buildDisabledTooltip ? buildDisabledTooltip(batchChange, batchSpec) : undefined
                            }
                        >
                            {tabName}
                        </span>
                    )}
                </li>
            )
        })}
    </ul>
)
