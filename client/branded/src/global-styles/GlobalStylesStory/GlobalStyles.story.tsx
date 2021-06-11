// This story is NOT a complete replication of the Bootstrap documentation. This means it is not an exhaustive
// documentation of all the Bootstrap classes we have available in our app, please see refer to the Bootstrap
// documentation for that. Its primary purpose is to show what Bootstrap's componenents look like with our styling
// customizations.

import { Menu, MenuButton, MenuList, MenuLink } from '@reach/menu-button'
import { action } from '@storybook/addon-actions'
import { number } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import classNames from 'classnames'
import SearchIcon from 'mdi-react/SearchIcon'
import React, { useState } from 'react'
import 'storybook-addon-designs'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { highlightCodeSafe } from '@sourcegraph/shared/src/util/markdown'

import { BrandedStory } from '../../components/BrandedStory'
import { CodeSnippet } from '../../components/CodeSnippet'
import { Form } from '../../components/Form'

import { AlertsStory } from './AlertsStory'
import { BadgeVariants } from './BadgeVariants/BadgeVariants'
import { ButtonVariants } from './ButtonVariants'
import { CardsStory } from './CardsStory'
import { ColorVariants } from './ColorVariants'
import { SEMANTIC_COLORS } from './constants'
import { FormFieldVariants } from './FormFieldVariants'
import { TextStory } from './TextStory'
import { preventDefault } from './utils'

const { add } = storiesOf('branded/Global styles', module).addDecorator(story => (
    <BrandedStory>{() => <div className="p-3 container">{story()}</div>}</BrandedStory>
))

add(
    'Text',
    () => (
        <>
            <h1>Typography</h1>

            <TextStory />
        </>
    ),
    {
        design: {
            name: 'Figma',
            type: 'figma',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=998%3A1515',
        },
    }
)

add(
    'Web content',
    () => (
        <div className="web-content">
            <h1>Web content</h1>
            <p>
                The <code>web-content</code> class changes the text styles of all descendants for content that more
                closely matches rich web sites as opposed to our high-information-density, application-like code content
                areas.
            </p>

            <TextStory />
        </div>
    ),
    {
        design: {
            name: 'Figma',
            type: 'figma',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=998%3A1515',
        },
    }
)

add('Code', () => (
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
))

add(
    'Colors',
    () => (
        <>
            <h1>Colors</h1>

            <h2>Semantic colors</h2>
            <p>These can be used to give semantic clues and always work both in light and dark theme.</p>
            <ColorVariants />
        </>
    ),
    {
        design: {
            name: 'Figma',
            type: 'figma',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A7608',
        },
    }
)

add('Layout', () => (
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
))

add('Alerts', AlertsStory, {
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
})

