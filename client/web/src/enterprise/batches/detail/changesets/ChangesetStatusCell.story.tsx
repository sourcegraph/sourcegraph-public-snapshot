import { Meta, DecoratorFn, Story } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetState } from '../../../../graphql-operations'

import { ChangesetStatusCell } from './ChangesetStatusCell'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>
const config: Meta = {
    title: 'web/batches/ChangesetStatusCell',
    decorators: [decorator],
}

export default config

export const Unpublished: Story = () => (
    <WebStory>
        {() => <ChangesetStatusCell state={ChangesetState.UNPUBLISHED} className="d-flex text-muted" />}
    </WebStory>
)

export const Failed: Story = () => (
    <WebStory>{() => <ChangesetStatusCell state={ChangesetState.FAILED} className="d-flex text-muted" />}</WebStory>
)

export const Retrying: Story = () => (
    <WebStory>{() => <ChangesetStatusCell state={ChangesetState.RETRYING} className="d-flex text-muted" />}</WebStory>
)

export const Scheduled: Story = () => (
    <WebStory>{() => <ChangesetStatusCell state={ChangesetState.SCHEDULED} className="d-flex text-muted" />}</WebStory>
)

export const Processing: Story = () => (
    <WebStory>{() => <ChangesetStatusCell state={ChangesetState.PROCESSING} className="d-flex text-muted" />}</WebStory>
)

export const Open: Story = () => (
    <WebStory>{() => <ChangesetStatusCell state={ChangesetState.OPEN} className="d-flex text-muted" />}</WebStory>
)

export const Draft: Story = () => (
    <WebStory>{() => <ChangesetStatusCell state={ChangesetState.DRAFT} className="d-flex text-muted" />}</WebStory>
)

export const Closed: Story = () => (
    <WebStory>{() => <ChangesetStatusCell state={ChangesetState.CLOSED} className="d-flex text-muted" />}</WebStory>
)

export const Merged: Story = () => (
    <WebStory>{() => <ChangesetStatusCell state={ChangesetState.MERGED} className="d-flex text-muted" />}</WebStory>
)

export const Deleted: Story = () => (
    <WebStory>{() => <ChangesetStatusCell state={ChangesetState.DELETED} className="d-flex text-muted" />}</WebStory>
)
