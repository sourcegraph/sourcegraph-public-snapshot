import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useState } from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { HeroPage } from '../../../../components/HeroPage'
import { QueryParameterProps } from '../../../../components/withQueryParameter/WithQueryParameter'
import { FileDiffNode } from '../../../../repo/compare/FileDiffNode'
import { useEffectAsync } from '../../../../util/useEffectAsync'
import { computeDiff, FileDiff } from '../../../threads/detail/changes/computeDiff'
import { ThreadSettings } from '../../../threads/settings'

interface Props extends QueryParameterProps, ExtensionsControllerProps, PlatformContextProps {
    thread: GQL.IDiscussionThread
    threadSettings: ThreadSettings

    location: H.Location
    history: H.History
    isLightTheme: boolean
}

/**
 * A list of changed files in a changeset.
 */
export const ChangesetFilesList: React.FunctionComponent<Props> = ({ thread, threadSettings, ...props }) => {
    const fileDiffsOrError = threadSettings.previewChangesetDiff

    if (!fileDiffsOrError) {
        return <p>TODO!(sqs): no changes precomputed</p>
    }

    return (
        <div className="changeset-files-list">
            {fileDiffsOrError.map((fileDiff, i) => (
                <FileDiffNode
                    key={i}
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
                        stat: { __typename: 'DiffStat', added: 1, changed: 2, deleted: 3 },
                        mostRelevantFile: {} as any,
                        newFile: {} as any,
                        oldFile: {} as any,
                    }}
                />
            ))}
        </div>
    )
}
