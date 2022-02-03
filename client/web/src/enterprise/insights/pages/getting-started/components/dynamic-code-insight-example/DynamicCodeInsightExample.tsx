import classNames from 'classnames'
import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'
import { noop } from 'rxjs'

import { Badge, Button, Card } from '@sourcegraph/wildcard'

import { FormInput } from '../../../../components/form/form-input/FormInput'
import { useField } from '../../../../components/form/hooks/useField'
import { useForm } from '../../../../components/form/hooks/useForm'
import { InsightQueryInput } from '../../../../components/form/query-input/InsightQueryInput'
import { RepositoriesField } from '../../../../components/form/repositories-field/RepositoriesField'
import { DATA_SERIES_COLORS, EditableDataSeries } from '../../../insights/creation/search-insight'
import { getQueryPatternTypeFilter } from '../../../insights/creation/search-insight/components/form-series-input/get-pattern-type-filter'
import { SearchInsightLivePreview } from '../../../insights/creation/search-insight/components/live-preview-chart/SearchInsightLivePreview'
import {
    repositoriesExistValidator,
    repositoriesFieldValidator,
} from '../../../insights/creation/search-insight/components/search-insight-creation-content/validators'

import styles from './DynamicCodeInsightExample.module.scss'

interface CodeInsightExampleFormValues {
    repositories: string
    query: string
}

const INITIAL_INSIGHT_VALUES: CodeInsightExampleFormValues = {
    repositories: 'github.com/sourcegraph/sourcegraph',
    query: 'TODO archived:no fork:no',
}

export const DynamicCodeInsightExample: React.FunctionComponent = () => {
    const form = useForm<CodeInsightExampleFormValues>({
        initialValues: INITIAL_INSIGHT_VALUES,
        onSubmit: noop,
    })

    const repositories = useField({
        name: 'repositories',
        formApi: form.formAPI,
        validators: {
            // Turn off any validations for the repositories field in we are in all repos mode
            sync: repositoriesFieldValidator,
            async: repositoriesExistValidator,
        },
    })

    const query = useField({
        name: 'query',
        formApi: form.formAPI,
    })

    const hasValidLivePreview = repositories.meta.validState === 'VALID' && query.meta.validState === 'VALID'

    const livePreviewSeries: EditableDataSeries[] = hasValidLivePreview
        ? [
              {
                  valid: true,
                  edit: false,
                  id: '1',
                  name: 'TODOs',
                  query: query.input.value,
                  stroke: DATA_SERIES_COLORS.ORANGE,
              },
          ]
        : []

    return (
        <Card>
            <article className={classNames(styles.wrapper)}>
                <form ref={form.ref} noValidate={true} onSubmit={form.handleSubmit}>
                    <Badge variant="primary" className="mb-2">
                        Interactive example
                    </Badge>
                    <SearchInsightLivePreview
                        title="In-line TODO statements"
                        withLivePreviewControls={false}
                        repositories={repositories.input.value}
                        series={livePreviewSeries}
                        stepValue="2"
                        step="months"
                        disabled={!hasValidLivePreview}
                        isAllReposMode={false}
                        className={styles.chart}
                    />

                    <FormInput
                        title="Search query"
                        required={true}
                        as={InsightQueryInput}
                        patternType={getQueryPatternTypeFilter(query.input.value)}
                        placeholder="Example: patternType:regexp const\s\w+:\s(React\.)?FunctionComponent"
                        valid={query.meta.touched && query.meta.validState === 'VALID'}
                        error={query.meta.touched && query.meta.error}
                        className="mt-4"
                        {...query.input}
                    />

                    <FormInput
                        as={RepositoriesField}
                        autoFocus={true}
                        required={true}
                        title="Repositories"
                        placeholder="Example: github.com/sourcegraph/sourcegraph"
                        loading={repositories.meta.validState === 'CHECKING'}
                        valid={repositories.meta.touched && repositories.meta.validState === 'VALID'}
                        error={repositories.meta.touched && repositories.meta.error}
                        {...repositories.input}
                        className="mb-0 d-flex flex-column"
                    />
                </form>

                <section>
                    <h2 className={classNames(styles.cardTitle)}>
                        Draw insights from your codebase about how different initiatives are tracking over time
                    </h2>

                    <p>
                        Create customizable, visual dashboards with meaningful codebase signals your team can use to
                        answer questions about how their code is changing and what’s in their code - questions that were
                        difficult or impossible to answer before.
                    </p>

                    <h3 className={classNames(styles.bulletTitle)}>Use Code Insights to...</h3>

                    <ul>
                        <li>Track migrations, adoption, and deprecations</li>
                        <li>Detect versions of languages, packages, or infrastructure</li>
                        <li>Ensure removal of security vulnerabilities</li>
                        <li>Track code smells, ownership, and configurations</li>
                    </ul>

                    <Button variant="primary">
                        <PlusIcon className="icon-inline" /> Create your first insight
                    </Button>
                </section>
            </article>
        </Card>
    )
}
