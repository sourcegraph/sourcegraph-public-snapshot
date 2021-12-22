import FolderIcon from 'mdi-react/FolderIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React from 'react'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { PageHeader } from '@sourcegraph/wildcard'

import {
    PrimaryComponentForTreeFields,
    RepositoryForTreeFields,
    TreeEntryForTreeFields,
} from '../../../../graphql-operations'
import { ComponentAncestorsPath } from '../../components/catalog-area-header/CatalogAreaHeader'
import { componentIconComponent } from '../../components/ComponentIcon'
import { catalogPagePathForComponent } from '../../pages/component/ComponentDetailContent'

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
                description={primaryComponent.description}
            />
            <ComponentAncestorsPath path={catalogPagePathForComponent(primaryComponent).slice(0, -1)} />
        </>
    )
}
