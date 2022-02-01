import { ParentSize } from '@visx/responsive'
import classNames from 'classnames'
import PlusIcon from 'mdi-react/PlusIcon'
import React from 'react'
import { noop } from 'rxjs'
import { LineChartContent } from 'sourcegraph'

import { Button, Card } from '@sourcegraph/wildcard'

import { LineChart } from '../../../../../../views/components/view/content/chart-view-content/charts/line/LineChart'
import { useField } from '../../../../components/form/hooks/useField'
import { useForm } from '../../../../components/form/hooks/useForm'
import { CreateInsightFormFields } from '../../../insights/creation/search-insight/types'
import { useSearchInsightInitialValues } from '../../../insights/creation/search-insight/utils/use-initial-values'

import styles from './DynamicCodeInsightExample.module.scss'

const DYNAMIC_EXAMPLE_STATIC_DATA: LineChartContent<any, string> = {
    chart: 'line' as const,
    data: [
        { x: 1588965700286 - 6 * 24 * 60 * 60 * 1000, a: 400 },
        { x: 1588965700286 - 5 * 24 * 60 * 60 * 1000, a: 700 },
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: 300 },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: 320 },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 200 },
        { x: 1588965700286, a: 190 },
    ],
    series: [
        {
            dataKey: 'a',
            name: 'TODOs',
            stroke: 'var(--oc-orange-7)',
        },
    ],
    xAxis: {
        dataKey: 'x',
        scale: 'time' as const,
        type: 'number',
    },
}

export const DynamicCodeInsightExample: React.FunctionComponent = () => {
    // const { series, repositories } = useInsightCreationForm({ mode: 'creation', subjects: [], onSubmit: noop })
    const { initialValues } = useSearchInsightInitialValues()
    const form = useForm<Partial<CreateInsightFormFields>>({
        initialValues,
        onSubmit: noop,
    })
    const repositories = useField({
        name: 'repositories',
        formApi: form.formAPI,
        // validators: {
        //     // Turn off any validations for the repositories field in we are in all repos mode
        //     sync: repositoriesFieldValidator,
        //     async: repositoriesExistValidator,
        // },
    })
    return (
        <Card>
            <article className={classNames(styles.wrapper)}>
                <section>
                    <ParentSize>
                        {({ width, height }) => (
                            <LineChart {...DYNAMIC_EXAMPLE_STATIC_DATA} width={width} height={height} />
                        )}
                    </ParentSize>

                    {/* <FormInput
                        as={RepositoriesField}
                        autoFocus={true}
                        required={true}
                        title="Repositories"
                        description="Separate repositories with commas"
                        placeholder="Example: github.com/sourcegraph/sourcegraph"
                        loading={repositories.meta.validState === 'CHECKING'}
                        valid={repositories.meta.touched && repositories.meta.validState === 'VALID'}
                        error={repositories.meta.touched && repositories.meta.error}
                        {...repositories.input}
                        className="mb-0 d-flex flex-column"
                    /> */}
                </section>
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
                        <li>Understand your code by team or repository</li>
                        <li>Track code smells and health </li>
                        <li>Visualize configurations like CI connections or service info</li>
                    </ul>

                    <Button variant="primary">
                        <PlusIcon className="icon-inline" /> Create your first insight
                    </Button>

                    <p className={styles.calloutWrapper}>
                        <svg
                            width="59"
                            height="41"
                            viewBox="0 0 59 41"
                            fill="none"
                            xmlns="http://www.w3.org/2000/svg"
                            className={styles.callout}
                        >
                            <path
                                d="M3.23717 0.288488C2.84421 0.157502 2.41947 0.369872 2.28849 0.762829L0.15395 7.16644C0.0229642 7.5594 0.235334 7.98414 0.628292 8.11512C1.02125 8.24611 1.44599 8.03374 1.57698 7.64078L3.47434 1.94868L9.16644 3.84605C9.5594 3.97704 9.98414 3.76467 10.1151 3.37171C10.2461 2.97875 10.0337 2.55401 9.64078 2.42302L3.23717 0.288488ZM57.9254 40.7463C58.3375 40.7875 58.7051 40.4868 58.7463 40.0746C58.7875 39.6625 58.4868 39.2949 58.0746 39.2537L57.9254 40.7463ZM2.32918 1.33541C14.452 25.5811 37.6871 38.7224 57.9254 40.7463L58.0746 39.2537C38.3129 37.2776 15.548 24.4189 3.67082 0.66459L2.32918 1.33541Z"
                                fill="#A6B6D9"
                            />
                        </svg>
                        This insight is interactive! Type any search query or change the repo.
                    </p>
                </section>
            </article>
        </Card>
    )
}