add(
    'Badges',
    () => (
        <>
            <h1>Badges</h1>
            <p>
                <a href="https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+count:1000+badge+badge-&patternType=literal">
                    Usages
                </a>{' '}
                | <a href="https://getbootstrap.com/docs/4.5/components/badge/">Bootstrap Documentation</a>{' '}
            </p>
            <p>Badges are used for labelling and displaying small counts.</p>

            <h2>Scaling</h2>
            <p>
                Badges scale to match the size of the immediate parent element by using relative font sizing and{' '}
                <code>em</code> units for padding.
            </p>
            <p>
                Use a superscript <code>{'<sup></sup>'}</code> to position the badge top-right of a word in{' '}
                <code>h1</code> headings. Do not use a superscript for smaller text, because the font size would become
                too small.
            </p>
            <table className="table">
                <tbody>
                    <tr>
                        <td>
                            <code>{'<h1>'}</code> + <code>{'<sup>'}</code>
                        </td>
                        <td>
                            <h1>
                                Lorem{' '}
                                <sup>
                                    <span className="badge badge-secondary">ipsum</span>
                                </sup>
                            </h1>
                            <small>Use a superscript to align the badge top-right of the heading text.</small>
                        </td>
                    </tr>
                    {(['h2', 'h3', 'h4', 'h5', 'h6'] as const).map(Heading => (
                        <tr key={Heading}>
                            <td>
                                <code>{`<${Heading}>`}</code>
                            </td>
                            <td>
                                <Heading>
                                    Lorem <span className="badge badge-secondary">ipsum</span>
                                </Heading>
                            </td>
                        </tr>
                    ))}
                    <tr>
                        <td>Regular text</td>
                        <td>
                            Lorem <span className="badge badge-secondary">ipsum</span>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <code>{'<small>'}</code>
                        </td>
                        <td>
                            <small>
                                Lorem <span className="badge badge-secondary">ipsum</span>
                            </small>
                            <p>
                                <small className="text-danger">
                                    Discouraged because the text becomes too small to read.
                                </small>
                            </p>
                        </td>
                    </tr>
                </tbody>
            </table>

            <h2>Reference</h2>
            <BadgeVariants variants={[...SEMANTIC_COLORS, 'outline-secondary']} />
            <h3>Size</h3>
            <p>We can also make our badges smaller.</p>
            <BadgeVariants small={true} variants={['primary', 'secondary']} />
            <h2>Pill badges</h2>
            <p>Pill badges are commonly used to display counts.</p>
            <div className="mb-4">
                Matches <span className="badge badge-pill badge-secondary">321+</span>
            </div>
            <div>
                <ul className="nav nav-tabs mb-2">
                    <li className="nav-item">
                        <a className="nav-link active" href="/" onClick={preventDefault}>
                            <span>
                                <span className="text-content" data-test-tab="Comments">
                                    Comments
                                </span>{' '}
                                <span className="badge badge-pill badge-secondary">14</span>
                            </span>
                        </a>
                    </li>
                    <li className="nav-item">
                        <a className="nav-link" href="/" onClick={preventDefault}>
                            <span>
                                <span className="text-content" data-test-tab="Changed files">
                                    Changed files
                                </span>{' '}
                                <span className="badge badge-pill badge-secondary">6</span>
                            </span>
                        </a>
                    </li>
                </ul>

                <span>No content here!</span>
            </div>

            <h2>Links</h2>

            <p>
                <LinkOrSpan className="badge badge-secondary" to="http://google.com">
                    Tooltip
                </LinkOrSpan>
            </p>
        </>
    ),
    {
        design: [
            {
                type: 'figma',
                name: 'Figma - Light',
                url:
                    'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A6149',
            },
            {
                type: 'figma',
                name: 'Figma - Dark',
                url:
                    'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A6448',
            },
        ],
    }
)

add(
    'Buttons',
    () => (
        <>
            <h1>Buttons</h1>
            <p>
                Use Bootstrap’s custom button styles for actions in forms, dialogs, and more with support for multiple
                sizes, states, and more.{' '}
                <a href="https://getbootstrap.com/docs/4.5/components/buttons/">Bootstrap documentation</a>
            </p>
            <h2>Semantic variants</h2>
            <ButtonVariants variants={SEMANTIC_COLORS} />
            <h2>Outline variants</h2>
            <ButtonVariants variants={['primary', 'secondary', 'danger']} variantType="btn-outline" />
            <h2>Icons</h2>
            <p>We can use icons with our buttons</p>
            <ButtonVariants variants={['danger']} icon={SearchIcon} />
            <ButtonVariants variants={['danger']} variantType="btn-outline" icon={SearchIcon} />
            <h2>Size</h2>
            <p>We can make our buttons smaller</p>
            <ButtonVariants variants={['primary']} variantType="btn-outline" small={true} />
            <h2>Links</h2>
            <p>Links can be made to look like buttons:</p>
            <a href="https://example.com" className="btn btn-secondary mb-3" target="_blank" rel="noopener noreferrer">
                I am a link
            </a>
            <p>Buttons can be made to look like links:</p>
            <button type="button" className="btn btn-link mr-3">
                Link button
            </button>
            <button type="button" className="btn btn-link mr-3 focus">
                Focused
            </button>
            <button type="button" className="btn btn-link mr-3" disabled={true}>
                Disabled
            </button>
        </>
    ),
    {
        design: {
            type: 'figma',
            name: 'Figma',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A2513',
        },
    }
)

