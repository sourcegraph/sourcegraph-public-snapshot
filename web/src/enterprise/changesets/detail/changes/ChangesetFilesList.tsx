import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { QueryParameterProps } from '../../../../components/withQueryParameter/WithQueryParameter'
import { RepositoryCompareDiffPage } from '../../../../repo/compare/RepositoryCompareDiffPage'
import { ThreadSettings } from '../../../threads/settings'

interface Props extends QueryParameterProps, ExtensionsControllerProps, PlatformContextProps {
    thread: GQL.IDiscussionThread
    xchangeset: GQL.IChangeset
    threadSettings: ThreadSettings

    location: H.Location
    history: H.History
    isLightTheme: boolean
}

/**
 * A list of changed files in a changeset.
 */
export const ChangesetFilesList: React.FunctionComponent<Props> = ({
    thread,
    xchangeset,
    threadSettings,
    ...props
}) => (
    <div className="changeset-files-list">
        {xchangeset.repositoryComparisons.map((c, i) => (
            <RepositoryCompareDiffPage
                key={i}
                {...props}
                repo={c.baseRepository}
                base={{
                    repoName: c.baseRepository.name,
                    repoID: c.baseRepository.id,
                    rev: c.range.baseRevSpec.expr,
                    commitID: c.range.baseRevSpec.object!.oid,
                }}
                head={{
                    repoName: c.headRepository.name,
                    repoID: c.headRepository.id,
                    rev: c.range.headRevSpec.expr,
                    commitID: c.range.headRevSpec.object!.oid,
                }}
            />
        ))}
    </div>
)
