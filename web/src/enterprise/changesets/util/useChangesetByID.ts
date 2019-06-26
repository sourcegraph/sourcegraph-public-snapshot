import { useState } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { fetchDiscussionThreadAndComments } from '../../../discussions/backend'
import { useEffectAsync } from '../../../util/useEffectAsync'

const LOADING: 'loading' = 'loading'

export function useChangesetByID(
    threadID: GQL.ID
): [typeof LOADING | GQL.IDiscussionThread | ErrorLike, (updated: GQL.IDiscussionThread) => void] {
    const [changesetOrError, setChangesetOrError] = useState<typeof LOADING | GQL.IDiscussionThread | ErrorLike>(
        LOADING
    )
    useEffectAsync(async () => {
        try {
            // TODO!(sqs)
            setChangesetOrError(await fetchDiscussionThreadAndComments(threadID).toPromise())
        } catch (err) {
            setChangesetOrError(asError(err))
        }
    }, [threadID])
    return [changesetOrError, setChangesetOrError]
}
