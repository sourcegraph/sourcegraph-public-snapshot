import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../WebStory'

import { DiffStat, DiffStatStack } from './DiffStat'

const { add } = storiesOf('web/diffs/DiffStat', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('Collapsed', () => (
    <WebStory>
        {() => (
            <div>
                <DiffStat added={10} changed={4} deleted={8} />
            </div>
        )}
    </WebStory>
))

add('Expanded', () => (
    <WebStory>
        {() => (
            <div>
                <DiffStat added={10} changed={4} deleted={8} expandedCounts={true} />
            </div>
        )}
    </WebStory>
))

add('DiffStatStack', () => <WebStory>{() => <DiffStatStack added={10} changed={4} deleted={8} />}</WebStory>)
