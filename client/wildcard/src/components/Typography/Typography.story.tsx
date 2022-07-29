import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Link } from '../Link'

import { TYPOGRAPHY_ALIGNMENTS, TYPOGRAPHY_MODES } from './constants'
import { Heading } from './Heading'

import { Code, Label, H1, H2, H3, H4, H5, H6, Text } from '.'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/Typography',

    decorators: [decorator],

    parameters: {
        component: Label,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
        design: {
            type: 'figma',
            name: 'Figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=5601%3A65477',
        },
    },
}

export default config

export const Simple: Story = (args = {}) => (
    <>
        <H2>Headings</H2>
        <table className="table">
            <tbody>
                <tr>
                    <td>
                        <Code>
                            {'<'}
                            H1
                            {'>'}
                        </Code>
                    </td>
                    <td>
                        <H1 mode={args.mode} alignment={args.alignment}>
                            This is H1
                        </H1>
                    </td>
                </tr>
                <tr>
                    <td>
                        <Code>
                            {'<'}
                            H2
                            {'>'}
                        </Code>
                    </td>
                    <td>
                        <H2 mode={args.mode} alignment={args.alignment}>
                            This is H2
                        </H2>
                    </td>
                </tr>
                <tr>
                    <td>
                        <Code>
                            {'<'}
                            H3
                            {'>'}
                        </Code>
                    </td>
                    <td>
                        <H3 mode={args.mode} alignment={args.alignment}>
                            This is H3
                        </H3>
                    </td>
                </tr>
                <tr>
                    <td>
                        <Code>
                            {'<'}
                            H4
                            {'>'}
                        </Code>
                    </td>
                    <td>
                        <H4 mode={args.mode} alignment={args.alignment}>
                            This is H4
                        </H4>
                    </td>
                </tr>
                <tr>
                    <td>
                        <Code>
                            {'<'}
                            H5
                            {'>'}
                        </Code>
                    </td>
                    <td>
                        <H5 mode={args.mode} alignment={args.alignment}>
                            This is H5
                        </H5>
                    </td>
                </tr>
                <tr>
                    <td>
                        <Code>
                            {'<'}
                            H6
                            {'>'}
                        </Code>
                    </td>
                    <td>
                        <H6 mode={args.mode} alignment={args.alignment}>
                            This is H6
                        </H6>
                    </td>
                </tr>
            </tbody>
        </table>

        <H2>Code</H2>
        <table className="table">
            <tbody>
                <tr>
                    <td>
                        <Code>
                            {'<'}
                            Code
                            {'>'}
                        </Code>
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

        <H2>Label</H2>
        <table className="table">
            <tbody>
                <tr>
                    <td>
                        <Code>
                            {'<'}
                            Label
                            {'>'}
                        </Code>
                    </td>
                    <td>
                        <div>
                            <Label mode={args.mode} alignment={args.alignment} size="base">
                                This is Label / Base
                            </Label>
                        </div>
                        <div>
                            <Label mode={args.mode} alignment={args.alignment} size="base" isUnderline={true}>
                                This is Label / Base - underline
                            </Label>
                        </div>
                        <div>
                            <Label mode={args.mode} alignment={args.alignment} size="small">
                                This is Label / Small
                            </Label>
                        </div>
                        <div>
                            <Label mode={args.mode} alignment={args.alignment} size="small" isUnderline={true}>
                                This is Label / Small - underline
                            </Label>
                        </div>
                        <div>
                            <Label mode={args.mode} alignment={args.alignment} isUppercase={true}>
                                This is Label / Uppercase / Base
                            </Label>
                        </div>
                        <div>
                            <Label mode={args.mode} alignment={args.alignment} size="small" isUppercase={true}>
                                This is Label / Uppercase / Small
                            </Label>
                        </div>
                    </td>
                </tr>
            </tbody>
        </table>

        <H2>Text</H2>
        <table className="table">
            <tbody>
                <tr>
                    <td>
                        <Code>
                            {'<'}
                            Text
                            {'>'}
                        </Code>
                    </td>
                    <td>
                        <Text mode={args.mode} alignment={args.alignment} size="base" weight="regular">
                            This is Body / Base / Regular
                        </Text>
                        <Text mode={args.mode} alignment={args.alignment} size="base" weight="medium">
                            This is Body / Base / Medium
                        </Text>
                        <Text mode={args.mode} alignment={args.alignment} size="base" weight="bold">
                            This is Body / Base / Bold
                        </Text>
                        <Text mode={args.mode} alignment={args.alignment} size="small" weight="regular">
                            This is Body / Small / Regular
                        </Text>
                        <Text mode={args.mode} alignment={args.alignment} size="small" weight="medium">
                            This is Body / Small / Medium
                        </Text>
                        <Text mode={args.mode} alignment={args.alignment} size="small" weight="bold">
                            This is Body / Small / Bold
                        </Text>
                    </td>
                </tr>
            </tbody>
        </table>
    </>
)
Simple.argTypes = {
    mode: {
        control: { type: 'select', options: TYPOGRAPHY_MODES },
        defaultValue: 'default',
    },
    alignment: {
        control: { type: 'select', options: TYPOGRAPHY_ALIGNMENTS },
        defaultValue: 'left',
    },
}

export const CrossingStyles: Story = () => (
    <>
        <H1>Crossing Header Styles</H1>
        <Text>
            Sometimes we need, for example, an <Code>{'<h2>'}</Code> with the styles of an <Code>{'<h3>'}</Code> to
            create a page structure that is both{' '}
            <Link
                target="_blank"
                rel="noopener noreferrer"
                to="https://docs.sourcegraph.com/dev/background-information/web/accessibility/detailed-checklist#headings"
            >
                accessible
            </Link>{' '}
            and visually-consistent with designs.
        </Text>
        <Text>
            It is possible both to <strong>"downscale"</strong> crossing header styles (e.g. <Code>{'<h2>'}</Code> with
            the styles of an <Code>{'<h3>'}</Code>) and <strong>"upscale"</strong> them (e.g. <Code>{'<h3>'}</Code> with
            the styles of an <Code>{'<h2>'}</Code>), but due to CSS priority rules, the two directions require using
            different APIs.
        </Text>

        <H2 className="mt-4 mb-3">Examples of downscaling</H2>

        <H2>I'm a normal H2.</H2>
        <H3 as={H2}>I'm an H2 with the style of an H3.</H3>
        <H4 as={H2}>I'm an H2 with the style of an H4.</H4>
        <H5 as={H2}>I'm an H2 with the style of an H5.</H5>

        <H2 className="mt-5 mb-3">Examples of upscaling</H2>
        <H4>I'm a normal H4.</H4>
        <Heading as="h4" styleAs="h3">
            I'm an H4 with the style of an H3.
        </Heading>
        <Heading as="h4" styleAs="h2">
            I'm an H4 with the style of an H2.
        </Heading>
        <Heading as="h4" styleAs="h1">
            I'm an H4 with the style of an H1.
        </Heading>
    </>
)
