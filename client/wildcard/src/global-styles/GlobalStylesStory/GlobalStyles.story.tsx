// This story is NOT a complete replication of the Bootstrap documentation. This means it is not an exhaustive
// documentation of all the Bootstrap classes we have available in our app, please see refer to the Bootstrap
// documentation for that. Its primary purpose is to show what Bootstrap's components look like with our styling
// customizations.
import { action } from '@storybook/addon-actions'
import type { Decorator, Meta, StoryFn } from '@storybook/react'

import '@storybook/addon-designs'

import { TextArea, Button, Link, Select, Checkbox, Input, Text, Code, H1, H2, H3, H4, Form } from '../../components'
import { BrandedStory } from '../../stories'
import { highlightCodeSafe, registerHighlightContributions } from '../../utils'

import { ColorVariants } from './ColorVariants'
import { FormFieldVariants } from './FormFieldVariants'
import { preventDefault } from './utils'

registerHighlightContributions()

const decorator: Decorator = story => (
    <BrandedStory>{() => <div className="p-3 container">{story()}</div>}</BrandedStory>
)
const config: Meta = {
    title: 'branded/Global styles',
    decorators: [decorator],
    parameters: {},
}

export default config

export const CodeTypography: StoryFn = () => (
    <>
        <H1>Code</H1>

        <H2>Inline Code</H2>
        <Text>
            Example of <Code>inline code</Code> that can be achieved with the <Code>{'<code>'}</Code> element.
        </Text>

        <H2>Highlighted multi-line code</H2>
        <Text>Custom highlight.js themes are defined for both light and dark themes.</Text>

        <H3>TypeScript</H3>
        <pre>
            <Code
                dangerouslySetInnerHTML={{
                    __html: highlightCodeSafe(
                        ['const foo = 123', 'const bar = "Hello World!"', 'console.log(foo)'].join('\n'),
                        'typescript'
                    ),
                }}
            />
        </pre>

        <H3>JSON</H3>
        <pre>
            <Code
                dangerouslySetInnerHTML={{
                    __html: highlightCodeSafe(
                        ['{', '  "someString": "Hello World!",', '  "someNumber": 123', '}'].join('\n'),
                        'json'
                    ),
                }}
            />
        </pre>

        <H3>Diffs</H3>
        <pre>
            <Code
                dangerouslySetInnerHTML={{
                    __html: highlightCodeSafe(
                        [
                            ' const foo = 123',
                            '-const bar = "Hello, world!"',
                            '+const bar = "Hello, traveller!"',
                            ' console.log(foo)',
                        ].join('\n'),
                        'diff'
                    ),
                }}
            />
        </pre>

        <H2>Keyboard shortcuts</H2>
        <Text>
            Keyboard shortcuts should use <Code>{'<kbd>'}</Code>, not <Code>{'<code>'}</Code>. For example,{' '}
            <kbd>cmd</kbd>+<kbd>C</kbd> is used to copy text to the clipboard.
        </Text>
        <H3>Code snippets</H3>
    </>
)

export const Colors: StoryFn = () => (
    <>
        <H1>Colors</H1>

        <H2>Semantic colors</H2>
        <Text>These can be used to give semantic clues and always work both in light and dark theme.</Text>
        <ColorVariants />
    </>
)

Colors.parameters = {
    design: {
        name: 'Figma',
        type: 'figma',
        url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A7608',
    },
}

