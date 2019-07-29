import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import RayEndArrowIcon from 'mdi-react/RayEndArrowIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { ThreadSettings } from '../../../settings'
import { ThreadSettingsForm } from '../components/threadSettingsForm/ThreadSettingsForm'

interface Props extends ExtensionsControllerProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'title' | 'settings'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings
    extraAction?: JSX.Element | null
}

// tslint:disable: jsx-no-lambda
export const ThreadPullRequestTemplateEditForm: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    extraAction,
    extensionsController,
}) => {
    const titlePlaceholder = `${thread.title} (Sourcegraph check)`
    const branchPlaceholder = `check/${thread.title.replace(/[^\w]+/g, '_')}`
    const descriptionDefaultValue =
        // tslint:disable-next-line: no-invalid-template-strings
        'This change was automatically created as part of [Sourcegraph check ${check_number}](${check_url}).\n\nOther PRs created by this check:\n\n${related_links}'

    const initialUncommittedSettings: ThreadSettings = {
        ...threadSettings,
        pullRequestTemplate: {
            title: titlePlaceholder,
            branch: branchPlaceholder,
            description: descriptionDefaultValue,
            ...threadSettings.pullRequestTemplate,
        },
    }

    return (
        <ThreadSettingsForm
            thread={thread}
            onThreadUpdate={onThreadUpdate}
            threadSettings={threadSettings}
            initialUncommittedSettings={initialUncommittedSettings}
            extensionsController={extensionsController}
        >
            {({ uncommittedSettings, setUncommittedSettings, isLoading }) => (
                <>
                    <div className="form-group">
                        <label htmlFor="thread-pull-request-template-edit-form__title">Pull request title</label>
                        <input
                            type="text"
                            className="form-control"
                            id="thread-pull-request-template-edit-form__title"
                            placeholder={titlePlaceholder}
                            onChange={e =>
                                setUncommittedSettings({
                                    ...uncommittedSettings,
                                    pullRequestTemplate: {
                                        ...uncommittedSettings.pullRequestTemplate,
                                        title: e.currentTarget.value,
                                    },
                                })
                            }
                        />
                    </div>
                    <div className="form-group">
                        <label htmlFor="thread-pull-request-template-edit-form__branchName">Branch</label>
                        <div className="d-flex align-items-center">
                            <code
                                className="border rounded text-muted p-1"
                                data-tooltip="Changing the base branch is not yet supported"
                            >
                                <SourceBranchIcon className="icon-inline mr-1" />
                                master
                            </code>{' '}
                            <RayEndArrowIcon className="icon-inline mx-2 text-muted" />
                            <input
                                type="text"
                                className="form-control form-control-sm flex-0 w-auto text-monospace"
                                id="thread-pull-request-template-edit-form__branchName"
                                placeholder={branchPlaceholder}
                                size={30}
                                onChange={e =>
                                    setUncommittedSettings({
                                        ...uncommittedSettings,
                                        pullRequestTemplate: {
                                            ...uncommittedSettings.pullRequestTemplate,
                                            branch: e.currentTarget.value,
                                        },
                                    })
                                }
                            />
                        </div>
                    </div>
                    <div className="form-group">
                        <label htmlFor="thread-pull-request-template-edit-form__description">
                            Pull request description
                        </label>
                        <textarea
                            className="form-control"
                            id="thread-pull-request-template-edit-form__description"
                            aria-describedby="thread-pull-request-template-edit-form__description-help"
                            rows={7}
                            value={
                                uncommittedSettings.pullRequestTemplate
                                    ? uncommittedSettings.pullRequestTemplate.description
                                    : ''
                            }
                            onChange={e =>
                                setUncommittedSettings({
                                    ...uncommittedSettings,
                                    pullRequestTemplate: {
                                        ...uncommittedSettings.pullRequestTemplate,
                                        description: e.currentTarget.value,
                                    },
                                })
                            }
                        />
                        <small
                            id="thread-pull-request-template-edit-form__description-help"
                            className="form-text text-muted"
                        >
                            Template variables:{' '}
                            <code data-tooltip="The check number (example: #49)">
                                {'${check_number}'} {/* tslint:disable-line: no-invalid-template-strings */}
                            </code>{' '}
                            &nbsp;{' '}
                            <code data-tooltip="The URL to the check's page on Sourcegraph">
                                {'${check_url}'} {/* tslint:disable-line: no-invalid-template-strings */}
                            </code>{' '}
                            &nbsp;{' '}
                            <code data-tooltip="Formatted links to all other pull requests (in other repositories) created by this codemod">
                                {'${related_links}'} {/* tslint:disable-line: no-invalid-template-strings */}
                            </code>
                        </small>
                    </div>
                    <div>
                        <button
                            type="submit"
                            className="btn btn-primary mr-2"
                            disabled={isLoading}
                            aria-describedby="thread-pull-request-template-edit-form__button-help"
                        >
                            {isLoading ? (
                                <LoadingSpinner className="icon-inline" />
                            ) : (
                                <SourcePullIcon className="icon-inline" />
                            )}{' '}
                            Save pull request template
                        </button>
                        {extraAction}
                    </div>
                </>
            )}
        </ThreadSettingsForm>
    )
}
