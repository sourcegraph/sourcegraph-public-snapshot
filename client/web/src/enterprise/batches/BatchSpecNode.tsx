import React, { useCallback, useState } from 'react'

import {
    mdiCheckCircle,
    mdiTimerSand,
    mdiCancel,
    mdiAlertCircle,
    mdiChevronDown,
    mdiChevronRight,
    mdiStar,
    mdiPencil,
    mdiFileDownload,
    mdiFileDocumentOutline,
} from '@mdi/js'
import classNames from 'classnames'
import { upperFirst } from 'lodash'

import { BatchSpecSource, BatchSpecState } from '@sourcegraph/shared/src/graphql-operations'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Code, Link, Icon, H3, H4, Tooltip, Text, Button } from '@sourcegraph/wildcard'

import { Duration } from '../../components/time/Duration'
import { Timestamp } from '../../components/time/Timestamp'
import { BatchSpecListFields, Scalars } from '../../graphql-operations'

import { BatchSpec } from './BatchSpec'

import styles from './BatchSpecNode.module.scss'

export interface BatchSpecNodeProps extends ThemeProps {
    node: BatchSpecListFields
    currentSpecID?: Scalars['ID']
    /** Used for testing purposes. Sets the current date */
    now?: () => Date
}

export const BatchSpecNode: React.FunctionComponent<React.PropsWithChildren<BatchSpecNodeProps>> = ({
    node,
    currentSpecID,
    isLightTheme,
    now = () => new Date(),
}) => {
    const [isExpanded, setIsExpanded] = useState(currentSpecID === node.id)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(() => {
        setIsExpanded(!isExpanded)
    }, [isExpanded])

    return (
        <>
            <span className={styles.nodeSeparator} />
            <Button
                variant="icon"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronDown : mdiChevronRight} />
            </Button>
            <div className="d-flex flex-column justify-content-center align-items-center px-2 pb-1">
                <StateIcon source={node.source} state={node.state} />
                <span className="text-muted">
                    {node.source === BatchSpecSource.LOCAL && 'Uploaded'}
                    {node.source !== BatchSpecSource.LOCAL && upperFirst(node.state.toLowerCase())}
                </span>
            </div>
            <div className="px-2 pb-1">
                <H3 className="pr-2">
                    {currentSpecID && (
                        <Link to={`${node.namespace.url}/batch-changes/${node.description.name}/executions/${node.id}`}>
                            {currentSpecID === node.id && (
                                <>
                                    <Tooltip content="Currently applied spec">
                                        <Icon
                                            aria-label="Currently applied spec"
                                            className="text-warning"
                                            svgPath={mdiStar}
                                        />
                                    </Tooltip>{' '}
                                </>
                            )}
                            Created by <strong>{node.creator?.username}</strong>{' '}
                            <Timestamp date={node.createdAt} now={now} />
                        </Link>
                    )}
                    {!currentSpecID && (
                        <>
                            <Link className="text-muted" to={`${node.namespace.url}/batch-changes`}>
                                {node.namespace.namespaceName}
                            </Link>
                            <span className="text-muted d-inline-block mx-1">/</span>
                            <Link
                                to={`${node.namespace.url}/batch-changes/${node.description.name}/executions/${node.id}`}
                            >
                                {node.description.name || '-'}
                            </Link>
                        </>
                    )}
                </H3>
                {!currentSpecID && (
                    <small className="text-muted d-block">
                        Created by <strong>{node.creator?.username}</strong>{' '}
                        <Timestamp date={node.createdAt} now={now} />
                    </small>
                )}
            </div>
            <div className="text-center pb-1">
                {node.startedAt && <Duration start={node.startedAt} end={node.finishedAt ?? undefined} />}
            </div>
            {isExpanded && (
                <div className={styles.nodeExpandedSection}>
                    <ExpandedBatchSpec spec={node} isLightTheme={isLightTheme} />
                </div>
            )}
        </>
    )
}

interface ExpandedBatchSpecProps {
    spec: BatchSpecListFields
    isLightTheme: boolean
}

interface BatchSpecFile {
    isBinary: boolean
    content: string
    isSpecFile: boolean
    name: string
    id: string
}

