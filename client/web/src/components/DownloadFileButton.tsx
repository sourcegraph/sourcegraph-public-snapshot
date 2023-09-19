import { forwardRef, type MouseEvent, ReactNode, useState } from 'react'

import type { ButtonProps } from '@sourcegraph/wildcard'

import { LoaderButton } from './LoaderButton'

interface DownloadFileButtonProps extends ButtonProps {
    fileUrl: string
    fileName: string
    alwaysShowLabel?: boolean
    withLoading?: boolean
    children?: ReactNode
}

export const DownloadFileButton = forwardRef<HTMLButtonElement, DownloadFileButtonProps>((props, ref) => {
    const { fileUrl, fileName, children, alwaysShowLabel = true, withLoading = true, onClick, ...attributes } = props

    const [isLoading, setLoading] = useState(false)

    const handleClick = async (event: MouseEvent<HTMLButtonElement>): Promise<void> => {
        setLoading(true)

        try {
            const file = await fetch(fileUrl, { headers: window.context.xhrHeaders })
            const fileBlob = await file.blob()
            const url = URL.createObjectURL(fileBlob)

            syntheticDownload(url, fileName)

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
            loading={withLoading && isLoading}
            alwaysShowLabel={alwaysShowLabel}
            onClick={handleClick}
        />
    )
})

function syntheticDownload(url: string, name: string): void {
    const element = document.createElement('a')
    element.setAttribute('href', url)
    element.setAttribute('download', name)

    document.body.append(element)
    element.click()

    element.remove()
}
