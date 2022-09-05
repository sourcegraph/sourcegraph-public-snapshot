import { Form } from '@sourcegraph/branded/src/components/Form'
import { Alert, Button, Input, Link, LoadingSpinner, Text } from '@sourcegraph/wildcard'
import React from 'react'

export const NewInstance: React.FunctionComponent<{}> = () => (
    <div>
        New instance flow
        <Form>
            <Input
                // onChange={this.onEmailFieldChange}
                // value={this.state.email}
                type="email"
                name="email"
                autoFocus={true}
                spellCheck={false}
                required={true}
                autoComplete="email"
                // disabled={this.state.submitOrError === 'loading'}
                className="form-group"
                label={<Text className="text-left">Email</Text>}
            />
            <Button
                className="mt-4"
                type="submit"
                // disabled={this.state.submitOrError === 'loading'}
                variant="primary"
                display="block"
            >
                {/* this.state.submitOrError === 'loading' ? <LoadingSpinner /> : 'Send reset password link' */}
                Sign up
            </Button>
        </Form>
    </div>
)
