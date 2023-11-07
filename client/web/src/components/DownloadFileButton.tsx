import { forwardRef, type MouseEvent, type ReactNode, useState } from 'react'

import type { ButtonProps } from '@sourcegraph/wildcard'

import { LoaderButton } from './LoaderButton'

interface DownloadFileButtonProps extends ButtonProps {
    fileUrl: string
    fileName?: string
    alwaysShowLabel?: boolean
    debounceTime?: number
    children?: ReactNode
}

export const DownloadFileButton = forwardRef<HTMLButtonElement, DownloadFileButtonProps>((props, ref) => {
    const {
        fileUrl,
        fileName,
        children,
        alwaysShowLabel = true,
        debounceTime = 0,
        disabled,
        onClick,
        ...attributes
    } = props

    const [isLoading, setLoading] = useState(false)

    const handleClick = async (event: MouseEvent<HTMLButtonElement>): Promise<void> => {
        setLoading(true)

        try {
            const debouncePromise =
                debounceTime === 0 ? Promise.resolve() : new Promise(resolve => setTimeout(resolve, debounceTime))

            const [file] = await Promise.all([fetch(fileUrl, { headers: window.context.xhrHeaders }), debouncePromise])
            const headerFileName = getHeaderFileName(file.headers.get('Content-Disposition') ?? '')
            const fileBlob = await file.blob()
            const url = URL.createObjectURL(fileBlob)

            syntheticDownload(url, fileName ?? headerFileName ?? getFileNameFromURL(fileUrl))

            if (onClick) {
                onClick(event)
            }
        } finally {
            setLoading(false)
        }
    }

    return (
        <LoaderButton
            ref={ref}
            {...attributes}
            label={children}
            loading={isLoading}
            disabled={isLoading || disabled}
            alwaysShowLabel={alwaysShowLabel}
            onClick={handleClick}
        />
    )
})
DownloadFileButton.displayName = 'DownloadFileButton'

function syntheticDownload(url: string, name: string): void {
    const element = document.createElement('a')
    element.setAttribute('href', url)
    element.setAttribute('download', name)

    document.body.append(element)
    element.click()

    element.remove()
}

function getHeaderFileName(disposition: string): string | null {
    const utf8FilenameRegex = /filename\*=utf-8''([\w%.]+)(?:; ?|$)/i
    const asciiFilenameRegex = /^filename=(["']?)(.*?[^\\])\1(?:; ?|$)/i

    let fileName = null
    if (utf8FilenameRegex.test(disposition)) {
        fileName = decodeURIComponent(utf8FilenameRegex.exec(disposition)![1])
    } else {
        // prevent ReDos attacks by anchoring the ascii regex to string start and
        //  slicing off everything before 'filename='
        const filenameStart = disposition.toLowerCase().indexOf('filename=')
        if (filenameStart >= 0) {
            const partialDisposition = disposition.slice(filenameStart)
            const matches = asciiFilenameRegex.exec(partialDisposition)
            if (matches?.[2]) {
                fileName = matches[2]
            }
        }
    }

    return fileName
}

const getFileNameFromURL = (url: string | null): string => {
    if (url === null) {
        return ''
    }

    const parts = url.split('/')
    return parts.at(-1)!
}
