import { Meta, Story } from '@storybook/react'
import { ResizableBox } from 'react-resizable'

import { BrandedStory } from '../../stories'

import { Breadcrumbs } from './Breadcrumbs'

const config: Meta = {
    title: 'wildcard/Badge',
    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],
}

export default config

export const BreadcrumbsGallery: Story = () => (
    <ResizableBox width={500} axis="x" minConstraints={[200, 0]}>
        <Breadcrumbs filename='sourcegraph/client/web/src/enteprise/insights/components/insight-view/components/InsighBackend.tsx'/>
    </ResizableBox>
)
