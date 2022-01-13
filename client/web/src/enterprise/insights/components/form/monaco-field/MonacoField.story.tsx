import { Meta } from '@storybook/react'
import RegexIcon from 'mdi-react/RegexIcon'
import React from 'react'

import { Button } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../../components/WebStory'

import { MonacoContainer, MonacoField } from './MonacoField'

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
        <MonacoField value="" placeholder="Example: type:diff repo:sourcegrph/* " />

        <MonacoField value="repo:github.com/sourcegraph/sourcegraph" />

        <MonacoField value={MULTI_LINE_VALUE} />

        <MonacoField value="repo:github.com/sourcegraph/sourcegraph" className="is-valid" />

        <MonacoField value="repo:github.com/sourcegraph/sourcegraph" className="is-invalid" />

        <MonacoField value="repo:github.com/sourcegraph/sourcegraph" className="is-invalid" />

        <MonacoContainer>
            <MonacoField value="" placeholder="Example: type:diff repo:sourcegrph/* " />
            <Button className="btn-icon" disabled={true}>
                <RegexIcon
                    size={21}
                    data-tooltip="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input."
                />
            </Button>
        </MonacoContainer>

        <MonacoContainer>
            <MonacoField value="repo:github.com/sourcegraph/sourcegraph" />
            <Button className="btn-icon" disabled={true}>
                <RegexIcon
                    size={21}
                    data-tooltip="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input."
                />
            </Button>
        </MonacoContainer>

        <MonacoContainer>
            <MonacoField value="repo:github.com/sourcegraph/sourcegraph" className="is-valid" />
            <Button className="btn-icon" disabled={true}>
                <RegexIcon
                    size={21}
                    data-tooltip="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input."
                />
            </Button>
        </MonacoContainer>

        <MonacoContainer>
            <MonacoField value="repo:github.com/sourcegraph/sourcegraph" className="is-invalid" />
            <Button className="btn-icon" disabled={true}>
                <RegexIcon
                    size={21}
                    data-tooltip="Regular expression is the only pattern type usable with capture groups and it’s enabled by default for this search input."
                />
            </Button>
        </MonacoContainer>
    </div>
)
