import React, { FunctionComponent } from 'react'

import { mdiChevronRight } from '@mdi/js'

import { Link, H3, Code, Icon } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../../components/time/Timestamp'
import { LockfileIndexFields } from '../../../../graphql-operations'

import styles from './CodeIntelLockfileIndexNode.module.scss'

export interface CodeIntelLockfileNodeProps {
    node: LockfileIndexFields
}

export const CodeIntelLockfileNode: FunctionComponent<React.PropsWithChildren<CodeIntelLockfileNodeProps>> = ({
    node,
}) => (
    <>
        <span className={styles.separator} />

        <div className="d-flex flex-column">
            <div className="m-0">
                <H3 className="m-0 d-block d-md-inline">
                    <Link to={node.repository.url}>{node.repository.name}</Link>
                </H3>
            </div>

            <div>
                <span className="mr-2 d-block d-mdinline-block">
                    Lockfile <Code>{node.lockfile}</Code> indexed at commit{' '}
                    <Link to={node.commit.url}>
                        <Code>{node.commit.abbreviatedOID}</Code>
                    </Link>
                    . Dependency graph fidelity: {node.fidelity}.
                </span>

                <small className="text-mute">
                    Indexed <Timestamp date={node.createdAt} />.{' '}
                    {node.createdAt !== node.updatedAt && (
                        <>
                            Updated <Timestamp date={node.updatedAt} />{' '}
                        </>
                    )}
                    .
                </small>
            </div>
        </div>

        <span className="d-none d-md-inline">
            <Link to={`./lockfiles/${node.id}`}>
                <Icon svgPath={mdiChevronRight} inline={false} aria-label="View more information" />
            </Link>
        </span>
    </>
)
