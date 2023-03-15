import React, { useCallback, useState } from 'react'

import { mdiAlertCircle, mdiDownload } from '@mdi/js'

import { asError, isErrorLike } from '@sourcegraph/common'
import { Button, LoadingSpinner, Icon, Tooltip } from '@sourcegraph/wildcard'

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
        },
        [changesetID]
    )

    let icon: JSX.Element
    if (isErrorLike(state)) {
        icon = (
            <Tooltip content={state?.message}>
                <Icon aria-label={state?.message} className="icon" svgPath={mdiAlertCircle} />
            </Tooltip>
        )
    } else if (state === DownloadState.LOADING) {
        icon = <LoadingSpinner />
    } else {
        icon = <Icon aria-hidden={true} svgPath={mdiDownload} />
    }

    return (
        <Tooltip content="This is the changeset diff created when src batch preview|apply executed the batch change">
            <Button
                className="mb-1"
                onClick={loadDiff}
                disabled={state === DownloadState.LOADING}
                outline={true}
                variant="secondary"
                size="sm"
            >
                {icon}
                <span className="pl-1">Download generated diff</span>
            </Button>
        </Tooltip>
    )
}
