import React from 'react'

import { UsageExamplesBox } from '../src/usageExamples/UsageExamplesBox'

export const Sandbox: React.FunctionComponent = () => (
    <div>
        <h1>Sourcebox sandbox</h1>

        <h2>React.Children</h2>

        <h3>
            <code>Children.count(children)</code>
        </h3>
        <p>
            Call <code>Children.count(children)</code> to count the number of children in the <code>children</code> data
            structure.
        </p>
        <h4>Parameters</h4>
        <ul>
            <li>
                <code>children</code>: The value of the <code>children</code> prop received by your component.
            </li>
        </ul>
        <h4>Returns</h4>
        <p>
            The number of nodes inside these <code>children</code>.
        </p>
        <h4>Caveats</h4>
        <ul>
            <li>
                Empty nodes (<code>null</code>, <code>undefined</code>, and Booleans), strings, numbers, and React
                elements count as individual nodes. Arrays don't count as individual nodes, but their children do. The
                traversal does not go deeper than React elements: they don't get rendered, and their children aren't
                traversed. Fragments don't get traversed.
            </li>
        </ul>
        <h4>Usage examples</h4>
        <UsageExamplesBox />

        <br />

        <section style={{ display: 'none' }}>
            <h3>
                <code>Children.forEach(children, fn, thisArg?)</code>
            </h3>
            <p>
                Call <code>Children.forEach(children, fn, thisArg?)</code> to run some code for each child in the{' '}
                <code>children</code> data structure.
            </p>
            <h4>Parameters</h4>
            <ul>
                <li>
                    <code>children</code>: The value of the <code>children</code> prop received by your component.
                </li>
                <li>
                    <code>fn</code>: The function you want to run for each child, similar to the array{' '}
                    <code>forEach</code> method callback. It will be called with the child as the first argument and its
                    index as the second argument. The index starts at <code>0</code> and increments on each call.
                </li>
            </ul>
            <h4>Returns</h4>
            <p>
                <code>Children.forEach</code> returns <code>undefined</code>.
            </p>
            <h4>Caveats</h4>
            <ul>
                <li>
                    Empty nodes (<code>null</code>, <code>undefined</code>, and Booleans), strings, numbers, and React
                    elements count as individual nodes. Arrays don't count as individual nodes, but their children do.
                    The traversal does not go deeper than React elements: they don't get rendered, and their children
                    aren't traversed. Fragments don't get traversed.
                </li>
            </ul>
        </section>
    </div>
)
