// This story is NOT a complete replication of the Bootstrap documentation. This means it is not an exhaustive
// documentation of all the Bootstrap classes we have available in our app, please see refer to the Bootstrap
// documentation for that. Its primary purpose is to show what Bootstrap's componenents look like with our styling
// customizations.
import { useState } from 'react'

import { action } from '@storybook/addon-actions'
import { number } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import classNames from 'classnames'
import 'storybook-addon-designs'

import { highlightCodeSafe, registerHighlightContributions } from '@sourcegraph/common'
import { TextArea, Button, ButtonGroup, Link, Select, BUTTON_SIZES, Checkbox, Typography } from '@sourcegraph/wildcard'

import { BrandedStory } from '../../components/BrandedStory'
import { CodeSnippet } from '../../components/CodeSnippet'
import { Form } from '../../components/Form'

import { ColorVariants } from './ColorVariants'
import { FormFieldVariants } from './FormFieldVariants'
import { TextStory } from './TextStory'
import { preventDefault } from './utils'

registerHighlightContributions()

const decorator: DecoratorFn = story => (
    <BrandedStory>{() => <div className="p-3 container">{story()}</div>}</BrandedStory>
)
const config: Meta = {
    title: 'branded/Global styles',
    decorators: [decorator],
    parameters: {
        chromatic: {
            enableDarkMode: true,
        },
    },
}

export default config

export const Text: Story = () => (
    <>
        <Typography.H1>Typography</Typography.H1>

        <TextStory />
    </>
)

Text.parameters = {
    design: {
        name: 'Figma',
        type: 'figma',
        url:
            'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=998%3A1515',
    },
}

type ButtonSizesType = typeof BUTTON_SIZES[number] | undefined

export const Code: Story = () => (
    <>
        <Typography.H1>Code</Typography.H1>

        <Typography.H2>Inline Code</Typography.H2>
        <p>
            Example of <code>inline code</code> that can be achieved with the <code>{'<code>'}</code> element.
        </p>

        <Typography.H2>Highlighted multi-line code</Typography.H2>
        <p>Custom highlight.js themes are defined for both light and dark themes.</p>

        <Typography.H3>TypeScript</Typography.H3>
        <pre>
            <code
                dangerouslySetInnerHTML={{
                    __html: highlightCodeSafe(
                        ['const foo = 123', 'const bar = "Hello World!"', 'console.log(foo)'].join('\n'),
                        'typescript'
                    ),
                }}
            />
        </pre>

        <Typography.H3>JSON</Typography.H3>
        <pre>
            <code
                dangerouslySetInnerHTML={{
                    __html: highlightCodeSafe(
                        ['{', '  "someString": "Hello World!",', '  "someNumber": 123', '}'].join('\n'),
                        'json'
                    ),
                }}
            />
        </pre>

        <Typography.H3>Diffs</Typography.H3>
        <pre>
            <code
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

        <Typography.H2>Keyboard shortcuts</Typography.H2>
        <p>
            Keyboard shortcuts should use <code>{'<kbd>'}</code>, not <code>{'<code>'}</code>. For example,{' '}
            <kbd>cmd</kbd>+<kbd>C</kbd> is used to copy text to the clipboard.
        </p>
        <Typography.H3>Code snippets</Typography.H3>
        <p>
            Highlighted code pieces should go in a panel separating it from the surrounding content. Use{' '}
            <code>{'<CodeSnippet />'}</code> for these uses.
        </p>
        <CodeSnippet code="property: 1" language="yaml" />
    </>
)

export const Colors: Story = () => (
    <>
        <Typography.H1>Colors</Typography.H1>

        <Typography.H2>Semantic colors</Typography.H2>
        <p>These can be used to give semantic clues and always work both in light and dark theme.</p>
        <ColorVariants />
    </>
)

Colors.parameters = {
    design: {
        name: 'Figma',
        type: 'figma',
        url:
            'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A7608',
    },
}

