import { Meta } from '@storybook/react'

import { Button, Icon } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../../components/WebStory'

import * as Monaco from '.'
import { mdiRegex } from "@mdi/js";

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
        <Monaco.Field value="" placeholder="Example: type:diff repo:sourcegraph/* " />

        <Monaco.Field value="repo:github.com/sourcegraph/sourcegraph" />

        <Monaco.Field value={MULTI_LINE_VALUE} />

        <Monaco.Field value="repo:github.com/sourcegraph/sourcegraph" className="is-valid" />

        <Monaco.Field value="repo:github.com/sourcegraph/sourcegraph" className="is-invalid" />

        <Monaco.Field value="repo:github.com/sourcegraph/sourcegraph" className="is-invalid" />

        <Monaco.Root>
            <Monaco.Field value="" placeholder="Example: type:diff repo:sourcegraph/* " />
            <Button variant="icon" disabled={true}>
                <Icon
                    data-tooltip="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input." svgPath={mdiRegex} inline={false} aria-hidden={true} height={21} width={21}
                />
            </Button>
        </Monaco.Root>

        <Monaco.Root>
            <Monaco.Field value="repo:github.com/sourcegraph/sourcegraph" />
            <Button variant="icon" disabled={true}>
                <Icon
                    data-tooltip="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input." svgPath={mdiRegex} inline={false} aria-hidden={true} height={21} width={21}
                />
            </Button>
        </Monaco.Root>

        <Monaco.Root>
            <Monaco.Field value="repo:github.com/sourcegraph/sourcegraph" className="is-valid" />
            <Button variant="icon" disabled={true}>
                <Icon
                    data-tooltip="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input." svgPath={mdiRegex} inline={false} aria-hidden={true} height={21} width={21}
                />
            </Button>
        </Monaco.Root>

        <Monaco.Root>
            <Monaco.Field value="repo:github.com/sourcegraph/sourcegraph" className="is-invalid" />
            <Button variant="icon" disabled={true}>
                <Icon
                    data-tooltip="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input." svgPath={mdiRegex} inline={false} aria-hidden={true} height={21} width={21}
                />
            </Button>
        </Monaco.Root>
    </div>
)
