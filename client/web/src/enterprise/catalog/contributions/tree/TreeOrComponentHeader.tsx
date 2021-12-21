import FolderIcon from 'mdi-react/FolderIcon'
import React from 'react'
import { TelemetryProps, TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    PrimaryComponentForTreeFields,
    RepositoryForTreeFields,
    TreeOrComponentPageResult,
    TreeEntryForTreeFields,
} from '../../../../graphql-operations'
import { PageHeader } from '@sourcegraph/wildcard'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

import { componentIconComponent } from '../../components/ComponentIcon'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'

interface Props {
    repository: RepositoryForTreeFields
    tree: TreeEntryForTreeFields
    primaryComponent: PrimaryComponentForTreeFields | null
}

export const TreeOrComponentHeader: React.FunctionComponent<Props> = ({ repository, tree, primaryComponent }) => {
    const a = 1

    return (
        <>
            <PageHeader
                path={[
                    primaryComponent !== null
                        ? { icon: componentIconComponent(primaryComponent), text: primaryComponent.name }
                        : !tree.isRoot
                        ? { icon: FolderIcon, text: tree.path }
                        : { icon: SourceRepositoryIcon, text: displayRepoName(repository.name) },
                ]}
                className="mb-3 test-tree-page-title"
            />
        </>
    )
}
