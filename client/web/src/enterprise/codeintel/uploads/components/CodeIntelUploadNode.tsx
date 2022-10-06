import React, { FunctionComponent } from 'react'

import { mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'

import { Link, H3, Icon, Checkbox } from '@sourcegraph/wildcard'

import { LsifUploadFields } from '../../../../graphql-operations'
import { CodeIntelState } from '../../shared/components/CodeIntelState'
import { CodeIntelUploadOrIndexCommit } from '../../shared/components/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexCommitTags } from '../../shared/components/CodeIntelUploadOrIndexCommitTags'
import { CodeIntelUploadOrIndexRepository } from '../../shared/components/CodeIntelUploadOrIndexerRepository'
import { CodeIntelUploadOrIndexIndexer } from '../../shared/components/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexLastActivity } from '../../shared/components/CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadOrIndexRoot } from '../../shared/components/CodeIntelUploadOrIndexRoot'

import styles from './CodeIntelUploadNode.module.scss'

export interface CodeIntelUploadNodeProps {
    node: LsifUploadFields
    now?: () => Date
    selection: Set<string> | 'all'
    onCheckboxToggle: (id: string, checked: boolean) => void
}

export const CodeIntelUploadNode: FunctionComponent<React.PropsWithChildren<CodeIntelUploadNodeProps>> = ({
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
            <div className="m-0">
                <H3 className="m-0 d-block d-md-inline">
                    <CodeIntelUploadOrIndexRepository node={node} />
                </H3>
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
                    <CodeIntelUploadOrIndexLastActivity node={{ ...node, queuedAt: null }} now={now} />
                </small>
            </div>
        </div>

        <span className={classNames(styles.state, 'd-none d-md-inline')}>
            <CodeIntelState node={node} className="d-flex flex-column align-items-center" />
        </span>
        <span>
            <Link to={`./uploads/${node.id}`}>
                <Icon svgPath={mdiChevronRight} inline={false} aria-label="View more information" />
            </Link>
        </span>
    </>
)
