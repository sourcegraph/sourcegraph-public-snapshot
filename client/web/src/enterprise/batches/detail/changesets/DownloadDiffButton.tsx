import React, { useCallback, useState } from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import DownloadIcon from 'mdi-react/DownloadIcon'

import { asError, isErrorLike } from '@sourcegraph/common'
import { Button, LoadingSpinner, Icon } from '@sourcegraph/wildcard'

import { getChangesetDiff } from '../backend'

export interface DownloadDiffButtonProps {
    changesetID: string
}

enum DownloadState {
    READY,
    LOADING,
}

type State = DownloadState | Error

export const DownloadDiffButton: React.FunctionComponent<React.PropsWithChildren<DownloadDiffButtonProps>> = ({
    changesetID,
}) => {
    const [state, setState] = useState<State>(DownloadState.READY)

    const loadDiff = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        async event => {
            event.preventDefault()

            if (!state) {
                setState(DownloadState.LOADING)

                try {
                    const diff = await getChangesetDiff(changesetID)
                    setState(DownloadState.READY)

                    // Create a URL that we can "click" on behalf of the user to
                    // prompt them to download the diff.
                    const blob = new Blob([diff], {
                        type: 'text/x-diff',
                    })
                    const url = URL.createObjectURL(blob)

                    try {
                        const link = document.createElement('a')
                        link.href = url
                        link.style.display = 'none'
                        link.download = `${changesetID}.diff`
                        document.body.append(link)
                        link.click()
                        link.remove()
                    } finally {
                        URL.revokeObjectURL(url)
                    }
                } catch (error) {
                    setState(asError(error))
                }
            }
        },
        [changesetID, state]
    )

    let icon: JSX.Element
    if (isErrorLike(state)) {
        icon = <Icon className="icon" data-tooltip={state?.message} as={AlertCircleIcon} />
    } else if (state === DownloadState.LOADING) {
        icon = <LoadingSpinner />
    } else {
        icon = <Icon as={DownloadIcon} />
    }

    return (
        <Button
            className="mb-1"
            aria-label="Download generated diff"
            data-tooltip="This is the changeset diff created when src batch preview|apply executed the batch change"
            onClick={loadDiff}
            disabled={state === DownloadState.LOADING}
            outline={true}
            variant="secondary"
            size="sm"
        >
            {icon}
            <span className="pl-1">Download generated diff</span>
        </Button>
    )
}
