import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useState } from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { HeroPage } from '../../../../components/HeroPage'
import { FileDiffHunks } from '../../../../repo/compare/FileDiffHunks'
import { FileDiffNode } from '../../../../repo/compare/FileDiffNode'
import { useEffectAsync } from '../../../../util/useEffectAsync'
import { QueryParameterProps } from '../../../threads/components/withQueryParameter/WithQueryParameter'
import { computeDiff, FileDiff } from '../../../threads/detail/changes/computeDiff'
import { Task } from '../../task'

interface Props extends QueryParameterProps, ExtensionsControllerProps, PlatformContextProps {
    task: Task

    location: H.Location
    history: H.History
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

/**
 * A list of changed files in a task.
 */
export const TaskFilesList: React.FunctionComponent<Props> = ({ task, ...props }) => {
    const [fileDiffsOrError, setFileDiffsOrError] = useState<typeof LOADING | FileDiff[] | ErrorLike>(LOADING)

    useEffectAsync(async () => {
        try {
            // TODO!(sqs)
            setFileDiffsOrError(await computeDiff(props.extensionsController, task.codeActions || []))
        } catch (err) {
            setFileDiffsOrError(asError(err))
        }
    }, [props.extensionsController, task.codeActions])
    if (fileDiffsOrError === LOADING) {
        return null // loading
    }
    if (isErrorLike(fileDiffsOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={fileDiffsOrError.message} />
    }

    return (
        <div className="task-files-list">
            {fileDiffsOrError.map((fileDiff: GQL.IFileDiff, key: string) => (
                <FileDiffNode
                    {...props}
                    lineNumbers={false}
                    base={{
                        repoName: 'github.com/sourcegraph/about',
                        repoID: '123' as any /* TODO!(sqs) */,
                        rev: 'master', // TODO!(sqs): un-hardcode master
                        commitID: 'master' /* TODO!(sqs) un-hardcode master */,
                    }}
                    head={{
                        repoName: 'github.com/sourcegraph/about',
                        repoID: '123' as any /* TODO!(sqs) */,
                        rev: 'master', // TODO!(sqs): un-hardcode master
                        commitID: 'master' /* TODO!(sqs) un-hardcode master */,
                    }}
                    node={{
                        ...fileDiff,
                        stat: { added: 1, changed: 2, deleted: 3 },
                        mostRelevantFile: {},
                        newFile: {},
                        oldFile: {},
                    }}
                />
            ))}
        </div>
    )
}
