import React from 'react'

import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { Page } from '../../../../../../components/Page'
import { PageTitle } from '../../../../../../components/PageTitle'
import { FormChangeEvent } from '../../../../components/form/hooks/useForm'

import { CaptureGroupCreationContent } from './components/CaptureGroupCreationContent'
import { CaptureGroupFormFields } from './types'

export const CaptureGroupCreationPage: React.FunctionComponent = props => {
    const [initialFormValues, setInitialFormValues] = useLocalStorage<CaptureGroupFormFields | undefined>(
        'insights.code-stats-creation-ui',
        undefined
    )

    const handleSubmit = (values: CaptureGroupFormFields): void => {
        console.log(values)
    }

    // eslint-disable-next-line unicorn/consistent-function-scoping
    const handleCancel = (): void => {
        console.log('CaptureGroupCreationPage Cancel')
    }

    const handleChange = (event: FormChangeEvent<CaptureGroupFormFields>): void => {
        setInitialFormValues(event.values)
    }

    return (
        <Page>
            <PageTitle title="Create new capture group code insight" />

            <header className="mb-5">
                <h2>Create new code insight</h2>

                <p className="text-muted">
                    Search-based code insights analyze your code based on any search query.{' '}
                    <a href="https://docs.sourcegraph.com/code_insights" target="_blank" rel="noopener">
                        Learn more.
                    </a>
                </p>
            </header>

            <CaptureGroupCreationContent
                mode="creation"
                className="pb-5"
                initialValues={initialFormValues}
                onSubmit={handleSubmit}
                onCancel={handleCancel}
                onChange={handleChange}
            />
        </Page>
    )
}
