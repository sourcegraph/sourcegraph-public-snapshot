import React from 'react'

import { TemplateBanner } from './TemplateBanner'

interface SearchTemplatesBannerProps {
    className?: string
}

export const SearchTemplatesBanner: React.FunctionComponent<SearchTemplatesBannerProps> = ({ className }) => (
    <TemplateBanner
        heading="You are creating a Batch Change from a Code Search"
        description="Let Sourcegraph help you refactor your code by preparing a Batch Change from your search query"
        className={className}
    />
)
