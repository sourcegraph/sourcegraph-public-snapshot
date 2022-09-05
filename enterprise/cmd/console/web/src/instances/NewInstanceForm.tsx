import { Form } from '@sourcegraph/branded/src/components/Form'
import { Button, Input, Text } from '@sourcegraph/wildcard'
import React from 'react'
import styles from './NewInstanceForm.module.scss'
import classNames from 'classnames'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'

export const NewInstanceForm: React.FunctionComponent<{}> = () => {
    const domainSuggestion = 'acme-corp'

    return (
        <Form className={styles.form}>
            <Input
                // onChange={this.onEmailFieldChange}
                // value={this.state.email}
                type="text"
                name="domain"
                autoFocus={true}
                spellCheck={false}
                required={true}
                className="form-group"
                placeholder={domainSuggestion}
                defaultValue={domainSuggestion}
                size={24}
                inputClassName={styles.domainInput}
                // disabled={this.state.submitOrError === 'loading'}
                label={<Text className="text-left mb-1">Workspace domain</Text>}
                inputSymbol={
                    <span className={classNames(styles.checkIcon, 'ml-1')}>
                        <CheckCircleIcon className={classNames(styles.checkIcon, 'ml-1', 'text-success')} />
                        <span className="text-muted">.sourcegraph.com</span>
                    </span>
                }
            />
            <Button
                className="mt-2"
                type="submit"
                // disabled={this.state.submitOrError === 'loading'}
                variant="primary"
                display="block"
                onClick={() => {
                    window.location.pathname = '/instances'
                }}
            >
                {/* this.state.submitOrError === 'loading' ? <LoadingSpinner /> : 'Send reset password link' */}
                Create instance
            </Button>
        </Form>
    )
}
