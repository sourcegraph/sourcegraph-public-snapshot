import { storiesOf } from '@storybook/react'
import React from 'react'
import { WebStory } from '../../components/WebStory'
import { CodeHostExtensionModal } from './GoToCodeHostAction'

const { add } = storiesOf('web/repo/actions', module).addDecorator(story => (
    <div className="container mt-3">{story()}</div>
))

add('CodeHostExtensionModal (GitHub)', () => {
    function noopCallback() {
        // noop
    }

    return (
        <WebStory>
            {() => (
                <div>
                    <CodeHostExtensionModal
                        url="https://github.com/sourcegraph/sourcegraph"
                        serviceType="github"
                        onClose={noopCallback}
                        onRejection={noopCallback}
                        onClickInstall={noopCallback}
                    />
                </div>
            )}
        </WebStory>
    )
})

/**
 * use for itest
 *
 *           <GoToCodeHostAction
                        revision="main"
                        repo={{
                            externalURLs: [
                                {
                                    __typename: 'ExternalLink',
                                    url: 'https://github.com/sourcegraph/sourcegraph',
                                    serviceType: 'github',
                                },
                            ],
                            name: 'github.com/sourcegraph/sourcegraph',
                            defaultBranch: { displayName: 'main' } as IGitRef,
                        }}
                        browserExtensionInstalled={of(false)}
                    />
 */
