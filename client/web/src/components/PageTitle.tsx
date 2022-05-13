import React, { useEffect, useCallback } from 'react'

import { screenReaderAnnounce } from '@sourcegraph/wildcard'

interface PageTitleProps {
    title?: string
}

let titleSet = false

export const PageTitle: React.FunctionComponent<PageTitleProps> = ({ title }) => {
    const getBrandName = useCallback(() => {
        if (!window.context) {
            return 'Sourcegraph'
        }
        const { branding } = window.context
        return branding ? branding.brandName : 'Sourcegraph'
    }, [])

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
    }, [getBrandName, title])

    return null
}
