import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from 'src/components/WebStory'

import { DrillDownFiltersForm } from './DrillDownFiltersForm'

const { add } = storiesOf('web/insights/DrillDownFilters', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

add('DrillDownFiltersForm', () => <DrillDownFiltersForm />)
