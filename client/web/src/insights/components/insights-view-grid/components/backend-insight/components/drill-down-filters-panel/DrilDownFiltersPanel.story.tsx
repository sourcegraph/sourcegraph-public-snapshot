import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { DrillDownFiltersPanel } from './DrillDownFiltersPanel'

const { add } = storiesOf('web/insights/DrillDownFilters', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

add('DrillDownFilters panel', () => <DrillDownFiltersPanel />)
