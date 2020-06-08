import { RepoSpec, RevisionSpec } from '../../../../shared/src/util/url'

export interface DiffResolvedRevSpec {
    baseCommitID: string
    headCommitID: string
}

export interface OpenInSourcegraphProps extends RepoSpec, RevisionSpec {
    sourcegraphURL: string
    filePath?: string
    commit?: {
        baseRev: string
        headRev: string
    }
    coords?: {
        line: number
        char: number
    }
    fragment?: 'references'
    query?: {
        search?: string
        diff?: {
            revision: string
        }
    }
    withModifierKey?: boolean
}

export interface OpenDiffInSourcegraphProps
    extends Pick<OpenInSourcegraphProps, Exclude<keyof OpenInSourcegraphProps, 'commit'>> {
    commit: {
        baseRev: string
        headRev: string
    }
}