export const Layout: StoryFn = () => (
    <>
        <H1>Layout</H1>

        <H2>Spacing</H2>
        <Text>
            Use margin <Code>m-*</Code> and padding <Code>p-*</Code> utilities to align with the{' '}
            <Link
                to="https://builttoadapt.io/intro-to-the-8-point-grid-system-d2573cde8632"
                target="_blank"
                rel="noopener noreferrer"
            >
                8pt grid
            </Link>
            . When hand-writing CSS, use <Code>rem</Code> units in multiples of <Code>0.25</Code>.
        </Text>

        <H2>One-dimensional layout</H2>
        <Text>
            Use{' '}
            <Link
                to="https://css-tricks.com/snippets/css/a-guide-to-flexbox/"
                target="_blank"
                rel="noopener noreferrer"
            >
                Flexbox
            </Link>{' '}
            for one-dimensional layouts (single rows or columns, with optional wrapping). You can use{' '}
            <Link to="https://getbootstrap.com/docs/4.5/utilities/flex/" target="_blank" rel="noopener noreferrer">
                utility classes
            </Link>{' '}
            for simple flexbox layouts.
        </Text>

        <H3>Row layout</H3>
        <H4>Equally distributed</H4>
        <div
            className="d-flex p-1 border mb-2 overflow-hidden"
            style={{ resize: 'both', minWidth: '16rem', minHeight: '3rem' }}
        >
            <div className="p-1 m-1 flex-grow-1 d-flex justify-content-center align-items-center border">Column 1</div>
            <div className="p-1 m-1 flex-grow-1 d-flex justify-content-center align-items-center border">Column 2</div>
            <div className="p-1 m-1 flex-grow-1 d-flex justify-content-center align-items-center border">Column 3</div>
        </div>

        <H4>Middle column growing</H4>
        <div
            className="d-flex p-1 border mb-2 overflow-hidden"
            style={{ resize: 'both', minWidth: '16rem', minHeight: '3rem' }}
        >
            <div className="p-1 m-1 d-flex justify-content-center align-items-center border border">Column 1</div>
            <div className="p-1 m-1 d-flex justify-content-center align-items-center border flex-grow-1 border">
                Column 2
            </div>
            <div className="p-1 m-1 d-flex justify-content-center align-items-center border border">Column 3</div>
        </div>

        <H3>Column layout</H3>
        <div
            className="d-flex flex-column p-1 border mb-2 overflow-hidden"
            style={{ minHeight: '8rem', height: '12rem', minWidth: '6rem', width: '12rem', resize: 'both' }}
        >
            <div className="p-1 m-1 flex-grow-1 border d-flex align-items-center justify-content-center">Row 1</div>
            <div className="p-1 m-1 flex-grow-1 border d-flex align-items-center justify-content-center">Row 2</div>
            <div className="p-1 m-1 flex-grow-1 border d-flex align-items-center justify-content-center">Row 3</div>
        </div>

        <H2>Two-dimensional layout</H2>
        <Text>
            Use <Link to="https://learncssgrid.com/">CSS Grid</Link> for complex two-dimensional layouts.
        </Text>
        <div
            className="p-2 border overflow-hidden"
            style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(3, 1fr)',
                gridAutoRows: '1fr',
                gridGap: '0.5rem',
                resize: 'both',
                minWidth: '16rem',
                height: '16rem',
                minHeight: '6rem',
                marginBottom: '16rem',
            }}
        >
            <div className="border d-flex align-items-center justify-content-center">Cell 1</div>
            <div className="border d-flex align-items-center justify-content-center">Cell 2</div>
            <div className="border d-flex align-items-center justify-content-center">Cell 3</div>
            <div className="border d-flex align-items-center justify-content-center">Cell 4</div>
            <div className="border d-flex align-items-center justify-content-center">Cell 5</div>
            <div className="border d-flex align-items-center justify-content-center">Cell 6</div>
        </div>
    </>
)

export const InputGroups: StoryFn = () => (
    <>
        <H1>Input groups</H1>

        <Text>
            Easily extend form controls by adding text, buttons, or button groups on either side of textual inputs,
            custom selects, and custom file inputs.{' '}
            <Link to="https://getbootstrap.com/docs/4.5/components/input-group/">Bootstrap documentation</Link>
        </Text>

        <H2>Example</H2>
        <div>
            <div className="input-group" style={{ maxWidth: '24rem' }}>
                <Input type="search" placeholder="Search code..." aria-label="Search query" />
                <div className="input-group-append">
                    <Button type="submit" variant="primary">
                        Submit
                    </Button>
                </div>
            </div>
        </div>
    </>
)

InputGroups.storyName = 'Input groups'

