import type { FC, ChangeEvent, FocusEventHandler } from 'react'

import { SourcegraphLogo } from '@sourcegraph/branded/src/components/SourcegraphLogo'
import { H1, H2, Label, Form, useForm, useCheckboxes } from '@sourcegraph/wildcard'

import { LoaderButton } from '../components/LoaderButton'

import styles from './PostSignInSubscription.module.scss'

export const PostSignInSubscription: FC = props => {
    const { formAPI, ref, handleSubmit } = useForm({
        initialValues: { subscriptions: [] },
        // eslint-disable-next-line no-console
        onSubmit: values => console.log(values),
    })

    const {
        input: { isChecked, onBlur, onChange },
    } = useCheckboxes('subscriptions', formAPI)

    return (
        <div className={styles.root}>
            <header className={styles.header}>
                <SourcegraphLogo className={styles.logo} />
            </header>

            <section className={styles.content}>
                <H1>Set your communication preferences</H1>

                <Form ref={ref} className={styles.form} onSubmit={handleSubmit}>
                    <SubscriptionOption
                        value="product-updates"
                        title="Product updates"
                        message="Stay in the know on the latest awesome features"
                        isChecked={isChecked}
                        onBlur={onBlur}
                        onChange={onChange}
                    />

                    <SubscriptionOption
                        value="tutorials"
                        title="Getting started tutorials"
                        message="Learn the nuts and bolts to become proficient in Sourcegraph tools"
                        isChecked={isChecked}
                        onBlur={onBlur}
                        onChange={onChange}
                    />

                    <SubscriptionOption
                        value="security-updates"
                        title="Security updates"
                        message="Stay informed and help keep your environment secure"
                        isChecked={isChecked}
                        onBlur={onBlur}
                        onChange={onChange}
                    />

                    <H2 className={styles.subHeading}>Help improve Sourcegraph</H2>

                    <SubscriptionOption
                        value="research-program"
                        title="Join our user research program"
                        message="Help us improve Sourcegraph for you and everyone!"
                        isChecked={isChecked}
                        onBlur={onBlur}
                        onChange={onChange}
                    />

                    <footer className={styles.footer}>
                        <LoaderButton
                            type="submit"
                            alwaysShowLabel={true}
                            data-testid="insight-save-button"
                            loading={formAPI.submitting}
                            label={formAPI.submitting ? 'Submitting' : 'Next'}
                            disabled={formAPI.submitting}
                            variant="primary"
                            className={styles.submit}
                        />
                    </footer>
                </Form>
            </section>
        </div>
    )
}

interface SubscriptionOptionProps {
    value: string
    title: string
    message: string
    isChecked: (value: string) => boolean
    onChange?: (event: ChangeEvent<HTMLInputElement>) => void
    onBlur?: FocusEventHandler<HTMLInputElement>
}

const SubscriptionOption: FC<SubscriptionOptionProps> = props => {
    const { value, title, message, isChecked, onChange, onBlur } = props

    return (
        <Label className={styles.option}>
            {/* eslint-disable-next-line react/forbid-elements */}
            <input
                type="checkbox"
                value={value}
                checked={isChecked(value)}
                className={styles.optionCheckbox}
                onBlur={onBlur}
                onChange={onChange}
            />

            <span className={styles.optionTitle}>{title}</span>
            <span className={styles.optionMessage}>{message}</span>
        </Label>
    )
}
