import type { Meta } from '@storybook/react'

import { H1, H2, Link, Text } from '..'
import { BrandedStory } from '../../stories/BrandedStory'

import { Badge } from '.'
import { BADGE_VARIANTS } from './constants'

const config: Meta = {
    title: 'wildcard/Badge',

    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],

    parameters: {
        component: Badge,
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
        <H1>Badges</H1>
        <H2>Variants</H2>
        <Text>Our badges can be styled with different variants.</Text>
        {BADGE_VARIANTS.map(variant => (
            <Badge key={variant} variant={variant} className="mr-2">
                {variant}
            </Badge>
        ))}
        <H2 className="mt-4">Size</H2>
        <Text>We can also make our badges smaller.</Text>
        <Badge small={true}>I am a small badge</Badge>
        <H2 className="mt-4">Pills</H2>
        <Text>Commonly used to display counts, we can style badges as pills.</Text>
        <Badge pill={true} variant="secondary">
            321+
        </Badge>
        <H2 className="mt-4">Links</H2>
        <Text>For more advanced functionality, badges can also function as links.</Text>
        {BADGE_VARIANTS.map(variant => (
            <Badge as={Link} to="https://example.com" key={variant} variant={variant} className="mr-2">
                link/{variant}
            </Badge>
        ))}
    </>
)
