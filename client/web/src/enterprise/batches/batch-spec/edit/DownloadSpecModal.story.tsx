import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'

import { DownloadSpecModal } from './DownloadSpecModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit',
    decorators: [decorator],
}

export default config

export const DownloadSpecModalStory: StoryFn = () => (
    <WebStory>
        {props => (
            <DownloadSpecModal
                setDownloadSpecModalDismissed={function (): void {
                    throw new Error('Function not implemented.')
                }}
                name=""
                originalInput=""
                setIsDownloadSpecModalOpen={function (): void {
                    throw new Error('Function not implemented.')
                }}
                // setDownloadSpecModalDismissed={function (): void {
                //     throw new Error('Function not implemented.')
                // }}
                {...props}
            />
        )}
    </WebStory>
)

DownloadSpecModalStory.storyName = 'DownloadSpecModal'
