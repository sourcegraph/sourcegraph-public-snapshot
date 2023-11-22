import type { FC } from 'react'

import { noop } from 'lodash'

import { Button, ErrorAlert } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../../components/LoaderButton'
import { LimitedAccessLabel } from '../../limited-access-label/LimitedAccessLabel'

export enum CodeInsightCreationMode {
    Creation = 'creation',
    Edit = 'edit',
}

interface CodeInsightsCreationActionsProps {
    mode: CodeInsightCreationMode
    submitting: boolean
    available?: boolean
    errors?: unknown
    licensed?: boolean
    clear?: boolean
    onCancel?: () => void
}

export const CodeInsightsCreationActions: FC<CodeInsightsCreationActionsProps> = props => {
    const { mode, submitting, licensed, available, clear, errors, onCancel = noop } = props

    const isEditMode = mode === CodeInsightCreationMode.Edit

    return (
        <footer>
            {!licensed && !isEditMode && (
                <LimitedAccessLabel
                    message="Unlock Code Insights to create unlimited insights"
                    className="my-3 mt-n2"
                />
            )}

            <div className="d-flex flex-wrap align-items-center">
                {!!errors && <ErrorAlert className="w-100" error={errors} />}

                <LoaderButton
                    type="submit"
                    variant="primary"
                    label={submitting ? 'Submitting' : isEditMode ? 'Save changes' : 'Create code insight'}
                    data-testid="insight-save-button"
                    alwaysShowLabel={true}
                    loading={submitting}
                    disabled={submitting || !available}
                    className="mr-2 mb-2"
                />

                <Button type="button" variant="secondary" outline={true} className="mb-2 mr-auto" onClick={onCancel}>
                    Cancel
                </Button>

                <Button type="reset" variant="secondary" outline={true} disabled={!clear} className="border-0">
                    Clear all fields
                </Button>
            </div>
        </footer>
    )
}
