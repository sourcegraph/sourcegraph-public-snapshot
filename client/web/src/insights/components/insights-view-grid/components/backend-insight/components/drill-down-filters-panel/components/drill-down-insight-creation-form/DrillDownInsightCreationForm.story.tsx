import { storiesOf } from '@storybook/react';
import React from 'react';

import { WebStory } from '../../../../../../../../../components/WebStory';

import { DrillDownInsightCreationForm } from './DrillDownInsightCreationForm';

const { add } = storiesOf('web/insights/DrillDownInsightCreationForm', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

add('DrillDownInsightCreationForm', () =>
    <DrillDownInsightCreationForm
        onCreateInsight={() => {} }
        onReset={() => {}}
    />
)
