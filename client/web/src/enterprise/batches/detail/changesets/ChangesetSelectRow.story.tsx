import { storiesOf } from '@storybook/react'
import React from 'react'

import { EnterpriseWebStory } from '../../../components/EnterpriseWebStory'

import { ChangesetSelectRow } from './ChangesetSelectRow'

const { add } = storiesOf('web/batches/ChangesetSelectRow', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

const onSubmit = (): void => {}

add('all states', () => (
    <EnterpriseWebStory>
        {props => (
            <>
                <ChangesetSelectRow
                    {...props}
                    onSubmit={onSubmit}
                    selected={new Set(['id-1', 'id-2'])}
                    isSubmitting={false}
                />
                <hr />
                <ChangesetSelectRow
                    {...props}
                    onSubmit={onSubmit}
                    selected={new Set(['id-1', 'id-2'])}
                    isSubmitting={true}
                />
                <hr />
                <ChangesetSelectRow
                    {...props}
                    onSubmit={onSubmit}
                    selected={new Set(['id-1', 'id-2'])}
                    isSubmitting={new Error('something went wrong with the backend :(')}
                />
            </>
        )}
    </EnterpriseWebStory>
))
