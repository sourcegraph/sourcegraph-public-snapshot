import { ChangesetExternalState, ChangesetCheckState, ChangesetReviewState } from '../../graphql-operations'
import { HoveredToken } from '@sourcegraph/codeintellify'
import {
    RepoSpec,
    RevisionSpec,
    FileSpec,
    ResolvedRevisionSpec,
    UIPositionSpec,
    ModeSpec,
} from '../../../../shared/src/util/url'
import { getModeFromPath } from '../../../../shared/src/languages'

export function isValidChangesetExternalState(input: string): input is ChangesetExternalState {
    return Object.values<string>(ChangesetExternalState).includes(input)
}

export function isValidChangesetReviewState(input: string): input is ChangesetReviewState {
    return Object.values<string>(ChangesetReviewState).includes(input)
}

export function isValidChangesetCheckState(input: string): input is ChangesetCheckState {
    return Object.values<string>(ChangesetCheckState).includes(input)
}

export function getLSPTextDocumentPositionParameters(
    hoveredToken: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec
): RepoSpec & RevisionSpec & ResolvedRevisionSpec & FileSpec & UIPositionSpec & ModeSpec {
    return {
        repoName: hoveredToken.repoName,
        revision: hoveredToken.revision,
        filePath: hoveredToken.filePath,
        commitID: hoveredToken.commitID,
        position: hoveredToken,
        mode: getModeFromPath(hoveredToken.filePath || ''),
    }
}
