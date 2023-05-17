import { findIndex } from 'lodash'

import { AuthenticatedUser } from '../auth'
import { RepoMetadataWritePermission } from '../rbac/constants'

export const canWriteRepoMetadata = (user: Pick<AuthenticatedUser, 'permissions'> | null): boolean =>
    !!user &&
    findIndex(user.permissions.nodes, permission => permission.displayName === RepoMetadataWritePermission) !== -1

export const NO_ACCESS_REPO_METADATA_WRITE =
    'Not sufficient permissions to edit repository metadata. Contact your site admin to request access.'