const ExpandedBatchSpec: React.FunctionComponent<ExpandedBatchSpecProps> = ({ spec, isLightTheme }) => {
    const specFile: BatchSpecFile = {
        isBinary: false,
        content: spec.originalInput,
        isSpecFile: true,
        name: 'spec_file.yaml',
        id: spec.id,
    }
    const [selectedFile, setSelectedFile] = useState<BatchSpecFile>(specFile)

    if (spec.files && spec.files.totalCount > 0) {
        const mountedFiles: BatchSpecFile[] = spec.files.nodes.map(file => ({
            isBinary: file.binary,
            content: file.highlight.html,
            isSpecFile: false,
            name: file.name,
            id: file.id,
        }))
        const allFiles: BatchSpecFile[] = [specFile, ...mountedFiles]

        return (
            <div className={styles.specFilesContainer}>
                <ul className={styles.specFilesList}>
                    {allFiles.map(file => (
                        <li
                            key={file.id}
                            className={classNames(styles.specFilesListNode, {
                                [styles.specFilesListNodeActive]: file.id === selectedFile.id,
                            })}
                        >
                            <Button title={file.name} className={styles.specFilesListNodeButton} onClick={() => setSelectedFile(file)}>
                                {file.name}
                            </Button>
                        </li>
                    ))}
                </ul>

                {selectedFile.isSpecFile ? (
                    <BatchSpec
                        isLightTheme={isLightTheme}
                        name={spec.description.name}
                        originalInput={spec.originalInput}
                        className={classNames(styles.batchSpec, 'mb-0')}
                    />
                ) : (
                    <BatchSpecWorkspaceFileContent
                        name={selectedFile.name}
                        content={selectedFile.content}
                        isBinary={selectedFile.isBinary}
                    />
                )}
            </div>
        )
    }

    return (
        <>
            <H4>Input spec</H4>
            <BatchSpec
                isLightTheme={isLightTheme}
                name={spec.description.name}
                originalInput={spec.originalInput}
                className={classNames(styles.batchSpec, 'mb-0')}
            />
        </>
    )
}

interface BatchSpecWorkspaceFileContentProps {
    content: string
    isBinary: boolean
    name: string
}

const BatchSpecWorkspaceFileContent: React.FunctionComponent<BatchSpecWorkspaceFileContentProps> = ({
    content,
    isBinary,
    name,
}) => {
    if (isBinary) {
        return (
            <div className={styles.specFileBinary}>
                <Icon aria-hidden={true} svgPath={mdiFileDocumentOutline} className={styles.specFileBinaryIcon} />
                <Text className={styles.specFileBinaryName}>
                    {name} <span className={styles.specFileBinarySize}>4.5mb</span>
                </Text>
                <Button className={styles.specFileBinaryBtn}>
                    <Icon aria-hidden={true} svgPath={mdiFileDownload} />
                    {'  '}
                    Download file
                </Button>
            </div>
        )
    }

    return (
        <pre className={styles.blobWrapper}>
            <Code
                className={styles.blobCode}
                dangerouslySetInnerHTML={{
                    __html: content,
                }}
            />
        </pre>
    )
}

const StateIcon: React.FunctionComponent<
    React.PropsWithChildren<{ state: BatchSpecState; source: BatchSpecSource }>
> = ({ state, source }) => {
    if (source === BatchSpecSource.LOCAL) {
        return (
            <Icon
                aria-hidden={true}
                className={classNames(styles.nodeStateIcon, 'text-success mb-1')}
                svgPath={mdiCheckCircle}
            />
        )
    }
    switch (state) {
        case BatchSpecState.COMPLETED:
            return (
                <Icon
                    aria-hidden={true}
                    className={classNames(styles.nodeStateIcon, 'text-success mb-1')}
                    svgPath={mdiCheckCircle}
                />
            )

        case BatchSpecState.PROCESSING:
        case BatchSpecState.QUEUED:
            return (
                <Icon
                    aria-hidden={true}
                    className={classNames(styles.nodeStateIcon, 'text-muted mb-1')}
                    svgPath={mdiTimerSand}
                />
            )

        case BatchSpecState.CANCELED:
        case BatchSpecState.CANCELING:
            return (
                <Icon
                    aria-hidden={true}
                    className={classNames(styles.nodeStateIcon, 'text-muted mb-1')}
                    svgPath={mdiCancel}
                />
            )

        case BatchSpecState.FAILED:
            return (
                <Icon
                    aria-hidden={true}
                    className={classNames(styles.nodeStateIcon, 'text-danger mb-1')}
                    svgPath={mdiAlertCircle}
                />
            )
        case BatchSpecState.PENDING:
            return (
                <Icon
                    aria-hidden={true}
                    className={classNames(styles.nodeStateIcon, 'text-muted mb-1')}
                    svgPath={mdiPencil}
                />
            )
    }
}
