import { DecoratorFn, Meta } from '@storybook/react'
import ArrowRightIcon from 'mdi-react/ArrowRightIcon'
import React from 'react'

import { Link, Icon } from '@sourcegraph/wildcard'

import { WebStory } from '../WebStory'

import { MarketingBlock } from './MarketingBlock'

const decorator: DecoratorFn = story => <WebStory>{() => <div className="container mt-3">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/markering/MarketingBlock',
    decorators: [decorator],
}

export default config

export const Basic = (): JSX.Element => (
    <MarketingBlock>
        <h3 className="pr-3">Need help getting started?</h3>

        <div>
            <Link to="https://sourcegraph.com/search">
                Speak to an engineer
                <Icon className="ml-2" as={ArrowRightIcon} />
            </Link>
        </div>
    </MarketingBlock>
)
