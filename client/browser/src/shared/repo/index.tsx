interface DiffRepoSpec {
    baseRepoName: string
    headRepoName: string
}

interface DiffRevSpec {
    baseRev: string
    headRev: string
}

export interface DiffResolvedRevSpec {
    baseCommitID: string
    headCommitID: string
}

export interface DiffRepoRev extends DiffRepoSpec, DiffRevSpec {}

export interface OpenInSourcegraphProps {
    repoName: string
    rev: string
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
            rev: string
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

export interface MaybeDiffSpec {
    isDelta: boolean
    isSplitDiff?: boolean
    isBase?: boolean
}

export interface CodeCell {
    cell: HTMLElement
    eventHandler: HTMLElement
    line: number
    isAddition?: boolean // for diff views
    isDeletion?: boolean // for diff views
}
