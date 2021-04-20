import classnames from 'classnames';
import React, { PropsWithChildren, ReactElement } from 'react';
import { Form } from 'reactstrap';

import { Page } from '../../../components/Page';
import { PageTitle } from '../../../components/PageTitle';

import './CreateinsightPage.scss'

interface InputFieldProps {
    name: string;
    description?: string;
    placeholder?: string;
    className?: string
}

function InputField(props: InputFieldProps): ReactElement {
    const { name, placeholder, description, className } = props;

    return (
        <label className={classnames('form-field', className)}>
            <h4 className='form-field__name'>{name}</h4>

            <input
                type="text"
                className="form-control form-field__input"
                placeholder={placeholder}
            />

            <span className='form-field__description text-muted'>
                {description}
            </span>
        </label>
    );
}

interface FormGroupProps {
    className?: string;
    name: string;
    description?: string;
}

function FormGroup(props: PropsWithChildren<FormGroupProps>): ReactElement {
    const { className, name, children, description } = props;

    return (
        <fieldset className={classnames('form-group', className)}>

            <h4 className='form-group__name'>{name}</h4>

            <div className='form-group__content'>
                {children}
            </div>

            { description && <span className='form-group__description text-muted'>{description}</span> }
        </fieldset>
    )
}

interface RadioInputProps {
    name: string;
    description?: string;
    className?: string;
}

function RadioInput(props: RadioInputProps): ReactElement {
    const { name, description, className } = props;

    return (
        <label className={classnames('radio-input', className)}>
            <input
                type="radio"
                className="form-control radio-input__input"
                required={true}
            />

            <div className='radio-input__description-content'>
                <span className='radio-input__name'>{name}</span>
                { description && <span className='radio-input__description text-muted'> – {description}</span>}
            </div>
        </label>
    );
}

interface FormColorPickerProps {
    className?: string;
    colours?: { color: string, name?: string, active?: boolean }[];
}

const DEFAULT_COLOURS = [
    { color: 'var(--oc-red-7)', name: 'red color'},
    { color: 'var(--oc-pink-7)' },
    { color: 'var(--oc-grape-7)', active: true, },
    { color: 'var(--oc-violet-7)' },
    { color: 'var(--oc-indigo-7)' },
    { color: 'var(--oc-blue-7)' },
    { color: 'var(--oc-cyan-7)' },
    { color: 'var(--oc-teal-7)' },
    { color: 'var(--oc-green-7)' },
    { color: 'var(--oc-lime-7)' },
    { color: 'var(--oc-yellow-7)' },
    { color: 'var(--oc-orange-7)' },
];

function FormColorPicker(props: FormColorPickerProps):ReactElement {
    const { className, colours = DEFAULT_COLOURS } = props;

    return (
        <div className={classnames('form-color-picker', className)}>

            <div className='form-color-picker__colour-content'>

                { colours.map(colorInfo =>
                    <div
                        key={colorInfo.color}
                        className={classnames(
                            'form-color-picker__color-block',
                            { 'form-color-picker__color-block--active': colorInfo.active})
                        }
                        /* eslint-disable-next-line react/forbid-dom-props */
                        style={{ color: colorInfo.color }}
                        title={colorInfo.name}/>
                )}
            </div>

            <div>or <span className='form-color-picker__custom-color'>use custom color</span></div>
        </div>
    )
}

interface CreateInsightPageProps {}

export const CreateInsightPage: React.FunctionComponent<CreateInsightPageProps> = () => (
        <Page className='create-insight col-8'>
            <PageTitle title='Create new code insight'/>

            <div className='create-insight__sub-title-container'>

                <h2 className='create-insight__sub-title'>Create new code insight</h2>

                <p className='text-muted'>
                    Search-based code insights analyse your code based on any search query.
                    {' '}
                    <a href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points"
                       target="_blank"
                       rel="noopener"
                       className='create-insight__doc-link'>Learn more.</a>
                </p>
            </div>

            <Form className='create-insight__form'>

                <InputField
                    name='Name'
                    description='Chose a unique for your insights'
                    placeholder='Enter the unique name for your insight'
                    className='create-insight__form-field'/>

                <InputField
                    name='Title'
                    description='Shown as title for your insight'
                    placeholder='ex. Migration to React function components'
                    className='create-insight__form-field'/>

                <InputField
                    name='Repositories'
                    description='Create a list of repositories to run your search over. Separate them with comas.'
                    placeholder='Add or search for repositories'
                    className='create-insight__form-field'/>

                <FormGroup
                    name='Visibility'
                    description='This insigh will be visible only on your personal dashboard. It will not be show to other
                        users in your organisation.'
                    className='create-insight__form-field'>

                    <div className='create-insight__radio-group-content'>

                        <RadioInput
                            name='Personal'
                            description='only for you'
                            className='create-insight__radio'/>
                        <RadioInput
                            name='Organization'
                            description='to all users in your organization'
                            className='create-insight__radio'/>
                    </div>
                </FormGroup>

                <hr className='create-insight__separator'/>

                <FormGroup name='Data series'>
                    <div className='create-insight__series-content'>

                        <InputField
                            name='Name'
                            placeholder='ex. Function component'
                            description='Name shown in the legend and tooltip'
                            className='create-insight__form-field--series'/>

                        <InputField
                            name='Query'
                            placeholder='ex. spatternType:regexp const\\s\\w+:\\s(React\\.)?FunctionComponent'
                            description='Do not include the repo: filter as it will be added automatically for the current repository'
                            className='create-insight__form-field--series'/>

                        <FormGroup
                            name='Color'
                            className='create-insight__form-field--series'>

                            <FormColorPicker/>
                        </FormGroup>

                        <button
                            type='button'
                            className='button create-insight__series-button'>

                            Done
                        </button>
                    </div>
                </FormGroup>


                <hr className='create-insight__separator'/>

                <FormGroup
                    name='Step between data points'
                    description='The distance between two data points on the chart'
                    className='create-insight__form-field'>

                    <div className='create-insight__radio-group-content'>

                        <input
                            type="text"
                            placeholder='ex. 2'
                            className="form-control create-insight__step-input form-field__input"
                        />

                        <RadioInput
                            name='Hours'
                            className='create-insight__radio'/>
                        <RadioInput
                            name='Days'
                            className='create-insight__radio'/>
                        <RadioInput
                            name='Weeks'
                            className='create-insight__radio'/>
                        <RadioInput
                            name='Months'
                            className='create-insight__radio'/>
                        <RadioInput
                            name='Years'
                            className='create-insight__radio'/>
                    </div>
                </FormGroup>

                <hr className='create-insight__separator'/>

                <div>
                    <button
                        type='button'
                        className='button create-insight__button create-insight__button--active'>

                        Create code insight
                    </button>
                    <button
                        type='button'
                        className='button create-insight__button'>

                        Cancel
                    </button>
                </div>
            </Form>
        </Page>
    )
