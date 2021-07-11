import classnames from 'classnames';
import React from 'react'
import { Link } from 'react-router-dom'
import { ErrorAlert } from '../../../../../../../../components/alerts'
import { FormInput } from '../../../../../../../components/form/form-input/FormInput';
import { useCheckboxes } from '../../../../../../../components/form/hooks/useCheckboxes';
import { useField } from '../../../../../../../components/form/hooks/useField';
import { SubmissionErrors, useForm, FORM_ERROR } from '../../../../../../../components/form/hooks/useForm';
import { ReachableInsight } from '../../hooks/use-reachable-insights';

import styles from './AddInsightModalContent.module.scss'
import { TruncatedText } from '../../../dashboard-select/components/trancated-text/TrancatedText';
import { Badge } from '../../../dashboard-select/components/badge/Badge';
import { LoaderButton } from '../../../../../../../../components/LoaderButton';

interface AddInsightModalContentProps {
    insights: ReachableInsight[]
    initialValues: AddInsightFormValues
    onSubmit: (values: AddInsightFormValues) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export interface AddInsightFormValues {
    searchInput: string
    insightIds: string[]
}

export const AddInsightModalContent: React.FunctionComponent<AddInsightModalContentProps> = props => {
    const { initialValues, insights, onSubmit, onCancel } = props

    const { formAPI, ref, handleSubmit, } = useForm({
        initialValues,
        onSubmit
    })

    const searchInput = useField('searchInput', formAPI)
    const { input: { isChecked, onChange, onBlur }} = useCheckboxes('insightIds', formAPI)

    const filteredInsights = insights
        .filter(insight => insight.title.match(new RegExp(searchInput.input.value, 'gi')))

    return (
        // eslint-disable-next-line react/forbid-elements
        <form ref={ref} onSubmit={handleSubmit}>

            <FormInput
                autoFocus={true}
                title='Filter your insights'
                description={<span className=''>
                    Couldn't find your insight? Check your insight's visibillty settings or <Link to='/insights/create'>create a new insight</Link>
                </span>}
                placeholder='Example: My graphql migrations insight'

                {...searchInput.input}
            />

            <fieldset className={classnames('mt-2', styles.insightsContainer)}>

                { filteredInsights.map(insight =>
                    <label key={insight.id} className={styles.insightItem}>

                        <input
                            type="checkbox"
                            name='insightIds'
                            checked={isChecked(insight.id)}
                            value={insight.id}
                            onChange={onChange}
                            onBlur={onBlur}
                            className='mr-2'/>

                        <TruncatedText>{insight.title}</TruncatedText>
                        <Badge value={insight.owner.name} className={styles.insightOwnerName}/>
                    </label>
                )}
            </fieldset>

            <hr/>

            {formAPI.submitErrors?.[FORM_ERROR] && <ErrorAlert className='mt-3' error={formAPI.submitErrors[FORM_ERROR]} />}

            <div className='d-flex justify-content-end mt-4'>
                <button
                    type='button'
                    className='btn btn-outline-secondary mr-2'
                    onClick={onCancel}>Cancel</button>

                <LoaderButton
                    alwaysShowLabel={true}
                    loading={formAPI.submitting}
                    label={formAPI.submitting ? 'Saving' : 'Save'}
                    type="submit"
                    disabled={formAPI.submitting}
                    spinnerClassName="mr-2"
                    className="btn btn-primary"
                />
            </div>
        </form>
    )
}
