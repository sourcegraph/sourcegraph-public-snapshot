import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import CheckIcon from 'mdi-react/CheckIcon'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { CommitStatusRule, ThreadSettings } from '../../../settings'
import { ThreadSettingsForm } from '../components/threadSettingsForm/ThreadSettingsForm'

interface Props extends ExtensionsControllerProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'title' | 'settings'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings
    extraAction?: JSX.Element | null
}

// tslint:disable: jsx-no-lambda
export const ThreadActionsCommitStatusRuleForm: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    extensionsController,
}) => {
    const DEFAULTS: CommitStatusRule = {
        branch: '',
        infoOnly: false,
        enabled: false,
    }
    const initialUncommittedSettings: ThreadSettings = {
        ...threadSettings,
        commitStatusRules: [
            {
                ...DEFAULTS,
                ...(threadSettings.commitStatusRules ? threadSettings.commitStatusRules[0] : null),
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
                const updateUncommittedSettings = (data: Partial<CommitStatusRule>) =>
                    setUncommittedSettings({
                        ...uncommittedSettings,
                        commitStatusRules: [
                            {
                                ...(uncommittedSettings.commitStatusRules
                                    ? uncommittedSettings.commitStatusRules[0]
                                    : null),
                                ...data,
                            },
                        ],
                    })
                const { branch, infoOnly, enabled } = uncommittedSettings.commitStatusRules
                    ? uncommittedSettings.commitStatusRules[0]
                    : DEFAULTS

                return (
                    <>
                        <div className="form-group">
                            <label htmlFor="thread-actions-commit-status-rule-form__branch">Branch</label>
                            <input
                                type="text"
                                className="form-control"
                                id="thread-actions-commit-status-rule-form__branch"
                                placeholder="master"
                                value={branch}
                                onChange={e =>
                                    updateUncommittedSettings({
                                        branch: e.currentTarget.value,
                                    })
                                }
                            />
                        </div>
                        <div className="form-check mb-3">
                            <input
                                className="form-check-input"
                                type="checkbox"
                                id="thread-actions-commit-status-rule-form__infoOnly"
                                checked={infoOnly}
                                onChange={e =>
                                    updateUncommittedSettings({
                                        infoOnly: e.currentTarget.checked,
                                    })
                                }
                            />
                            <label
                                className="form-check-label"
                                htmlFor="thread-actions-commit-status-rule-form__infoOnly"
                            >
                                Informational only (never fail build)
                            </label>
                        </div>
                        <div className="form-check mb-3">
                            <input
                                className="form-check-input"
                                type="checkbox"
                                id="thread-actions-commit-status-rule-form__enabled"
                                checked={enabled}
                                onChange={e =>
                                    updateUncommittedSettings({
                                        enabled: e.currentTarget.checked,
                                    })
                                }
                            />
                            <label
                                className="form-check-label"
                                htmlFor="thread-actions-commit-status-rule-form__enabled"
                            >
                                Enabled
                            </label>
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
