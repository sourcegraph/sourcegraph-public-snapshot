import React from 'react'

import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { DownloadSpecModal } from './DownloadSpecModal'

const { add } = storiesOf('web/batches/create/DownloadSpecModal', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('Download Spec Modal', () => (
    <WebStory>
        {props => (
            <DownloadSpecModal
                name=""
                originalInput=""
                setIsDownloadSpecModalOpen={function (condition: boolean): void {
                    throw new Error('Function not implemented.')
                }}
                setDownloadSpecModalDismissed={function (condition: boolean): void {
                    throw new Error('Function not implemented.')
                }}
                {...props}
            />
        )}
    </WebStory>
))
