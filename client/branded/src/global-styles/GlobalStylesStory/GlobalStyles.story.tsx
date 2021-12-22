// This story is NOT a complete replication of the Bootstrap documentation. This means it is not an exhaustive
// documentation of all the Bootstrap classes we have available in our app, please see refer to the Bootstrap
// documentation for that. Its primary purpose is to show what Bootstrap's componenents look like with our styling
// customizations.
import { action } from '@storybook/addon-actions'
import { number } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import classNames from 'classnames'
import React, { useState } from 'react'
import 'storybook-addon-designs'

import { registerHighlightContributions } from '@sourcegraph/shared/src/highlight/contributions'
import { highlightCodeSafe } from '@sourcegraph/shared/src/util/markdown'
import { Button } from '@sourcegraph/wildcard'

import { BrandedStory } from '../../components/BrandedStory'
import { CodeSnippet } from '../../components/CodeSnippet'
import { Form } from '../../components/Form'

import { AlertsStory } from './AlertsStory'
import { CardsStory } from './CardsStory'
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
}

export default config

export const Text: Story = () => (
    <>
        <h1>Typography</h1>

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

export const Code: Story = () => (
    <>
        <h1>Code</h1>

        <h2>Inline Code</h2>
        <p>
            Example of <code>inline code</code> that can be achieved with the <code>{'<code>'}</code> element.
        </p>

        <h2>Highlighted multi-line code</h2>
        <p>Custom highlight.js themes are defined for both light and dark themes.</p>

        <h3>TypeScript</h3>
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

        <h3>JSON</h3>
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

        <h3>Diffs</h3>
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

        <h2>Keyboard shortcuts</h2>
        <p>
            Keyboard shortcuts should use <code>{'<kbd>'}</code>, not <code>{'<code>'}</code>. For example,{' '}
            <kbd>cmd</kbd>+<kbd>C</kbd> is used to copy text to the clipboard.
        </p>
        <h3>Code snippets</h3>
        <p>
            Highlighted code pieces should go in a panel separating it from the surrounding content. Use{' '}
            <code>{'<CodeSnippet />'}</code> for these uses.
        </p>
        <CodeSnippet code="property: 1" language="yaml" />
    </>
)

export const Colors: Story = () => (
    <>
        <h1>Colors</h1>

        <h2>Semantic colors</h2>
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
        <h1>Layout</h1>

        <h2>Spacing</h2>
        <p>
            Use margin <code>m-*</code> and padding <code>p-*</code> utilities to align with the{' '}
            <a
                href="https://builttoadapt.io/intro-to-the-8-point-grid-system-d2573cde8632"
                target="_blank"
                rel="noopener noreferrer"
            >
                8pt grid
            </a>
            . When hand-writing CSS, use <code>rem</code> units in multiples of <code>0.25</code>.
        </p>

        <h2>One-dimensional layout</h2>
        <p>
            Use{' '}
            <a href="https://css-tricks.com/snippets/css/a-guide-to-flexbox/" target="_blank" rel="noopener noreferrer">
                Flexbox
            </a>{' '}
            for one-dimensional layouts (single rows or columns, with optional wrapping). You can use{' '}
            <a href="https://getbootstrap.com/docs/4.5/utilities/flex/" target="_blank" rel="noopener noreferrer">
                utility classes
            </a>{' '}
            for simple flexbox layouts.
        </p>

        <h3>Row layout</h3>
        <h4>Equally distributed</h4>
        <div
            className="d-flex p-1 border mb-2 overflow-hidden"
            style={{ resize: 'both', minWidth: '16rem', minHeight: '3rem' }}
        >
            <div className="p-1 m-1 flex-grow-1 d-flex justify-content-center align-items-center border">Column 1</div>
            <div className="p-1 m-1 flex-grow-1 d-flex justify-content-center align-items-center border">Column 2</div>
            <div className="p-1 m-1 flex-grow-1 d-flex justify-content-center align-items-center border">Column 3</div>
        </div>

        <h4>Middle column growing</h4>
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

        <h3>Column layout</h3>
        <div
            className="d-flex flex-column p-1 border mb-2 overflow-hidden"
            style={{ minHeight: '8rem', height: '12rem', minWidth: '6rem', width: '12rem', resize: 'both' }}
        >
            <div className="p-1 m-1 flex-grow-1 border d-flex align-items-center justify-content-center">Row 1</div>
            <div className="p-1 m-1 flex-grow-1 border d-flex align-items-center justify-content-center">Row 2</div>
            <div className="p-1 m-1 flex-grow-1 border d-flex align-items-center justify-content-center">Row 3</div>
        </div>

        <h2>Two-dimensional layout</h2>
        <p>
            Use <a href="https://learncssgrid.com/">CSS Grid</a> for complex two-dimensional layouts.
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

export const Alerts = AlertsStory

Alerts.parameters = {
    design: [
        {
            type: 'figma',
            name: 'Figma Light',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=1563%3A196',
        },
        {
            type: 'figma',
            name: 'Figma Dark',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=1563%3A525',
        },
    ],
}

export const ButtonGroups: Story = () => {
    const [active, setActive] = useState<'Left' | 'Middle' | 'Right'>('Left')
    return (
        <>
            <h1>Button groups</h1>
            <p>
                Group a series of buttons together on a single line with the button group.{' '}
                <a href="https://getbootstrap.com/docs/4.5/components/buttons/">Bootstrap documentation</a>
            </p>

            <h2>Example</h2>
            <div className="mb-2">
                <p>
                    Button groups have no styles on their own, they just group buttons together. This means they can be
                    used to group any other semantic or outline button variant.
                </p>
                <div className="mb-2">
                    <div className="btn-group" role="group" aria-label="Basic example">
                        <Button variant="secondary">Left</Button>
                        <Button variant="secondary">Middle</Button>
                        <Button variant="secondary">Right</Button>
                    </div>{' '}
                    Example with <code>btn-secondary</code>
                </div>
                <div className="mb-2">
                    <div className="btn-group" role="group" aria-label="Basic example">
                        <Button outline={true} variant="secondary">
                            Left
                        </Button>
                        <Button outline={true} variant="secondary">
                            Middle
                        </Button>
                        <Button outline={true} variant="secondary">
                            Right
                        </Button>
                    </div>{' '}
                    Example with <code>btn-outline-secondary</code>
                </div>
                <div className="mb-2">
                    <div className="btn-group" role="group" aria-label="Basic example">
                        <Button outline={true} variant="primary">
                            Left
                        </Button>
                        <Button outline={true} variant="primary">
                            Middle
                        </Button>
                        <Button outline={true} variant="primary">
                            Right
                        </Button>
                    </div>{' '}
                    Example with <code>btn-outline-primary</code>
                </div>
            </div>

            <h2 className="mt-3">Sizing</h2>
            <p>
                Just like buttons, button groups have <code>sm</code> and <code>lg</code> size variants.
            </p>
            <div className="mb-2">
                {['btn-group-lg', '', 'btn-group-sm'].map(size => (
                    <div key={size} className="mb-2">
                        <div className={classNames('btn-group', size)} role="group" aria-label="Sizing example">
                            <Button outline={true} variant="primary">
                                Left
                            </Button>
                            <Button outline={true} variant="primary">
                                Middle
                            </Button>
                            <Button outline={true} variant="primary">
                                Right
                            </Button>
                        </div>
                    </div>
                ))}
            </div>

            <h2 className="mt-3">Active state</h2>
            <p>
                The <code>active</code> class can be used to craft toggles out of button groups.
            </p>
            <div className="mb-2">
                <div className="btn-group" role="group" aria-label="Basic example">
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
                </div>{' '}
                Example with <code>btn-outline-secondary</code>
            </div>
            <div className="mb-2">
                <div className="btn-group" role="group" aria-label="Basic example">
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
                </div>{' '}
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
        <h1>Input groups</h1>

        <p>
            Easily extend form controls by adding text, buttons, or button groups on either side of textual inputs,
            custom selects, and custom file inputs.{' '}
            <a href="https://getbootstrap.com/docs/4.5/components/input-group/">Bootstrap documentation</a>
        </p>

        <h2>Example</h2>
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
        <h1>Forms</h1>
        <p>
            Forms are validated using native HTML validation. Submit the below form with invalid input to try it out.{' '}
            <a href="https://getbootstrap.com/docs/4.5/components/forms/" target="_blank" rel="noopener noreferrer">
                Bootstrap documentation
            </a>
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
            <div className="form-group">
                <label htmlFor="example-example-select">Example select</label>
                <select id="example-select" className="custom-select">
                    <option>Option A</option>
                    <option>Option B</option>
                    <option>Option C</option>
                </select>
            </div>
            <div className="form-group">
                <label htmlFor="example-textarea">Example textarea</label>
                <textarea className="form-control" id="example-textarea" rows={3} />
            </div>
            <div className="form-group form-check">
                <input type="checkbox" className="form-check-input" id="exampleCheck1" />
                <label className="form-check-label" htmlFor="exampleCheck1">
                    Check me out
                </label>
            </div>
            <Button type="submit" variant="primary">
                Submit
            </Button>
        </Form>

        <h2 className="mt-3">Disabled</h2>
        <Form>
            <fieldset disabled={true}>
                <div className="form-group">
                    <label htmlFor="disabledTextInput">Disabled input</label>
                    <input type="text" id="disabledTextInput" className="form-control" placeholder="Disabled input" />
                </div>
                <div className="form-group">
                    <label htmlFor="disabledSelect">Disabled select menu</label>
                    <select id="disabledSelect" className="custom-select">
                        <option>Disabled select</option>
                    </select>
                </div>
                <div className="form-group">
                    <div className="form-check">
                        <input
                            className="form-check-input"
                            type="checkbox"
                            id="disabledFieldsetCheck"
                            disabled={true}
                        />
                        <label className="form-check-label" htmlFor="disabledFieldsetCheck">
                            Can't check this
                        </label>
                    </div>
                </div>
                <Button type="submit" variant="primary">
                    Submit
                </Button>
            </fieldset>
        </Form>

        <h2 className="mt-3">Readonly</h2>
        <input className="form-control" type="text" value="I'm a readonly value" readOnly={true} />

        <h2 className="mt-3">Sizing</h2>
        <p>Form fields can be made smaller</p>
        <div className="d-flex">
            <fieldset>
                <div className="form-group">
                    <input className="form-control form-control-sm mb-1" type="text" placeholder="Small input" />
                    <textarea className="form-control form-control-sm mb-1" placeholder="Small textarea" />
                    <select className="custom-select custom-select-sm mb-1">
                        <option>Small select</option>
                    </select>
                </div>
            </fieldset>
        </div>
        <h2 className="mt-3">Field reference</h2>
        <FormFieldVariants />
    </>
)

Forms.parameters = {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/BkY8Ak997QauG0Iu2EqArv/Sourcegraph-Components?node-id=30%3A24',
    },
}

export const Cards = CardsStory

Cards.parameters = {
    design: {
        name: 'Figma',
        type: 'figma',
        url:
            'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=1172%3A285',
    },
}

export const ListGroups: Story = () => (
    <>
        <h1>List groups</h1>
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

        <h2>Interactive</h2>
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
            <h1>Meter</h1>
            <p>
                The HTML{' '}
                <a
                    href="https://developer.mozilla.org/en-US/docs/Web/HTML/Element/meter"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    <code>{'<meter>'}</code>
                </a>{' '}
                element represents either a scalar value within a known range or a fractional value.
            </p>
            <h2>Examples</h2>
            <hr />
            <div className="pb-3">
                <h3>Optimum</h3>
                <meter min={0} max={1} optimum={1} value={1} />
            </div>
            <hr />
            <div className="pb-3">
                <h3>Sub optimum</h3>
                <meter min={0} max={1} high={0.8} low={0.2} optimum={1} value={0.6} />
            </div>
            <hr />
            <div className="pb-3">
                <h3>Sub sub optimum</h3>
                <meter min={0} max={1} high={0.8} low={0.2} optimum={1} value={0.1} />
            </div>
            <hr />
            <div className="pb-3">
                <h3>Customize with knobs</h3>
                <meter min={min} max={max} high={high} low={low} optimum={optimum} value={value} />
            </div>
        </>
    )
}
