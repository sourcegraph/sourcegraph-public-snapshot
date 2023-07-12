import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '../../stories'

import { Breadcrumbs } from './Breadcrumbs'

const config: Meta = {
    title: 'wildcard/Badge',
    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],
}

export default config

export const BreadcrumbsGallery: Story = () => (
    <div style={{ width: 500, overflow: 'hidden', border: '1px solid black'}}>
        <Breadcrumbs filename='sourcegraph/client/web/src/enteprise/insights/components/insight-view/components/InsighBackend.tsx'/>
    </div>
)