add(
    'Button groups',
    () => {
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
                        Button groups have no styles on their own, they just group buttons together. This means they can
                        be used to group any other semantic or outline button variant.
                    </p>
                    <div className="mb-2">
                        <div className="btn-group" role="group" aria-label="Basic example">
                            <button type="button" className="btn btn-secondary">
                                Left
                            </button>
                            <button type="button" className="btn btn-secondary">
                                Middle
                            </button>
                            <button type="button" className="btn btn-secondary">
                                Right
                            </button>
                        </div>{' '}
                        Example with <code>btn-secondary</code>
                    </div>
                    <div className="mb-2">
                        <div className="btn-group" role="group" aria-label="Basic example">
                            <button type="button" className="btn btn-outline-secondary">
                                Left
                            </button>
                            <button type="button" className="btn btn-outline-secondary">
                                Middle
                            </button>
                            <button type="button" className="btn btn-outline-secondary">
                                Right
                            </button>
                        </div>{' '}
                        Example with <code>btn-outline-secondary</code>
                    </div>
                    <div className="mb-2">
                        <div className="btn-group" role="group" aria-label="Basic example">
                            <button type="button" className="btn btn-outline-primary">
                                Left
                            </button>
                            <button type="button" className="btn btn-outline-primary">
                                Middle
                            </button>
                            <button type="button" className="btn btn-outline-primary">
                                Right
                            </button>
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
                                <button type="button" className="btn btn-outline-primary">
                                    Left
                                </button>
                                <button type="button" className="btn btn-outline-primary">
                                    Middle
                                </button>
                                <button type="button" className="btn btn-outline-primary">
                                    Right
                                </button>
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
                            <button
                                key={option}
                                className={classNames('btn', 'btn-outline-secondary', option === active && 'active')}
                                onClick={() => setActive(option)}
                                aria-pressed={option === active}
                            >
                                {option}
                            </button>
                        ))}
                    </div>{' '}
                    Example with <code>btn-outline-secondary</code>
                </div>
                <div className="mb-2">
                    <div className="btn-group" role="group" aria-label="Basic example">
                        {(['Left', 'Middle', 'Right'] as const).map(option => (
                            <button
                                key={option}
                                className={classNames('btn', 'btn-outline-primary', option === active && 'active')}
                                onClick={() => setActive(option)}
                                aria-pressed={option === active}
                            >
                                {option}
                            </button>
                        ))}
                    </div>{' '}
                    Example with <code>btn-outline-primary</code>
                </div>
            </>
        )
    },
    {
        design: {
            type: 'figma',
            name: 'Figma',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A2514',
        },
    }
)

add('Dropdowns', () => (
    <>
        <h1>Dropdowns</h1>
        <p>
            Toggle contextual overlays for displaying lists of links and more with the Bootstrap dropdown component.{' '}
            <a href="https://getbootstrap.com/docs/4.5/components/dropdowns/">Bootstrap documentation</a>
        </p>
        <Menu>
            <MenuButton className="btn btn-secondary dropdown-toggle">Dropdown button</MenuButton>
            <MenuList className="dropdown-menu show" style={{ outline: 'none' }}>
                <h6 className="dropdown-header">Dropdown header</h6>
                <MenuLink
                    className="dropdown-item"
                    href="https://example.com"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    Action
                </MenuLink>
                <MenuLink
                    className="dropdown-item"
                    href="https://example.com"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    Another action
                </MenuLink>
                <div className="dropdown-divider" />
                <MenuLink
                    className="dropdown-item"
                    href="https://example.com"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    Something else here
                </MenuLink>
            </MenuList>
        </Menu>
    </>
))

add('Input groups', () => (
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
                    <button className="btn btn-primary" type="submit">
                        Submit
                    </button>
                </div>
            </div>
        </div>
    </>
))

add(
    'Forms',
    () => (
        <>
            <h1>Forms</h1>
            <p>
                Forms are validated using native HTML validation. Submit the below form with invalid input to try it
                out.{' '}
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
                <button type="submit" className="btn btn-primary">
                    Submit
                </button>
            </Form>

            <h2 className="mt-3">Disabled</h2>
            <Form>
                <fieldset disabled={true}>
                    <div className="form-group">
                        <label htmlFor="disabledTextInput">Disabled input</label>
                        <input
                            type="text"
                            id="disabledTextInput"
                            className="form-control"
                            placeholder="Disabled input"
                        />
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
                    <button type="submit" className="btn btn-primary">
                        Submit
                    </button>
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
    ),
    {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/BkY8Ak997QauG0Iu2EqArv/Sourcegraph-Components?node-id=30%3A24',
        },
    }
)

add('Cards', CardsStory, {
    design: {
        name: 'Figma',
        type: 'figma',
        url:
            'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=1172%3A285',
    },
})

add('List groups', () => (
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
))

add('Meter', () => {
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
})
