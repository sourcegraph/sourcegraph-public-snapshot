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

const SEMANTIC_COLORS = ['primary', 'secondary', 'success', 'danger', 'warning', 'info', 'merged'] as const
export const Prose: Story = () => (
    <>
        <H2>Prose</H2>
        <Text>Text uses system fonts. The fonts should never be overridden.</Text>
        <Text>
            Minim nisi tempor Lorem do incididunt exercitation ipsum consectetur laboris elit est aute irure velit.
            Voluptate irure excepteur sint reprehenderit culpa laboris. Elit id nostrud enim laboris irure. Est sunt ex
            ipisicing aute elit voluptate consectetur.Do laboris anim fugiat ipsum sunt elit sunt amet consequat trud
            irure labore cupidatat laboris. Voluptate eiusmod veniam nisi reprehenderit cillum Lorem veniam at amet ea
            dolore enim. Ea laborum fugiat Lorem ea amet amet exercitation dolor culpa. Do consequat dolor ad elit ipsum
            nostrud non laboris voluptate aliquip est reprehenderit incididunt. Eu nulla ad te enim. Pariatur duis
            pariatur sit adipisicing pariatur nulla quis do sint deserunt aliqua Lorem aborum. Dolor esse aute cupidatat
            deserunt anim ad eiusmod quis quis laborum magna nisi occaecat. Eu is eiusmod sint aliquip duis est sit
            irure velit reprehenderit id. Cillum est esse et nulla ut adipisicing velit anim id exercitation nostrud.
            Duis veniam sit laboris tempor quis sit cupidatat elit.
        </Text>

        <Text>
            Text can contain links, which <Link to="/">trigger a navigation to a different page</Link>.
        </Text>

        <Text>
            Text can be <em>emphasized</em> or made <strong>strong</strong>.
        </Text>

        <Text>
            Text can be <i>idiomatic</i> with <Code>{'<i>'}</Code>. See{' '}
            <Link
                target="__blank"
                to="https://developer.mozilla.org/en-US/docs/Web/HTML/Element/em#%3Ci%3E_vs._%3Cem%3E"
            >
                {'<i>'} vs. {'<em>'}
            </Link>{' '}
            for more info.
        </Text>

        <Text>
            You can bring attention to the <b>element</b> with <Code>{'<b>'}</Code>.
        </Text>

        <Text>
            Text can have superscripts<sup>sup</sup> with <Code>{'<sup>'}</Code>.
        </Text>

        <Text>
            Text can have subscripts<sub>sub</sub> with <Code>{'<sub>'}</Code>.
        </Text>

        <Text>
            <small>
                You can use <Code>{'<small>'}</Code> to make small text. Use sparingly.
            </small>
        </Text>

        <H2>Color variations</H2>
        <Text>
            <Code>text-*</Code> classes can be used to apply semantic coloring to text.
        </Text>
        <div className="mb-3">
            {['muted', ...SEMANTIC_COLORS].map(color => (
                <div key={color} className={'text-' + color}>
                    This is text-{color}
                </div>
            ))}
        </div>

        <H2>Lists</H2>
        <H3>Ordered</H3>
        <ol>
            <li>
                Dolor est laborum aute adipisicing quis duis mollit pariatur nostrud eiusmod Lorem pariatur elit mollit.
                Sint pariatur culpa occaecat aute mollit enim amet nisi sunt aute ea aliqua esse laboris. Incididunt ad
                duis laborum elit dolore esse sint nisi. Nulla in ea ipsum dolore irure sit labore commodo aute aliquip
                esse. Consectetur non tempor qui sunt cillum est velit ut id sint id amet et commodo.
            </li>
            <li>
                Eu nulla Lorem et ipsum commodo. Sint anim minim aute deserunt elit adipisicing minim sunt est tempor.
                Exercitation non ad minim culpa fugiat nulla nulla.
            </li>
            <li>
                Ex officia amet excepteur Lorem officia sit elit. Aute esse laboris consequat ea sint aute amet anim.
                Laboris dolore dolor Lorem anim voluptate eiusmod nisi occaecat anim ipsum laboris ad.
            </li>
        </ol>

        <H3>Unordered</H3>

        <H4>Dots</H4>
        <ul>
            <li>
                Ullamco exercitation voluptate veniam et in incididunt Lorem id consequat dolor reprehenderit amet. Id
                exercitation et labore do sint eiusmod irure. Lorem cupidatat dolor nulla sunt qui culpa esse cupidatat
                ea. Esse elit voluptate ea officia excepteur nostrud veniam dolore tempor sint anim dolor ipsum eu.
            </li>
            <li>
                Magna veniam in anim ea cupidatat nostrud. Pariatur mollit eiusmod incididunt irure pariatur amet. Est
                adipisicing voluptate nulla Lorem esse laborum aliqua.
            </li>
            <li>
                Proident nisi velit incididunt labore sunt eiusmod magna occaecat aliqua. Labore veniam ex adipisicing
                ex magna qui officia dolor. Eiusmod excepteur dolor consequat deserunt enim ullamco eiusmod ullamco.
            </li>
        </ul>

        <H4>Dashes</H4>
        <Text>
            Dashed lists are created using <Code>list-dashed</Code>.
        </Text>
        <ul className="list-dashed">
            <li>
                Ad deserunt amet Lorem in exercitation. Deserunt labore anim non minim. Dolor dolore adipisicing anim
                cupidatat nulla. Sit voluptate aliqua exercitation occaecat nulla aute ex quis excepteur quis
                exercitation fugiat et. Voluptate sint magna labore culpa nulla eu tempor labore in eiusmod excepteur.
            </li>
            <li>
                Quis do proident non deserunt aliquip eiusmod dolor nisi et eiusmod irure labore irure. Veniam labore
                aliquip ea irure dolore est cillum laborum exercitation. Anim pariatur occaecat reprehenderit ea et elit
                excepteur nisi mollit tempor. Consequat ullamco do velit irure laboris adipisicing nulla enim.
            </li>
            <li>
                Incididunt occaecat consequat aliqua fugiat sint veniam anim cupidatat. Laborum ex aliqua quis et labore
                laboris. Quis laborum excepteur do nisi proident dolor duis sint cupidatat commodo proident sunt. Tempor
                nisi consectetur ex culpa occaecat. Qui mollit mollit reprehenderit ea consequat quis aliqua minim anim
                ullamco ullamco incididunt duis amet. Occaecat anim adipisicing laborum excepteur mollit do ullamco id
                fugiat duis.
            </li>
        </ul>
    </>
)
