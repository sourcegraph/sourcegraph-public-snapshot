import { GitCommitFields, RepositoryType } from '../graphql-operations'

export const isPerforceDepotSource = (sourceType: string): boolean => sourceType === RepositoryType.PERFORCE_DEPOT

export const getRefType = (sourceType: RepositoryType | string): string =>
    isPerforceDepotSource(sourceType) ? 'changelist' : 'commit'

export const getCanonicalURL = (sourceType: RepositoryType | string, node: GitCommitFields): string =>
    window.context.experimentalFeatures?.perforceChangelistMapping === 'enabled' &&
    isPerforceDepotSource(sourceType) &&
    node.perforceChangelist
        ? node.perforceChangelist.canonicalURL
        : node.canonicalURL
