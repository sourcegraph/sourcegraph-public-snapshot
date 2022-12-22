import { FunctionComponent } from 'react'

import { mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'

import { Link, H3, Icon, Checkbox, Tooltip } from '@sourcegraph/wildcard'

import { LsifIndexFields } from '../../../../graphql-operations'
import { CodeIntelState } from '../../shared/components/CodeIntelState'
import { CodeIntelUploadOrIndexCommit } from '../../shared/components/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexCommitTags } from '../../shared/components/CodeIntelUploadOrIndexCommitTags'
import { CodeIntelUploadOrIndexRepository } from '../../shared/components/CodeIntelUploadOrIndexerRepository'
import { CodeIntelUploadOrIndexIndexer } from '../../shared/components/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexLastActivity } from '../../shared/components/CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadOrIndexRoot } from '../../shared/components/CodeIntelUploadOrIndexRoot'

import styles from './CodeIntelIndexNode.module.scss'

export interface CodeIntelIndexNodeProps {
    node: LsifIndexFields
    now?: () => Date
    selection: Set<string> | 'all'
    onCheckboxToggle: (id: string, checked: boolean) => void
}

export const CodeIntelIndexNode: FunctionComponent<React.PropsWithChildren<CodeIntelIndexNodeProps>> = ({
    node,
    now,
    selection,
    onCheckboxToggle,
}) => (
    <>
        <span className={styles.separator} />

        <Checkbox
            label=""
            id="disabledFieldsetCheck"
            disabled={selection === 'all'}
            checked={selection === 'all' ? true : selection.has(node.id)}
            onChange={input => onCheckboxToggle(node.id, input.target.checked)}
        />

        <div className={classNames(styles.information, 'd-flex flex-column')}>
            <div className="m-0 d-flex flex-row">
                <H3 className="m-0 d-block d-md-inline">
                    <CodeIntelUploadOrIndexRepository node={node} />
                </H3>
                {node.shouldReindex && (
                    <Tooltip content="This index has been marked for reindexing.">
                        <div className={classNames(styles.tag, 'ml-1 rounded')}>(marked for reindexing)</div>
                    </Tooltip>
                )}
            </div>

            <div>
                <span className="mr-2 d-block d-mdinline-block">
                    Directory <CodeIntelUploadOrIndexRoot node={node} /> indexed at commit{' '}
                    <CodeIntelUploadOrIndexCommit node={node} />
                    {node.tags.length > 0 && (
                        <>
                            , <CodeIntelUploadOrIndexCommitTags tags={node.tags} />,
                        </>
                    )}{' '}
                    by <CodeIntelUploadOrIndexIndexer node={node} />
                </span>

                <small className="text-mute">
                    <CodeIntelUploadOrIndexLastActivity node={{ ...node, uploadedAt: null }} now={now} />
                </small>
            </div>
        </div>

        <span className={classNames(styles.state, 'd-none d-md-inline')}>
            <CodeIntelState node={node} className="d-flex flex-column align-items-center" />
        </span>
        <span>
            <Link to={`./indexes/${node.id}`}>
                <Icon svgPath={mdiChevronRight} inline={false} aria-label="View more information" />
            </Link>
        </span>
    </>
)
