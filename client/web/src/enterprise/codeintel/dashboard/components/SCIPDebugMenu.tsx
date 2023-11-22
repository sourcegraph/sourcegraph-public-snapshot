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

import styles from './SCIPDebugMenu.module.scss'

interface Props {
    repoName: string
    commit: string
    path?: string
}

export const SCIPDebugMenu: React.FunctionComponent<Props> = ({ repoName, commit, path }) => (
    <Menu>
        <Tooltip content="View code intelligence summary">
            <MenuButton className={classNames('text-decoration-none', styles.menuButton)} aria-label="Code graph">
                <Icon aria-hidden={true} svgPath={mdiBrain} />
            </MenuButton>
        </Tooltip>
        <MenuList position={Position.bottomEnd} className={styles.dropdownMenu}>
            <MenuHeader>
                <Link to={`/${repoName}/-/code-graph/dashboard`}>
                    View details
                    <Icon aria-hidden={true} svgPath={mdiArrowRightThin} />
                </Link>
            </MenuHeader>

            <MenuDivider />

            <SCIPDebugContent repoName={repoName} commit={commit} path={path} />
        </MenuList>
    </Menu>
)

const SCIPDebugContent: React.FunctionComponent<Props> = ({ repoName, commit, path }) => {
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
