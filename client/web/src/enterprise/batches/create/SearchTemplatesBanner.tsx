import React from 'react'

import { TemplateBanner } from './TemplateBanner'

export const SearchTemplatesBanner: React.FunctionComponent = () => (
    <TemplateBanner
        heading="You are creating a Batch Change from a Code Search"
        description="Let Sourcegraph help you refactor your code by preparing a Batch Change from your search query"
    />
)
