import React, { useEffect } from 'react'

import { screenReaderAnnounce } from '@sourcegraph/wildcard'

interface PageTitleProps {
    title?: string
}

const getBrandName = (): string => {
    if (!window.context) {
        return 'Sourcegraph'
    }
    const { branding } = window.context
    return branding ? branding.brandName : 'Sourcegraph'
}

let titleSet = false

export const PageTitle: React.FunctionComponent<PageTitleProps> = ({ title }) => {
    useEffect(() => {
        if (titleSet) {
            console.error('more than one PageTitle used at the same time')
        }
        titleSet = true
        document.title = title ? `${title} - ${getBrandName()}` : getBrandName()
        screenReaderAnnounce(document.title)

        return () => {
            titleSet = false
            document.title = getBrandName()
        }
        // Only run once, on mount
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    return null
}
