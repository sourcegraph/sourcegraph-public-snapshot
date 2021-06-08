import { storiesOf } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Container } from './Container'

const { add } = storiesOf('wildcard/Container', module).addDecorator(story => (
    <BrandedStory styles={webStyles}>{() => <div className="container web-content mt-3">{story()}</div>}</BrandedStory>
))

add(
    'Overview',
    () => (
        <>
            <div className="alert alert-info">
                <p>
                    A container is meant to group content semantically together. Every page using it should have a
                    header, optionally a description for the page and the container itself. Depending on the scope of a
                    button, it should live inside or outside of the container.
                </p>
                <p>If the button</p>
                <ul className="mb-0">
                    <li>
                        affects everything inside the container (ie. saves all form fields within the container), it
                        should live outside of the container. See example 1
                    </li>
                    <li>
                        affects just a subset of content inside the container (ie. submits one of multiple forms), it
                        should live inside of the container, next to the content it is modifying. See example 2
                    </li>
                </ul>
            </div>
            <hr />
            <h1>Example 1</h1>
            <h2>Some page explanation</h2>
            <p className="text-muted">Optional: Add some descriptive text about what this page does.</p>
            <Container className="mb-3">
                <h3>Section I</h3>
                <p>Here you change the username.</p>
                <div className="form-group">
                    <input type="text" className="form-control" />
                </div>
                <h3>Section II</h3>
                <p>Here you change your email.</p>
                <div className="form-group mb-0">
                    <input type="text" className="form-control" />
                </div>
            </Container>
            <div className="mb-3">
                <button type="button" className="btn btn-primary mr-2">
                    Save
                </button>
                <button type="button" className="btn btn-secondary">
                    Cancel
                </button>
            </div>
            <hr />
            <h1>Example 2</h1>
            <h2>Some page explanation</h2>
            <p className="text-muted">Optional: Add some descriptive text about what this page does.</p>
            <Container className="mb-3">
                <h3>Section I</h3>
                <p>Here you change the username.</p>
                <div className="form-group">
                    <input type="text" className="form-control" />
                </div>
                <button type="button" className="btn btn-secondary mb-2">
                    Save
                </button>
                <hr className="mb-2" />
                <h3>Section II</h3>
                <p>Here you change your email.</p>
                <div className="form-group">
                    <input type="text" className="form-control" />
                </div>
                <button type="button" className="btn btn-secondary">
                    Save
                </button>
            </Container>
        </>
    ),
    {
        design: {
            type: 'figma',
            name: 'Figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/?node-id=1478%3A3044',
        },
    }
)
