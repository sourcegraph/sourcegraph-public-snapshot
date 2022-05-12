import { select } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { TYPOGRAPHY_ALIGNMENTS, TYPOGRAPHY_MODES } from './constants'
import * as Typography from './Typography'

const decorator: DecoratorFn = story => <BrandedStory styles={webStyles}>{() => <div>{story()}</div>}</BrandedStory>

const config: Meta = {
    title: 'wildcard/Typography/All',

    decorators: [decorator],

    parameters: {
        component: Typography.Label,
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

export const Simple: Story = () => (
    <>
        <Typography.H2>Headings</Typography.H2>
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
                        <Typography.H1
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                        >
                            This is H1
                        </Typography.H1>
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
                        <Typography.H2
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                        >
                            This is H2
                        </Typography.H2>
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
                        <Typography.H3
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                        >
                            This is H3
                        </Typography.H3>
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
                        <Typography.H4
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                        >
                            This is H4
                        </Typography.H4>
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
                        <Typography.H5
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                        >
                            This is H5
                        </Typography.H5>
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
                        <Typography.H6
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                        >
                            This is H6
                        </Typography.H6>
                    </td>
                </tr>
            </tbody>
        </table>

        <Typography.H2>Code</Typography.H2>
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
                            <Typography.Code size="base" weight="regular">
                                This is Code / Base / Regular
                            </Typography.Code>
                        </div>
                        <div>
                            <Typography.Code size="base" weight="bold">
                                This is Code / Base / Bold
                            </Typography.Code>
                        </div>
                        <div>
                            <Typography.Code size="small" weight="regular">
                                This is Code / Small / Regular
                            </Typography.Code>
                        </div>
                        <div>
                            <Typography.Code size="small" weight="bold">
                                This is Code / Small / Bold
                            </Typography.Code>
                        </div>
                    </td>
                </tr>
            </tbody>
        </table>

        <Typography.H2>Label</Typography.H2>
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
                            <Typography.Label
                                mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                                alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                                size="base"
                            >
                                This is Label / Base
                            </Typography.Label>
                        </div>
                        <div>
                            <Typography.Label
                                mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                                alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                                size="base"
                                isUnderline={true}
                            >
                                This is Label / Base - underline
                            </Typography.Label>
                        </div>
                        <div>
                            <Typography.Label
                                mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                                alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                                size="small"
                            >
                                This is Label / Small
                            </Typography.Label>
                        </div>
                        <div>
                            <Typography.Label
                                mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                                alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                                size="small"
                                isUnderline={true}
                            >
                                This is Label / Small - underline
                            </Typography.Label>
                        </div>
                        <div>
                            <Typography.Label
                                mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                                alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                                isUppercase={true}
                            >
                                This is Label / Uppercase / Base
                            </Typography.Label>
                        </div>
                        <div>
                            <Typography.Label
                                mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                                alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                                size="small"
                                isUppercase={true}
                            >
                                This is Label / Uppercase / Small
                            </Typography.Label>
                        </div>
                    </td>
                </tr>
            </tbody>
        </table>

        <Typography.H2>Text</Typography.H2>
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
                        <Typography.Text
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                            size="base"
                            weight="regular"
                        >
                            This is Body / Base / Regular
                        </Typography.Text>
                        <Typography.Text
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                            size="base"
                            weight="medium"
                        >
                            This is Body / Base / Medium
                        </Typography.Text>
                        <Typography.Text
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                            size="base"
                            weight="bold"
                        >
                            This is Body / Base / Bold
                        </Typography.Text>
                        <Typography.Text
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                            size="small"
                            weight="regular"
                        >
                            This is Body / Small / Regular
                        </Typography.Text>
                        <Typography.Text
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                            size="small"
                            weight="medium"
                        >
                            This is Body / Small / Medium
                        </Typography.Text>
                        <Typography.Text
                            mode={select('mode', TYPOGRAPHY_MODES, undefined)}
                            alignment={select('alignment', TYPOGRAPHY_ALIGNMENTS, undefined)}
                            size="small"
                            weight="bold"
                        >
                            This is Body / Small / Bold
                        </Typography.Text>
                    </td>
                </tr>
            </tbody>
        </table>
    </>
)
