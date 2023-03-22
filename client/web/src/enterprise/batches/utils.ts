import {
    ChangesetCheckState,
    ChangesetReviewState,
    ChangesetSpecOperation,
    ChangesetState,
} from '../../graphql-operations'

export function isValidChangesetReviewState(input: string): input is ChangesetReviewState {
    return Object.values<string>(ChangesetReviewState).includes(input)
}

export function isValidChangesetCheckState(input: string): input is ChangesetCheckState {
    return Object.values<string>(ChangesetCheckState).includes(input)
}

export function isValidChangesetSpecOperation(input: string): input is ChangesetSpecOperation {
    return Object.values<string>(ChangesetSpecOperation).includes(input)
}

export function isValidChangesetState(input: string): input is ChangesetState {
    return Object.values<string>(ChangesetState).includes(input)
}
