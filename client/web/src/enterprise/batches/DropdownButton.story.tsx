import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'

import { BATCH_CHANGES_SITE_CONFIGURATION } from './backend'
import { type Action, DropdownButton } from './DropdownButton'
import { rolloutWindowConfigMockResult, noRolloutWindowMockResult } from './mocks'

// eslint-disable-next-line @typescript-eslint/require-await
const onTrigger = async (onDone: () => void) => onDone()

const action: Action = {
    type: 'action-type',
    buttonLabel: 'Action',
    dropdownTitle: 'Action',
    dropdownDescription: 'Perform an action',
    onTrigger,
}

const disabledAction: Action = {
    type: 'disabled-action-type',
    buttonLabel: 'Disabled action',
    disabled: true,
    dropdownTitle: 'Disabled action',
    dropdownDescription: 'Perform an action, if only this were enabled',
    onTrigger,
}

const experimentalAction: Action = {
    type: 'experimental-action-type',
    buttonLabel: 'Experimental action',
    dropdownTitle: 'Experimental action',
    dropdownDescription: 'Perform a super cool action that might explode',
    onTrigger,
    experimental: true,
}

const publishAction: Action = {
    type: 'publish',
    buttonLabel: 'Publish Changeset',
    dropdownTitle: 'Publish Changeset',
    dropdownDescription: 'Attempt to publish all changesets to the code hosts.',
    onTrigger,
    experimental: false,
}

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/DropdownButton',
    decorators: [decorator],
    argTypes: {
        disabled: {
            control: { type: 'boolean' },
        },
    },
    args: {
        disabled: false,
    },
}

export default config

export const NoActions: StoryFn = args => <WebStory>{() => <DropdownButton actions={[]} {...args} />}</WebStory>
NoActions.argTypes = {
    disabled: {
        table: {
            disable: true,
        },
    },
}

NoActions.storyName = 'No actions'

export const SingleAction: StoryFn = args => (
    <WebStory>{() => <DropdownButton actions={[action]} {...args} />}</WebStory>
)

SingleAction.storyName = 'Single action'

export const MultipleActionsWithoutDefault: StoryFn = args => (
    <WebStory>
        {() => (
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(BATCH_CHANGES_SITE_CONFIGURATION),
                        },
                        result: noRolloutWindowMockResult,
                    },
                ]}
            >
                <DropdownButton actions={[action, disabledAction, experimentalAction, publishAction]} {...args} />
            </MockedTestProvider>
        )}
    </WebStory>
)

MultipleActionsWithoutDefault.storyName = 'Multiple actions without default'

export const MultipleActionsWithDefault: StoryFn = args => (
    <WebStory>
        {() => (
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(BATCH_CHANGES_SITE_CONFIGURATION),
                        },
                        result: noRolloutWindowMockResult,
                    },
                ]}
            >
                <DropdownButton actions={[action, disabledAction, experimentalAction]} defaultAction={0} {...args} />
            </MockedTestProvider>
        )}
    </WebStory>
)

MultipleActionsWithDefault.storyName = 'Multiple actions with default'

export const PublishActionWithRolloutWindowConfigured: StoryFn = args => (
    <WebStory>
        {() => (
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(BATCH_CHANGES_SITE_CONFIGURATION),
                        },
                        result: rolloutWindowConfigMockResult,
                    },
                ]}
            >
                <DropdownButton actions={[action, publishAction]} defaultAction={0} {...args} />
            </MockedTestProvider>
        )}
    </WebStory>
)

PublishActionWithRolloutWindowConfigured.storyName = 'Publish Action with rollout window configured'
