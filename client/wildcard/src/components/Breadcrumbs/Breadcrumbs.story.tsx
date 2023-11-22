import type { Meta, StoryFn } from '@storybook/react'
import { ResizableBox } from 'react-resizable'

import { BrandedStory } from '../../stories'

import { Breadcrumbs } from './Breadcrumbs'

const config: Meta = {
    title: 'wildcard/Badge',
    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],
}

export default config

export const BreadcrumbsGallery: StoryFn = () => (
    <ResizableBox width={500} height={30} axis="x" minConstraints={[100, 0]}>
        <Breadcrumbs
            filename="sourcegraph/client/web/src/enteprise/insights/components/insight-view/components/InsighBackend.tsx"
            getSegmentLink={segment => `https://www.google.com/search?q=${segment}`}
        />
    </ResizableBox>
)
