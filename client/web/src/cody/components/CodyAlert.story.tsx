import type { Meta, StoryFn, Decorator } from '@storybook/react'

import { Text, H2, Button } from '@sourcegraph/wildcard'

import { WebStory } from '../../../src/components/WebStory'

import { CodyAlert } from './CodyAlert'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/cody/PLG/CodyAlert',
    decorators: [decorator],
}

export default config

export const AlertWithActions: StoryFn = () => (
    <WebStory>
        {props => (
            <CodyAlert variant="purple">
                <H2>Join new Cody Pro team?</H2>
                <Text>
                    You've been invited to a new Cody Pro team by rob@biglike.com <br />
                    This will terminate your current Cody Pro plan, and place you on the new Cody Pro team. You will not
                    lose access to your Cody Pro benefits.
                </Text>
                <div className="mt-3">
                    <Button variant="primary" disabled={true} className="mr-3">
                        Accept
                    </Button>
                    <Button variant="link">Decline</Button>
                </div>
            </CodyAlert>
        )}
    </WebStory>
)

export const PurpulePro: StoryFn = () => (
    <WebStory>
        {props => (
            <CodyAlert variant="purple" displayCard={'CodyPro'}>
                <H2>Card with display card</H2>
                <Text>A success message</Text>
            </CodyAlert>
        )}
    </WebStory>
)

export const Purple: StoryFn = () => (
    <WebStory>
        {props => (
            <CodyAlert variant="purple">
                <H2>A Tile</H2>
                <Text>Purple success message</Text>
            </CodyAlert>
        )}
    </WebStory>
)

export const Green: StoryFn = () => (
    <WebStory>
        {props => (
            <CodyAlert variant="green">
                <H2>A Tile</H2>
                <Text>Success message</Text>
            </CodyAlert>
        )}
    </WebStory>
)

export const GreenPro: StoryFn = () => (
    <WebStory>
        {props => (
            <CodyAlert variant="green" displayCard={'CodyPro'}>
                <H2>A Tile</H2>
                <Text>Success message</Text>
            </CodyAlert>
        )}
    </WebStory>
)

export const Error: StoryFn = () => (
    <WebStory>
        {props => (
            <CodyAlert variant="error" displayCard={'Alert'}>
                <H2>A Tile</H2>
                <Text>A success message</Text>
            </CodyAlert>
        )}
    </WebStory>
)
