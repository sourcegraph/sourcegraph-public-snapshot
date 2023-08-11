import { findIndex } from 'lodash'

import type { AuthenticatedUser } from '../auth'
import { RepoMetadataWritePermission } from '../rbac/constants'

export const canWriteRepoMetadata = (user: Pick<AuthenticatedUser, 'permissions'> | null): boolean =>
    !!user &&
    findIndex(user.permissions.nodes, permission => permission.displayName === RepoMetadataWritePermission) !== -1
