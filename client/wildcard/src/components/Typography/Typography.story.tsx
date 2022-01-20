import { select } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { TYPOGRAPHY_ALIGNMENTS, TYPOGRAPHY_MODES } from './constants'
import { H1, H2, H3, H4, H5, H6 } from './Heading'
import { Label } from './Label'

import { Code, Text } from '.'

const decorator: DecoratorFn = story => <BrandedStory styles={webStyles}>{() => <div>{story()}</div>}</BrandedStory>

const config: Meta = {
    title: 'wildcard/Typography/All',

    decorators: [decorator],

    parameters: {
        component: Label,
        design: {
            type: 'figma',
            name: 'Figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=5601%3A65477',
        },
    },
}

export default config

export const Simple: Story = () => (
    <>
        <h2>Headings</h2>
        <table className="table">
            <tbody>
                <tr>
                    <td>
                        <code>
                            {'<'}
                            H1
                            {'>'}
                        </code>
                    </td>
                    <td>
                        <H1
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                        >
                            This is H1
                        </H1>
                    </td>
                </tr>
                <tr>
                    <td>
                        <code>
                            {'<'}
                            H2
                            {'>'}
                        </code>
                    </td>
                    <td>
                        <H2
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                        >
                            This is H2
                        </H2>
                    </td>
                </tr>
                <tr>
                    <td>
                        <code>
                            {'<'}
                            H3
                            {'>'}
                        </code>
                    </td>
                    <td>
                        <H3
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                        >
                            This is H3
                        </H3>
                    </td>
                </tr>
                <tr>
                    <td>
                        <code>
                            {'<'}
                            H4
                            {'>'}
                        </code>
                    </td>
                    <td>
                        <H4
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                        >
                            This is H4
                        </H4>
                    </td>
                </tr>
                <tr>
                    <td>
                        <code>
                            {'<'}
                            H5
                            {'>'}
                        </code>
                    </td>
                    <td>
                        <H5
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                        >
                            This is H5
                        </H5>
                    </td>
                </tr>
                <tr>
                    <td>
                        <code>
                            {'<'}
                            H6
                            {'>'}
                        </code>
                    </td>
                    <td>
                        <H6
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                        >
                            This is H6
                        </H6>
                    </td>
                </tr>
            </tbody>
        </table>

        <h2>Code</h2>
        <table className="table">
            <tbody>
                <tr>
                    <td>
                        <code>
                            {'<'}
                            Code
                            {'>'}
                        </code>
                    </td>
                    <td>
                        <div>
                            <Code size="base" weight="regular">
                                This is Code / Base / Regular
                            </Code>
                        </div>
                        <div>
                            <Code size="base" weight="bold">
                                This is Code / Base / Bold
                            </Code>
                        </div>
                        <div>
                            <Code size="small" weight="regular">
                                This is Code / Small / Regular
                            </Code>
                        </div>
                        <div>
                            <Code size="small" weight="bold">
                                This is Code / Small / Bold
                            </Code>
                        </div>
                    </td>
                </tr>
            </tbody>
        </table>

        <h2>Label</h2>
        <table className="table">
            <tbody>
                <tr>
                    <td>
                        <code>
                            {'<'}
                            Label
                            {'>'}
                        </code>
                    </td>
                    <td>
                        <div>
                            <Label
                                mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                                alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                                size="base"
                            >
                                This is Label / Base
                            </Label>
                        </div>
                        <div>
                            <Label
                                mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                                alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                                size="base"
                                isUnderline={true}
                            >
                                This is Label / Base - underline
                            </Label>
                        </div>
                        <div>
                            <Label
                                mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                                alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                                size="small"
                            >
                                This is Label / Small
                            </Label>
                        </div>
                        <div>
                            <Label
                                mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                                alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                                size="small"
                                isUnderline={true}
                            >
                                This is Label / Small - underline
                            </Label>
                        </div>
                        <div>
                            <Label
                                mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                                alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                                isUppercase={true}
                            >
                                This is Label / Uppercase / Base
                            </Label>
                        </div>
                        <div>
                            <Label
                                mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                                alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                                size="small"
                                isUppercase={true}
                            >
                                This is Label / Uppercase / Small
                            </Label>
                        </div>
                    </td>
                </tr>
            </tbody>
        </table>

        <h2>Text</h2>
        <table className="table">
            <tbody>
                <tr>
                    <td>
                        <code>
                            {'<'}
                            Text
                            {'>'}
                        </code>
                    </td>
                    <td>
                        <Text
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                            size="base"
                            weight="regular"
                        >
                            This is Body / Base / Regular
                        </Text>
                        <Text
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                            size="base"
                            weight="medium"
                        >
                            This is Body / Base / Medium
                        </Text>
                        <Text
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                            size="base"
                            weight="bold"
                        >
                            This is Body / Base / Bold
                        </Text>
                        <Text
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                            size="small"
                            weight="regular"
                        >
                            This is Body / Small / Regular
                        </Text>
                        <Text
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                            size="small"
                            weight="medium"
                        >
                            This is Body / Small / Medium
                        </Text>
                        <Text
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                            size="small"
                            weight="bold"
                        >
                            This is Body / Small / Bold
                        </Text>
                    </td>
                </tr>
            </tbody>
        </table>
    </>
)
