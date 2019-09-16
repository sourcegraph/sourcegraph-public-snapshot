import { ChangesetState } from '../../../../../../shared/src/graphql/schema'
export const changesetStatusColorClasses: Record<ChangesetState, string> = {
    [ChangesetState.OPEN]: 'success',
    [ChangesetState.CLOSED]: 'danger',
    [ChangesetState.MERGED]: 'purple',
}
