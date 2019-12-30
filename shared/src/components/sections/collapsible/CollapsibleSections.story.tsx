import { CollapsibleSectionsWithLocalStorageViewStatePersistence } from './CollapsibleSections'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { Section } from '../Sections'
import './index.scss'

const { add } = storiesOf('CollapsibleSections', module).addDecorator(story => (
    // tslint:disable-next-line: jsx-ban-props
    <div style={{ maxWidth: '20rem', margin: '2rem' }}>{story()}</div>
))

type SectionID = 'foo' | 'bar' | 'baz' | 'qux'

const SECTIONS: Section<SectionID>[] = [
    { id: 'foo', label: 'Foo' },
    { id: 'bar', label: 'Bar' },
    {
        id: 'baz',
        label: (
            <div className="flex align-items-center">
                Baz{' '}
                <small style={{ backgroundColor: 'green', color: 'white', borderRadius: '50%', padding: '4px' }}>
                    17
                </small>
            </div>
        ),
    },
    { id: 'qux', label: 'Qux' },
]

add('localStorage', () => (
    <CollapsibleSectionsWithLocalStorageViewStatePersistence
        sections={SECTIONS}
        storageKey="CollapsibleSections.story.simple"
    >
        <div key="foo">
            <h2>Foo</h2>{' '}
            <p>
                foo content
                <br />
                more foo content
            </p>
        </div>
        <div key="bar">
            <h2>Bar</h2>{' '}
            <p>
                bar content
                <br />
                more bar content
            </p>
        </div>
        <div key="baz">
            <h2>Baz</h2>{' '}
            <p>
                baz content
                <br />
                more baz content
            </p>
        </div>
        <div key="qux">
            <h2>Qux</h2>{' '}
            <p>
                qux content
                <br />
                more qux content
            </p>
        </div>
    </CollapsibleSectionsWithLocalStorageViewStatePersistence>
))
