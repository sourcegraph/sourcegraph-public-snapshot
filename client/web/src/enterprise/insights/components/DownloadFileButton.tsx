import { FC, MouseEvent, useState } from 'react'

import { ButtonProps } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../components/LoaderButton'

interface DownloadFileButtonProps extends ButtonProps {
    fileUrl: string
    fileName: string
    children?: string
}

export const DownloadFileButton: FC<DownloadFileButtonProps> = props => {
    const { fileUrl, fileName, children, onClick, ...attributes } = props

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
            {...attributes}
            label={children}
            loading={isLoading}
            alwaysShowLabel={true}
            onClick={handleClick}
        />
    )
}

function syntheticDownload(url: string, name: string): void {
    const element = document.createElement('a')
    element.setAttribute('href', url)
    element.setAttribute('download', name)

    document.body.append(element)
    element.click()

    element.remove()
}