export const Forms: StoryFn = () => (
    <>
        <H1>Forms</H1>
        <Text>
            Forms are validated using native HTML validation. Submit the below form with invalid input to try it out.{' '}
            <Link to="https://getbootstrap.com/docs/4.5/components/forms/" target="_blank" rel="noopener noreferrer">
                Bootstrap documentation
            </Link>
        </Text>
        <Form onSubmit={preventDefault}>
            <Input
                type="email"
                id="example-email-input"
                placeholder="me@example.com"
                label="Email address"
                message="We'll never share your email with anyone else."
                className="form-group"
                inputClassName="mb-0"
            />
            <Input
                type="password"
                id="example-input-password"
                className="form-group"
                inputClassName="mb-0"
                label="Password"
            />

            <Select isCustomStyle={true} aria-label="Example select" label="Example select">
                <option>Option A</option>
                <option>Option B</option>
                <option>Option C</option>
            </Select>

            <div className="form-group">
                <TextArea label="Example textarea" id="example-textarea" rows={3} />
            </div>

            <Checkbox label="Check me out" wrapperClassName="mb-3" id="exampleCheck1" />

            <Button type="submit" variant="primary">
                Submit
            </Button>
        </Form>

        <H2 className="mt-3">Disabled</H2>
        <Form>
            <fieldset disabled={true}>
                <Input
                    id="disabledTextInput"
                    placeholder="Disabled input"
                    className="form-group"
                    inputClassName="mb-0"
                    label="Disabled input"
                />

                <Select
                    isCustomStyle={true}
                    disabled={true}
                    label="Disabled select menu"
                    aria-label="Disabled select menu"
                >
                    <option>Disabled select</option>
                </Select>

                <div className="form-group">
                    <Checkbox label="Can't check this" id="disabledFieldsetCheck" disabled={true} />
                </div>
                <Button type="submit" variant="primary">
                    Submit
                </Button>
            </fieldset>
        </Form>

        <H2 className="mt-3">Readonly</H2>
        <Input value="I'm a readonly value" readOnly={true} />
        <H2 className="mt-3">Sizing</H2>
        <Text>Form fields can be made smaller</Text>
        <div className="d-flex">
            <fieldset>
                <div className="form-group">
                    <Input className="mb-1" placeholder="Small input" variant="small" />
                    <TextArea size="small" className="mb-1" placeholder="Small textarea" />
                    <Select
                        isCustomStyle={true}
                        selectSize="sm"
                        className="mb-0"
                        selectClassName="mb-1"
                        aria-label=""
                        id=""
                    >
                        <option>Small select</option>
                    </Select>
                </div>
            </fieldset>
        </div>
        <H2 className="mt-3">Field reference</H2>
        <FormFieldVariants />
    </>
)

Forms.parameters = {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/BkY8Ak997QauG0Iu2EqArv/Sourcegraph-Components?node-id=30%3A24',
    },
}

export const ListGroups: StoryFn = () => (
    <>
        <H1>List groups</H1>
        <Text>
            List groups are a flexible and powerful component for displaying a series of content. Modify and extend them
            to support just about any content within.
        </Text>
        <ul className="list-group mb-3">
            <li className="list-group-item">Cras justo odio</li>
            <li className="list-group-item">Dapibus ac facilisis in</li>
            <li className="list-group-item">Morbi leo risus</li>
            <li className="list-group-item">Porta ac consectetur ac</li>
            <li className="list-group-item">Vestibulum at eros</li>
        </ul>

        <H2>Interactive</H2>
        <div className="list-group">
            <button
                type="button"
                className="list-group-item list-group-item-action active"
                onClick={action('List group item clicked')}
            >
                Cras justo odio
            </button>
            <button
                type="button"
                className="list-group-item list-group-item-action"
                onClick={action('List group item clicked')}
            >
                Dapibus ac facilisis in
            </button>
            <button
                type="button"
                className="list-group-item list-group-item-action"
                onClick={action('List group item clicked')}
            >
                Morbi leo risus
            </button>
            <button
                type="button"
                className="list-group-item list-group-item-action"
                onClick={action('List group item clicked')}
            >
                Porta ac consectetur ac
            </button>
            <button
                type="button"
                className="list-group-item list-group-item-action disabled"
                tabIndex={-1}
                aria-disabled="true"
                onClick={action('List group item clicked')}
            >
                Disabled
            </button>
        </div>
    </>
)

ListGroups.storyName = 'List groups'

export const Meter: StoryFn = args => (
    <>
        <H1>Meter</H1>
        <Text>
            The HTML{' '}
            <Link
                to="https://developer.mozilla.org/en-US/docs/Web/HTML/Element/meter"
                target="_blank"
                rel="noopener noreferrer"
            >
                <Code>{'<meter>'}</Code>
            </Link>{' '}
            element represents either a scalar value within a known range or a fractional value.
        </Text>
        <H2>Examples</H2>
        <hr />
        <div className="pb-3">
            <H3>Optimum</H3>
            <meter min={0} max={1} optimum={1} value={1} />
        </div>
        <hr />
        <div className="pb-3">
            <H3>Sub optimum</H3>
            <meter min={0} max={1} high={0.8} low={0.2} optimum={1} value={0.6} />
        </div>
        <hr />
        <div className="pb-3">
            <H3>Sub sub optimum</H3>
            <meter min={0} max={1} high={0.8} low={0.2} optimum={1} value={0.1} />
        </div>
        <hr />
        <div className="pb-3">
            <H3>Customize with controls</H3>
            <meter {...args} />
        </div>
    </>
)

Meter.argTypes = {
    min: {
        type: 'number',
    },
    max: {
        type: 'number',
    },
    high: {
        type: 'number',
    },
    low: {
        type: 'number',
    },
    optimum: {
        type: 'number',
    },
    value: {
        type: 'number',
    },
}
Meter.args = {
    min: 0,
    max: 1,
    high: 0.8,
    low: 0.2,
    optimum: 1,
    value: 0.1,
}
