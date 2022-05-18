import React from 'react'
import 'storybook-addon-designs'

import { Link, Typography } from '@sourcegraph/wildcard'

import { SEMANTIC_COLORS } from './constants'
import { TextVariants } from './TextVariants'

export const TextStory: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <>
        <Typography.H2>Headings</Typography.H2>
        <table className="table">
            <tbody>
                {(['h1', 'h2', 'h3', 'h4', 'h5', 'h6'] as const).map(Heading => (
                    <tr key={Heading}>
                        <td>
                            <code>
                                {'<'}
                                {Heading}
                                {'>'}
                            </code>
                        </td>
                        <td>
                            <Heading>This is an {Heading.toUpperCase()}</Heading>
                        </td>
                    </tr>
                ))}
            </tbody>
        </table>

        <Typography.H2>Text variations</Typography.H2>
        <TextVariants />

        <Typography.H2>Prose</Typography.H2>
        <p>Text uses system fonts. The fonts should never be overridden.</p>
        <p>
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
        </p>

        <p>
            Text can contain links, which <Link to="/">trigger a navigation to a different page</Link>.
        </p>

        <p>
            Text can be <em>emphasized</em> or made <strong>strong</strong>.
        </p>

        <p>
            Text can be <i>idiomatic</i> with <code>{'<i>'}</code>. See{' '}
            <Link
                target="__blank"
                to="https://developer.mozilla.org/en-US/docs/Web/HTML/Element/em#%3Ci%3E_vs._%3Cem%3E"
            >
                {'<i>'} vs. {'<em>'}
            </Link>{' '}
            for more info.
        </p>

        <p>
            You can bring attention to the <b>element</b> with <code>{'<b>'}</code>.
        </p>

        <p>
            Text can have superscripts<sup>sup</sup> with <code>{'<sup>'}</code>.
        </p>

        <p>
            Text can have subscripts<sub>sub</sub> with <code>{'<sub>'}</code>.
        </p>

        <p>
            <small>
                You can use <code>{'<small>'}</code> to make small text. Use sparingly.
            </small>
        </p>

        <Typography.H2>Color variations</Typography.H2>
        <p>
            <code>text-*</code> classes can be used to apply semantic coloring to text.
        </p>
        <div className="mb-3">
            {['muted', ...SEMANTIC_COLORS].map(color => (
                <div key={color} className={'text-' + color}>
                    This is text-{color}
                </div>
            ))}
        </div>

        <Typography.H2>Lists</Typography.H2>
        <Typography.H3>Ordered</Typography.H3>
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

        <Typography.H3>Unordered</Typography.H3>

        <Typography.H4>Dots</Typography.H4>
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

        <Typography.H4>Dashes</Typography.H4>
        <p>
            Dashed lists are created using <code>list-dashed</code>.
        </p>
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
