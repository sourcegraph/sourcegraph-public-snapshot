import React from 'react'

import { mdiBrain } from '@mdi/js'
import classNames from 'classnames'

import {
    Icon,
    LoadingSpinner,
    MenuDivider,
    MenuHeader,
    RadioButton,
    useSessionStorage,
    Code,
    Text,
} from '@sourcegraph/wildcard'

import { useVisibleIndexes } from '../hooks/useVisibleIndexes'

import styles from './BrainDot.module.scss'

export interface BrainDotProps {
    repoName: string
    commit: string
    path?: string
}

export const BrainDot: React.FunctionComponent<BrainDotProps> = ({ repoName, commit, path }) => (
    <>
        <MenuDivider />
        <MenuHeader className="d-flex">
            <Icon aria-hidden={true} svgPath={mdiBrain} fill="text-muted" />
            <Text className="mb-0 ml-2">Code intelligence preview</Text>
        </MenuHeader>
        <BrainDotContent repoName={repoName} commit={commit} path={path} />
    </>
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
                <div className={styles.container}>
                    <RadioButton
                        id="none"
                        key="none"
                        name="none"
                        label="None"
                        wrapperClassName={classNames(styles.radioBtn, 'py-1')}
                        checked={visibleIndexID === undefined}
                        onChange={() => {
                            delete indexIDsForSnapshotData[repoName]
                            setIndexIDForSnapshotData(indexIDsForSnapshotData)
                        }}
                    />
                    {visibleIndexes.map(index => (
                        <RadioButton
                            key={index.id}
                            id={index.id}
                            name={index.id}
                            checked={visibleIndexID === index.id}
                            wrapperClassName={classNames(styles.radioBtn, 'py-1')}
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
                    ))}
                </div>
            )}
            {(visibleIndexes?.length ?? 0) === 0 && !visibleIndexesLoading && (
                <small className="px-2">No precise indexes to display debug information for.</small>
            )}
        </>
    )
}
