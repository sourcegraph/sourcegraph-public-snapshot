import { Meta, Story } from '@storybook/react'
import { startCase } from 'lodash'
import SearchIcon from 'mdi-react/SearchIcon'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { ButtonLink, Typography } from '..'
import { BUTTON_VARIANTS } from '../Button/constants'
import { Grid } from '../Grid'
import { Icon } from '../Icon'

const Config: Meta = {
    title: 'wildcard/ButtonLink',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: ButtonLink,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
        design: [
            {
                type: 'figma',
                name: 'Figma Light',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A2513',
            },
            {
                type: 'figma',
                name: 'Figma Dark',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A5794',
            },
        ],
    },
}

export default Config

export const Overview: Story = () => (
    <>
        <Typography.H1>ButtonLink</Typography.H1>
        <Typography.H2>Variants</Typography.H2>
        <Grid className="mb-3" columnCount={3}>
            {BUTTON_VARIANTS.map(variant => (
                <div key={variant}>
                    <ButtonLink variant={variant} to="https://sourcegraph.com" target="_blank" onClick={console.log}>
                        {startCase(variant)}
                    </ButtonLink>
                </div>
            ))}
        </Grid>
        <Typography.H2>Outline</Typography.H2>
        <ButtonLink
            variant="danger"
            outline={true}
            to="https://sourcegraph.com"
            target="_blank"
            onClick={console.log}
            className="mb-2"
        >
            Outline
        </ButtonLink>
        <Typography.H2>Icons</Typography.H2>
        <p>We can use icons with our buttons.</p>{' '}
        <ButtonLink
            variant="secondary"
            to="https://sourcegraph.com"
            target="_blank"
            onClick={console.log}
            className="mb-2"
        >
            <Icon role="img" aria-hidden={true} as={SearchIcon} className="mr-1" />
            Search
        </ButtonLink>
        <Typography.H2>Smaller</Typography.H2>
        <p>We can make our buttons smaller.</p>
        <ButtonLink
            variant="secondary"
            to="https://sourcegraph.com"
            target="_blank"
            onClick={console.log}
            className="mb-2"
            size="sm"
        >
            Smaller
        </ButtonLink>
    </>
)
