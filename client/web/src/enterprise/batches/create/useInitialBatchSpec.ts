import { useMemo } from 'react'

import { EditBatchChangeFields } from '../../../graphql-operations'

import helloWorldSample from './library/hello-world.batch.yaml'
import { insertNameIntoLibraryItem, isMinimalBatchSpec } from './yaml-util'

interface UseInitialBatchSpecResult {
    /** The latest batch spec for the batch change. */
    batchSpec: EditBatchChangeFields['currentSpec']
    /** Whether or not the latest batch spec has already been applied. */
    isApplied: boolean
    /** The raw YAML code that should initially populate the SSBC Monaco editor. */
    initialCode: string
}

/**
 * Custom hook for "CreateOrEdit" page which packages up business logic and exposes an API
 * for determining the latest batch spec created for a batch change and whether or not
 * that batch spec was already applied.
 *
 * @param batchChange The parent batch change to which the initial batch spec will belong.
 */
export const useInitialBatchSpec = (batchChange: EditBatchChangeFields): UseInitialBatchSpecResult => {
    const {
        currentSpec,
        batchSpecs: { nodes },
    } = batchChange

    // The first node from the batch specs is the latest batch spec for a batch change. If
    // it's different from the `currentSpec` on the batch change, that means the latest
    // batch spec has not yet been applied.
    const latest = nodes[0] || currentSpec
    // TODO: This should probably just be resolved on the backend as field on the
    // `BatchChange` from the GraphQL.
    const isLatestApplied = useMemo(() => currentSpec.id === latest.id, [currentSpec.id, latest])

    // Show the hello world sample code initially in the Monaco editor if the user hasn't
    // written any batch spec code yet, otherwise show the latest spec for the batch
    // change.
    const initialCode = useMemo(
        () =>
            isMinimalBatchSpec(latest.originalInput)
                ? insertNameIntoLibraryItem(helloWorldSample, batchChange.name)
                : latest.originalInput,
        [latest.originalInput, batchChange.name]
    )

    return {
        batchSpec: latest,
        isApplied: isLatestApplied,
        initialCode,
    }
}
