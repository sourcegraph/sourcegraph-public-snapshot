import { storiesOf } from '@storybook/react'
import React, { useCallback, useState } from 'react'
import { LoaderInput } from './LoaderInput'
import { WebStory } from '../../../web/src/components/WebStory'

const { add } = storiesOf('shared/LoaderInput', module).addDecorator(story => (
    <div className="container mt-3" style={{ width: 800 }}>
        {story()}
    </div>
))

add('Interactive', () => (
    <WebStory>
        {() => {
            const [loading, setLoading] = useState(true)
            const toggleLoading = useCallback(() => setLoading(loading => !loading), [])

            return (
                <>
                    <button type="button" className="btn btn-primary mb-2" onClick={toggleLoading}>
                        Toggle Loading
                    </button>
                    <p>
                        <LoaderInput loading={loading}>
                            <input type="text" placeholder="Loader input" className="form-control" />
                        </LoaderInput>
                    </p>
                </>
            )
        }}
    </WebStory>
))
