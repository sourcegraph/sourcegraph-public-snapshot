import { Meta } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Typography } from '..'

import { PRODUCT_STATUSES } from './constants'
import { ProductStatusBadge } from './ProductStatusBadge'

const config: Meta = {
    title: 'wildcard/ProductStatusBadge',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: ProductStatusBadge,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
        design: [
            {
                type: 'figma',
                name: 'Figma Light',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A6149',
            },

            {
                type: 'figma',
                name: 'Figma Dark',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A6447',
            },
        ],
    },
}

export default config

export const Badges = () => (
    <>
        <Typography.H1>Product status badges</Typography.H1>
        <p>
            We often want to label different parts of our products with badges to ensure they are accurately presented
            to users.
        </p>
        {PRODUCT_STATUSES.map(status => (
            <ProductStatusBadge key={status} status={status} className="mr-2" />
        ))}
        <Typography.H2 className="mt-4">Linked product status badges</Typography.H2>
        <p>
            In some cases, we will want to automatically link to a relevant docs page for a particular status. This is
            also possible!
        </p>
        <ProductStatusBadge status="beta" linkToDocs={true} className="mr-3" />
        <ProductStatusBadge status="experimental" linkToDocs={true} className="mr-3" />
    </>
)
