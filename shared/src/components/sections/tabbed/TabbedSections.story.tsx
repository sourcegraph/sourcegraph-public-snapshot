import { storiesOf } from '@storybook/react'
import React from 'react'
import { Section } from '../Sections'
import { TabbedSectionsWithLocalStorageViewStatePersistence } from './TabbedSections'
import './index.scss'

const { add } = storiesOf('TabbedSections', module).addDecorator(story => (
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

add('simple', () => (
    <TabbedSectionsWithLocalStorageViewStatePersistence sections={SECTIONS} storageKey="TabbedSections.story.simple">
        <div key="foo">Foo content</div>
        <div key="bar">Bar content</div>
        <div key="baz">Baz content</div>
        <div key="qux">Qux content</div>
    </TabbedSectionsWithLocalStorageViewStatePersistence>
))
