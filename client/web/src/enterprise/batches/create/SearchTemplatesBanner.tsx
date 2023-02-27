import React from 'react'

import { TemplateBanner } from './TemplateBanner'

interface SearchTemplatesBannerProps {
    className?: string
}

export const SearchTemplatesBanner: React.FunctionComponent<SearchTemplatesBannerProps> = ({ className }) => (
    <TemplateBanner
        heading="You are creating a batch change from a code search"
        description="Let Sourcegraph help you refactor your code by preparing a batch change from your search query."
        className={className}
    />
)