export const Layout: Story = () => (
    <>
        <Typography.H1>Layout</Typography.H1>

        <Typography.H2>Spacing</Typography.H2>
        <p>
            Use margin <code>m-*</code> and padding <code>p-*</code> utilities to align with the{' '}
            <Link
                to="https://builttoadapt.io/intro-to-the-8-point-grid-system-d2573cde8632"
                target="_blank"
                rel="noopener noreferrer"
            >
                8pt grid
            </Link>
            . When hand-writing CSS, use <code>rem</code> units in multiples of <code>0.25</code>.
        </p>

        <Typography.H2>One-dimensional layout</Typography.H2>
        <p>
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
        </p>

        <Typography.H3>Row layout</Typography.H3>
        <Typography.H4>Equally distributed</Typography.H4>
        <div
            className="d-flex p-1 border mb-2 overflow-hidden"
            style={{ resize: 'both', minWidth: '16rem', minHeight: '3rem' }}
        >
            <div className="p-1 m-1 flex-grow-1 d-flex justify-content-center align-items-center border">Column 1</div>
            <div className="p-1 m-1 flex-grow-1 d-flex justify-content-center align-items-center border">Column 2</div>
            <div className="p-1 m-1 flex-grow-1 d-flex justify-content-center align-items-center border">Column 3</div>
        </div>

        <Typography.H4>Middle column growing</Typography.H4>
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

        <Typography.H3>Column layout</Typography.H3>
        <div
            className="d-flex flex-column p-1 border mb-2 overflow-hidden"
            style={{ minHeight: '8rem', height: '12rem', minWidth: '6rem', width: '12rem', resize: 'both' }}
        >
            <div className="p-1 m-1 flex-grow-1 border d-flex align-items-center justify-content-center">Row 1</div>
            <div className="p-1 m-1 flex-grow-1 border d-flex align-items-center justify-content-center">Row 2</div>
            <div className="p-1 m-1 flex-grow-1 border d-flex align-items-center justify-content-center">Row 3</div>
        </div>

        <Typography.H2>Two-dimensional layout</Typography.H2>
        <p>
            Use <Link to="https://learncssgrid.com/">CSS Grid</Link> for complex two-dimensional layouts.
        </p>
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

export const ButtonGroups: Story = () => {
    const [active, setActive] = useState<'Left' | 'Middle' | 'Right'>('Left')
    const buttonSizes: ButtonSizesType[] = ['lg', undefined, 'sm']
    return (
        <>
            <Typography.H1>Button groups</Typography.H1>
            <p>
                Group a series of buttons together on a single line with the button group.{' '}
                <Link to="https://getbootstrap.com/docs/4.5/components/buttons/">Bootstrap documentation</Link>
            </p>

            <Typography.H2>Example</Typography.H2>
            <div className="mb-2">
                <p>
                    Button groups have no styles on their own, they just group buttons together. This means they can be
                    used to group any other semantic or outline button variant.
                </p>
                <div className="mb-2">
                    <ButtonGroup aria-label="Basic example">
                        <Button variant="secondary">Left</Button>
                        <Button variant="secondary">Middle</Button>
                        <Button variant="secondary">Right</Button>
                    </ButtonGroup>{' '}
                    Example with <code>btn-secondary</code>
                </div>
                <div className="mb-2">
                    <ButtonGroup aria-label="Basic example">
                        <Button outline={true} variant="secondary">
                            Left
                        </Button>
                        <Button outline={true} variant="secondary">
                            Middle
                        </Button>
                        <Button outline={true} variant="secondary">
                            Right
                        </Button>
                    </ButtonGroup>{' '}
                    Example with <code>btn-outline-secondary</code>
                </div>
                <div className="mb-2">
                    <ButtonGroup aria-label="Basic example">
                        <Button outline={true} variant="primary">
                            Left
                        </Button>
                        <Button outline={true} variant="primary">
                            Middle
                        </Button>
                        <Button outline={true} variant="primary">
                            Right
                        </Button>
                    </ButtonGroup>{' '}
                    Example with <code>btn-outline-primary</code>
                </div>
            </div>

            <Typography.H2 className="mt-3">Sizing</Typography.H2>
            <p>
                Just like buttons, button groups have <code>sm</code> and <code>lg</code> size variants.
            </p>
            <div className="mb-2">
                {buttonSizes.map(size => (
                    <div key={size} className="mb-2">
                        <ButtonGroup aria-label="Sizing example">
                            <Button size={size} outline={true} variant="primary">
                                Left
                            </Button>
                            <Button size={size} outline={true} variant="primary">
                                Middle
                            </Button>
                            <Button size={size} outline={true} variant="primary">
                                Right
                            </Button>
                        </ButtonGroup>
                    </div>
                ))}
            </div>

            <Typography.H2 className="mt-3">Active state</Typography.H2>
            <p>
                The <code>active</code> class can be used to craft toggles out of button groups.
            </p>
            <div className="mb-2">
                <ButtonGroup aria-label="Basic example">
                    {(['Left', 'Middle', 'Right'] as const).map(option => (
                        <Button
                            key={option}
                            className={classNames(option === active && 'active')}
                            onClick={() => setActive(option)}
                            aria-pressed={option === active}
                            outline={true}
                            variant="secondary"
                        >
                            {option}
                        </Button>
                    ))}
                </ButtonGroup>{' '}
                Example with <code>btn-outline-secondary</code>
            </div>
            <div className="mb-2">
                <ButtonGroup aria-label="Basic example">
                    {(['Left', 'Middle', 'Right'] as const).map(option => (
                        <Button
                            key={option}
                            className={classNames(option === active && 'active')}
                            onClick={() => setActive(option)}
                            aria-pressed={option === active}
                            outline={true}
                            variant="primary"
                        >
                            {option}
                        </Button>
                    ))}
                </ButtonGroup>{' '}
                Example with <code>btn-outline-primary</code>
            </div>
        </>
    )
}

ButtonGroups.storyName = 'Button groups'

ButtonGroups.parameters = {
    design: {
        type: 'figma',
        name: 'Figma',
        url:
            'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A2514',
    },
}

export const InputGroups: Story = () => (
    <>
        <Typography.H1>Input groups</Typography.H1>

        <p>
            Easily extend form controls by adding text, buttons, or button groups on either side of textual inputs,
            custom selects, and custom file inputs.{' '}
            <Link to="https://getbootstrap.com/docs/4.5/components/input-group/">Bootstrap documentation</Link>
        </p>

        <Typography.H2>Example</Typography.H2>
        <div>
            <div className="input-group" style={{ maxWidth: '24rem' }}>
                <input type="search" className="form-control" placeholder="Search code..." aria-label="Search query" />
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

export const Forms: Story = () => (
    <>
        <Typography.H1>Forms</Typography.H1>
        <p>
            Forms are validated using native HTML validation. Submit the below form with invalid input to try it out.{' '}
            <Link to="https://getbootstrap.com/docs/4.5/components/forms/" target="_blank" rel="noopener noreferrer">
                Bootstrap documentation
            </Link>
        </p>
        <Form onSubmit={preventDefault}>
            <div className="form-group">
                <label htmlFor="example-email-input">Email address</label>
                <input
                    type="email"
                    className="form-control"
                    id="example-email-input"
                    aria-describedby="email-help"
                    placeholder="me@example.com"
                />
                <small id="email-help" className="form-text text-muted">
                    We'll never share your email with anyone else.
                </small>
            </div>
            <div className="form-group">
                <label htmlFor="example-input-password">Password</label>
                <input type="password" className="form-control" id="example-input-password" />
            </div>

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

        <Typography.H2 className="mt-3">Disabled</Typography.H2>
        <Form>
            <fieldset disabled={true}>
                <div className="form-group">
                    <label htmlFor="disabledTextInput">Disabled input</label>
                    <input type="text" id="disabledTextInput" className="form-control" placeholder="Disabled input" />
                </div>

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

        <Typography.H2 className="mt-3">Readonly</Typography.H2>
        <input className="form-control" type="text" value="I'm a readonly value" readOnly={true} />

        <Typography.H2 className="mt-3">Sizing</Typography.H2>
        <p>Form fields can be made smaller</p>
        <div className="d-flex">
            <fieldset>
                <div className="form-group">
                    <input className="form-control form-control-sm mb-1" type="text" placeholder="Small input" />
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
        <Typography.H2 className="mt-3">Field reference</Typography.H2>
        <FormFieldVariants />
    </>
)

Forms.parameters = {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/BkY8Ak997QauG0Iu2EqArv/Sourcegraph-Components?node-id=30%3A24',
    },
}

export const ListGroups: Story = () => (
    <>
        <Typography.H1>List groups</Typography.H1>
        <p>
            List groups are a flexible and powerful component for displaying a series of content. Modify and extend them
            to support just about any content within.
        </p>
        <ul className="list-group mb-3">
            <li className="list-group-item">Cras justo odio</li>
            <li className="list-group-item">Dapibus ac facilisis in</li>
            <li className="list-group-item">Morbi leo risus</li>
            <li className="list-group-item">Porta ac consectetur ac</li>
            <li className="list-group-item">Vestibulum at eros</li>
        </ul>

        <Typography.H2>Interactive</Typography.H2>
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

export const Meter: Story = () => {
    const min = number('min', 0)
    const max = number('max', 1)
    const high = number('high', 0.8)
    const low = number('low', 0.2)
    const optimum = number('optimum', 1)
    const value = number('value', 0.1)

    return (
        <>
            <Typography.H1>Meter</Typography.H1>
            <p>
                The HTML{' '}
                <Link
                    to="https://developer.mozilla.org/en-US/docs/Web/HTML/Element/meter"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    <code>{'<meter>'}</code>
                </Link>{' '}
                element represents either a scalar value within a known range or a fractional value.
            </p>
            <Typography.H2>Examples</Typography.H2>
            <hr />
            <div className="pb-3">
                <Typography.H3>Optimum</Typography.H3>
                <meter min={0} max={1} optimum={1} value={1} />
            </div>
            <hr />
            <div className="pb-3">
                <Typography.H3>Sub optimum</Typography.H3>
                <meter min={0} max={1} high={0.8} low={0.2} optimum={1} value={0.6} />
            </div>
            <hr />
            <div className="pb-3">
                <Typography.H3>Sub sub optimum</Typography.H3>
                <meter min={0} max={1} high={0.8} low={0.2} optimum={1} value={0.1} />
            </div>
            <hr />
            <div className="pb-3">
                <Typography.H3>Customize with knobs</Typography.H3>
                <meter min={min} max={max} high={high} low={low} optimum={optimum} value={value} />
            </div>
        </>
    )
}
