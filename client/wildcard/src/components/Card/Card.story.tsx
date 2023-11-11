import type { Meta, StoryFn } from '@storybook/react'

import { H1, H2, H3, Text } from '..'
import { BrandedStory } from '../../stories/BrandedStory'
import { Button } from '../Button'
import { Grid } from '../Grid'

import { Card, CardBody, CardFooter, CardHeader, CardSubtitle, CardText, CardTitle } from '.'

const config: Meta = {
    title: 'wildcard/Card',
    component: Card,

    decorators: [story => <BrandedStory>{() => <div className="container mt-3 pb-3">{story()}</div>}</BrandedStory>],

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

export const Simple: StoryFn = () => (
    <>
        <H1>Cards</H1>
        <Text>
            A card is a flexible and extensible content container. It includes options for headers and footers, a wide
            variety of content, contextual background colors, and powerful display options.{' '}
        </Text>

        <H2>Examples</H2>

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
                <Button variant="link">Edit</Button>
            </div>
        </CardBody>
    </Card>
)

export const InteractiveCard: StoryFn = () => (
    <>
        <H2>Interactive Cards</H2>
        {cardItem}

        <H3 className="mt-4">Cards List</H3>

        <div className="d-flex flex-column">
            {cardItem}
            {cardItem}
            {cardItem}
            {cardItem}
        </div>
    </>
)
