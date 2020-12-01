import React, { useCallback, useState } from 'react'
import FileDownloadIcon from 'mdi-react/FileDownloadIcon'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { getChangesetDiff } from '../backend'
import { asError } from '../../../../../../shared/src/util/errors'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'

export interface DownloadDiffButtonProps {
    changesetID: string
}

enum DownloadState {
    Ready,
    Loading,
    Error,
}

export const DownloadDiffButton: React.FunctionComponent<DownloadDiffButtonProps> = ({ changesetID }) => {
    const [error, setError] = useState<Error | null>(null)
    const [state, setState] = useState(DownloadState.Ready)

    const loadDiff = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        async event => {
            event.preventDefault()

            if (!state) {
                setState(DownloadState.Loading)
                setError(null)

                try {
                    const diff = await getChangesetDiff(changesetID)
                    setState(DownloadState.Ready)

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
                    setError(asError(error))
                    setState(DownloadState.Error)
                }
            }
        },
        [changesetID, state]
    )

    let icon: JSX.Element | undefined
    switch (state) {
        case DownloadState.Ready:
            icon = <FileDownloadIcon className="icon-inline" />
            break
        case DownloadState.Loading:
            icon = <LoadingSpinner />
            break
        case DownloadState.Error:
            icon = <AlertCircleIcon className="icon icon-inline" data-tooltip={error?.message} />
            break
    }

    return (
        <div className="download-diff-button p-2">
            <button type="button" className="btn btn-icon" aria-label="Download diff" onClick={loadDiff}>
                {icon}
                <span className="pl-1">Download diff</span>
            </button>
        </div>
    )
}
