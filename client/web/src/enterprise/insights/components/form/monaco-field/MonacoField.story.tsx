import { mdiRegex } from '@mdi/js'
import { Meta } from '@storybook/react'

import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../../components/WebStory'

import * as Monaco from '.'

export default {
    title: 'web/insights/form/MonacoField',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta

const MULTI_LINE_VALUE = `
repo:^(github\\.com/sourcegraph/sourcegraph)$
type:diff after:2021-12-16T00:00:00+03:00
before:2021-12-23T00:00:00+03:00 database.NewDB
`.trim()

export const SimpleMonacoField = () => (
    <div className="d-flex flex-column" style={{ gap: '2rem', width: 400 }}>
        <Monaco.Field queryState={{ query: '' }} placeholder="Example: type:diff repo:sourcegraph/* " />

        <Monaco.Field queryState={{ query: 'repo:github.com/sourcegraph/sourcegraph' }} />

        <Monaco.Field queryState={{ query: MULTI_LINE_VALUE }} />

        <Monaco.Field queryState={{ query: 'repo:github.com/sourcegraph/sourcegraph' }} className="is-valid" />

        <Monaco.Field queryState={{ query: 'repo:github.com/sourcegraph/sourcegraph' }} className="is-invalid" />

        <Monaco.Field queryState={{ query: 'repo:github.com/sourcegraph/sourcegraph' }} className="is-invalid" />

        <Monaco.Root>
            <Monaco.Field queryState={{ query: '' }} placeholder="Example: type:diff repo:sourcegraph/* " />
            <Tooltip content="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input.">
                <Button variant="icon" disabled={true}>
                    <Icon svgPath={mdiRegex} aria-hidden={true} inline={false} height={21} width={21} />
                </Button>
            </Tooltip>
        </Monaco.Root>

        <Monaco.Root>
            <Monaco.Field queryState={{ query: 'repo:github.com/sourcegraph/sourcegraph' }} />
            <Tooltip content="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input.">
                <Button variant="icon" disabled={true}>
                    <Icon svgPath={mdiRegex} aria-hidden={true} inline={false} height={21} width={21} />
                </Button>
            </Tooltip>
        </Monaco.Root>

        <Monaco.Root>
            <Monaco.Field queryState={{ query: 'repo:github.com/sourcegraph/sourcegraph' }} className="is-valid" />
            <Tooltip content="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input.">
                <Button variant="icon" disabled={true}>
                    <Icon svgPath={mdiRegex} aria-hidden={true} inline={false} height={21} width={21} />
                </Button>
            </Tooltip>
        </Monaco.Root>

        <Monaco.Root>
            <Monaco.Field queryState={{ query: 'repo:github.com/sourcegraph/sourcegraph' }} className="is-invalid" />
            <Tooltip content="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input.">
                <Button variant="icon" disabled={true}>
                    <Icon svgPath={mdiRegex} aria-hidden={true} inline={false} height={21} width={21} />
                </Button>
            </Tooltip>
        </Monaco.Root>
    </div>
)
