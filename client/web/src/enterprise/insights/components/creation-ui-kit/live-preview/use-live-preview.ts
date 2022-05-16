import { useEffect, useState } from 'react'

import { asError } from '@sourcegraph/common'
import { useDebounce } from '@sourcegraph/wildcard'

interface Input<D> {
    disabled: boolean
    fetcher: () => Promise<D>
}

interface Output<D> {
    state: State<D>
    update: () => void
}

export enum StateStatus {
    Intact,
    Loading,
    Error,
    Data,
}

type State<D> =
    | { status: StateStatus.Data; data: D }
    | { status: StateStatus.Error; error: Error }
    | { status: StateStatus.Loading }
    | { status: StateStatus.Intact }

export function useLivePreview<D>(input: Input<D>): Output<D> {
    const [lastPreviewVersion, setLastPreviewVersion] = useState(0)
    const [state, setState] = useState<State<D>>({ status: StateStatus.Intact })

    // Synthetic deps to trigger dry run for fetching live preview data
    const debouncedInput = useDebounce(input, 500)

    useEffect(() => {
        let hasRequestCanceled = false
        const { fetcher, disabled } = debouncedInput

        if (!disabled) {
            setState({ status: StateStatus.Loading })

            fetcher()
                .then(data => !hasRequestCanceled && setState({ status: StateStatus.Data, data }))
                .catch(error => !hasRequestCanceled && setState({ status: StateStatus.Error, error: asError(error) }))
        } else {
            setState({ status: StateStatus.Intact })
        }

        return () => {
            hasRequestCanceled = true
        }
    }, [debouncedInput, lastPreviewVersion])

    return {
        state,
        update: () => setLastPreviewVersion(count => count + 1),
    }
}
