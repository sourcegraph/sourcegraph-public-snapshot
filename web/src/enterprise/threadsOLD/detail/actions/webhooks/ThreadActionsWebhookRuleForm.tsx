import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import CheckIcon from 'mdi-react/CheckIcon'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { ThreadSettings, WebhookRule } from '../../../settings'
import { ThreadSettingsForm } from '../components/threadSettingsForm/ThreadSettingsForm'

interface Props extends ExtensionsControllerProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind' | 'title' | 'settings'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings
    extraAction?: JSX.Element | null
}

// tslint:disable: jsx-no-lambda
export const ThreadActionsWebhookRuleForm: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    extensionsController,
}) => {
    const DEFAULTS: WebhookRule = {
        url: '',
    }
    const initialUncommittedSettings: ThreadSettings = {
        ...threadSettings,
        webhooks: [
            {
                ...DEFAULTS,
                ...(threadSettings.webhooks ? threadSettings.webhooks[0] : null),
            },
        ],
    }

    return (
        <ThreadSettingsForm
            thread={thread}
            onThreadUpdate={onThreadUpdate}
            threadSettings={threadSettings}
            initialUncommittedSettings={initialUncommittedSettings}
            extensionsController={extensionsController}
        >
            {({ uncommittedSettings, setUncommittedSettings, isLoading }) => {
                const updateUncommittedSettings = (data: Partial<WebhookRule>) =>
                    setUncommittedSettings({
                        ...uncommittedSettings,
                        webhooks: [
                            {
                                ...(uncommittedSettings.webhooks ? uncommittedSettings.webhooks[0] : null),
                                ...data,
                            },
                        ],
                    })
                const { url } = uncommittedSettings.webhooks ? uncommittedSettings.webhooks[0] : DEFAULTS

                return (
                    <>
                        <div className="form-group">
                            <label htmlFor="thread-actions-webhook-notification-rule-form__url">URL</label>
                            <input
                                type="url"
                                className="form-control"
                                id="thread-actions-webhook-notification-rule-form__url"
                                placeholder="https://example.com/receive-hook"
                                value={url}
                                onChange={e =>
                                    updateUncommittedSettings({
                                        url: e.currentTarget.value,
                                    })
                                }
                                required={true}
                            />
                        </div>
                        <div>
                            <button type="submit" className="btn btn-primary mr-2" disabled={isLoading}>
                                {isLoading ? (
                                    <LoadingSpinner className="icon-inline" />
                                ) : (
                                    <CheckIcon className="icon-inline" />
                                )}{' '}
                                Save
                            </button>
                        </div>
                    </>
                )
            }}
        </ThreadSettingsForm>
    )
}
