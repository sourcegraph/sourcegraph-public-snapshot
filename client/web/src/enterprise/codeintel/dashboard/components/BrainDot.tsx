import React from 'react'

import { mdiArrowRightThin, mdiBrain } from '@mdi/js'
import classNames from 'classnames'

import {
    Position,
    Icon,
    Link,
    LoadingSpinner,
    Menu,
    MenuButton,
    MenuDivider,
    MenuHeader,
    MenuList,
    Tooltip,
    RadioButton,
    useSessionStorage,
    Code,
} from '@sourcegraph/wildcard'

import { useVisibleIndexes } from '../hooks/useVisibleIndexes'

import styles from './BrainDot.module.scss'

export interface BrainDotProps {
    repoName: string
    commit: string
    path?: string
}

export const BrainDot: React.FunctionComponent<BrainDotProps> = ({ repoName, commit, path }) => (
    <Menu>
        <Tooltip content="View code intelligence summary">
            <MenuButton className={classNames('text-decoration-none', styles.braindot)} aria-label="Code graph">
                <Icon aria-hidden={true} svgPath={mdiBrain} />
            </MenuButton>
        </Tooltip>
        <MenuList position={Position.bottomEnd} className={styles.dropdownMenu}>
            <MenuHeader>
                Click to view code intelligence summary
                <span className="float-right">
                    <Tooltip content="View code intelligence summary">
                        <Link to={`/${repoName}/-/code-graph/dashboard`}>
                            <Icon aria-hidden={true} svgPath={mdiArrowRightThin} />
                        </Link>
                    </Tooltip>
                </span>
            </MenuHeader>

            <MenuDivider />

            <BrainDotContent repoName={repoName} commit={commit} path={path} />
        </MenuList>
    </Menu>
)

const BrainDotContent: React.FunctionComponent<BrainDotProps> = ({ repoName, commit, path }) => {
    const { data: visibleIndexes, loading: visibleIndexesLoading } = useVisibleIndexes({
        repository: repoName,
        commit,
        path: path ?? '',
    })

    const [indexIDsForSnapshotData, setIndexIDForSnapshotData] = useSessionStorage<{
        [repoName: string]: string | undefined
    }>('blob.preciseIndexIDForSnapshotData', {})
    let visibleIndexID = indexIDsForSnapshotData[repoName]

    if (!visibleIndexes?.some(value => value.id === visibleIndexID)) {
        visibleIndexID = undefined
    }

    return (
        <>
            {visibleIndexesLoading && <LoadingSpinner className="mx-2" />}
            {visibleIndexes && visibleIndexes.length > 0 && (
                <MenuHeader>
                    <Tooltip content="Not intended for regular use">
                        <span>Display debug information for uploaded index.</span>
                    </Tooltip>
                    {[
                        <RadioButton
                            id="none"
                            key="none"
                            name="none"
                            label="None"
                            wrapperClassName="py-1"
                            checked={visibleIndexID === undefined}
                            onChange={() => {
                                delete indexIDsForSnapshotData[repoName]
                                setIndexIDForSnapshotData(indexIDsForSnapshotData)
                            }}
                        />,
                        ...visibleIndexes.map(index => (
                            <Tooltip content={`Uploaded at ${index.uploadedAt}`} key={index.id}>
                                <RadioButton
                                    id={index.id}
                                    name={index.id}
                                    checked={visibleIndexID === index.id}
                                    wrapperClassName="py-1"
                                    label={
                                        <>
                                            Index at <Code>{index.inputCommit.slice(0, 7)}</Code>
                                        </>
                                    }
                                    onChange={() => {
                                        indexIDsForSnapshotData[repoName] = index.id
                                        setIndexIDForSnapshotData(indexIDsForSnapshotData)
                                    }}
                                />
                            </Tooltip>
                        )),
                    ]}
                </MenuHeader>
            )}
            {(visibleIndexes?.length ?? 0) === 0 && !visibleIndexesLoading && (
                <MenuHeader>No precise indexes to display debug information for.</MenuHeader>
            )}
        </>
    )
}
