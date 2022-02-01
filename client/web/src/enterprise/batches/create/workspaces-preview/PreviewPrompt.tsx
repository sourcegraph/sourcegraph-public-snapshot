import SearchIcon from 'mdi-react/SearchIcon'
import React from 'react'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'
import { Button } from '@sourcegraph/wildcard'

import styles from './PreviewPrompt.module.scss'
import { PreviewPromptIcon } from './PreviewPromptIcon'

/** Example snippet show in preview prompt if user has not yet added an on: statement. */
const ON_STATEMENT = `on:
  - repositoriesMatchingQuery: repo:my-org/.*
`

/**
 * The preview prompt shows different elements depending on the state of the editor and
 * workspaces preview resolution.
 * - Initial: If the user has not yet requested any workspaces preview.
 * - Error: If the latest workspaces preview request failed to reach a resolution.
 * - Update: If the user has requested a workspaces preview before but has made changes to
 * their batch spec input YAML since the last time it had a resolution.
 */
export type PreviewPromptForm = 'Initial' | 'Error' | 'Update'

interface PreviewPromptProps {
    /**
     * Function to submit the current input batch spec YAML to trigger a workspaces
     * preview request.
     */
    preview: () => void
    /**
     * Whether or not the preview button should be disabled due to their being a problem
     * with the input batch spec YAML, or a preview request is already happening. An
     * optional tooltip string to display may be provided in place of `true`.
     */
    disabled: boolean | string
    form: PreviewPromptForm
}

/**
 * The preview prompt provides a CTA for users to submit their working batch spec YAML to
 * the backend in order to preview the workspaces it will affect.
 */
export const PreviewPrompt: React.FunctionComponent<PreviewPromptProps> = ({ preview, disabled, form }) => {
    const previewButton = (
        <Button
            className="mb-2"
            variant="success"
            disabled={!!disabled}
            data-tooltip={typeof disabled === 'string' ? disabled : undefined}
            onClick={preview}
        >
            <SearchIcon className="icon-inline mr-1" />
            Preview workspaces
        </Button>
    )

    switch (form) {
        case 'Initial':
            return (
                <>
                    <PreviewPromptIcon className="mt-4" />
                    <h4 className={styles.header}>
                        Use an <span className="text-monospace">on:</span> statement to preview repositories.
                    </h4>
                    {previewButton}
                    <div className={styles.onExample}>
                        <h3 className="align-self-start pt-4 text-muted">Example:</h3>
                        <CodeSnippet className="w-100" code={ON_STATEMENT} language="yaml" />
                    </div>
                </>
            )
        case 'Error':
            return previewButton
        case 'Update':
            return (
                <>
                    <h4 className={styles.header}>
                        Finish editing your batch spec, then manually preview repositories.
                    </h4>
                    {previewButton}
                    <div className="mb-4" />
                </>
            )
    }
}
