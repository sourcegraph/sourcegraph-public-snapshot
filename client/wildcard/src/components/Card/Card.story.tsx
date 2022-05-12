import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Typography } from '..'
import { Button } from '../Button'
import { Grid } from '../Grid'

import { Card, CardBody, CardFooter, CardHeader, CardSubtitle, CardText, CardTitle } from '.'

const config: Meta = {
    title: 'wildcard/Card',
    component: Card,

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3 pb-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: Card,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
        design: [
            {
                type: 'figma',
                name: 'Figma Light',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=1172%3A285',
            },
            {
                type: 'figma',
                name: 'Figma Dark',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=1172%3A558',
            },
        ],
    },
}

export default config

export const Simple: Story = () => (
    <>
        <Typography.H1>Cards</Typography.H1>
        <p>
            A card is a flexible and extensible content container. It includes options for headers and footers, a wide
            variety of content, contextual background colors, and powerful display options.{' '}
        </p>

        <Typography.H2>Examples</Typography.H2>

        <Grid className="mb-3" columnCount={1}>
            <Card>
                <CardBody>This is some text within a card body.</CardBody>
            </Card>

            <Card>
                <CardBody>
                    <CardTitle>Card title</CardTitle>
                    <CardSubtitle>Card subtitle</CardSubtitle>
                    <CardText>
                        Some quick example text to build on the card title and make up the bulk of the card's content.
                    </CardText>
                    <Button variant="primary">Do something</Button>
                </CardBody>
                <CardFooter>Card footer</CardFooter>
            </Card>

            <Card>
                <CardHeader>Featured</CardHeader>
                <CardBody>
                    <CardTitle>Special title treatment</CardTitle>
                    <CardText>With supporting text below as a natural lead-in to additional content.</CardText>
                    <Button variant="primary">Do something</Button>
                </CardBody>
                <CardFooter>Card footer</CardFooter>
            </Card>
        </Grid>
    </>
)

const cardItem = (
    <Card as="button" className="mb-1 p-0 w-100">
        <CardBody className="w-100 d-flex justify-content-between align-items-center">
            <div className="d-flex flex-column">
                <CardTitle className="mb-0 text-left">Watch for secrets in new commits</CardTitle>
                <CardSubtitle>New search result → Sends email notifications, delivers webhook</CardSubtitle>
            </div>
            <div className="d-flex align-items-center">
                <Toggle
                    display="inline"
                    onClick={() => {}}
                    value={true}
                    className="mr-3 align-item-baseline"
                    disabled={false}
                />
                <Button variant="link">Edit</Button>
            </div>
        </CardBody>
    </Card>
)

export const InteractiveCard: Story = () => (
    <>
        <Typography.H2>Interactive Cards</Typography.H2>
        {cardItem}

        <Typography.H3 className="mt-4">Cards List</Typography.H3>

        <div className="d-flex flex-column">
            {cardItem}
            {cardItem}
            {cardItem}
            {cardItem}
        </div>
    </>
)
