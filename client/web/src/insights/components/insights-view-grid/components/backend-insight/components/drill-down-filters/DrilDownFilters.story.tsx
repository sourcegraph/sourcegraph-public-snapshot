import { storiesOf } from '@storybook/react';
import React from 'react';

import { WebStory } from '../../../../../../../components/WebStory';

import { DrillDownFilters } from './DrillDownFilters';

const { add } = storiesOf('web/insights/DrillDownFilters', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

add('DrillDownFilters panel', () => <DrillDownFilters/>)
