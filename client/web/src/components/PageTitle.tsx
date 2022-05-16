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

            // This is a fallback, in case the next page does *not* set the title.
            // Ideally, we should always overwrite this, so we don't announce it to screen readers to reduce noise.
            document.title = getBrandName()
        }
    }, [title])

    return null
}
