import { Form } from '@sourcegraph/branded/src/components/Form'
import { Button, Input, Text } from '@sourcegraph/wildcard'
import React from 'react'
import styles from './SignupPage.module.scss'
import classNames from 'classnames'
import { Link } from 'react-router-dom'
import { GoogleIcon } from '../../../GoogleIcon'
import { TrialStartFlowContainer } from '../../TrialStartFlowContainer'
import ArrowRightThickIcon from 'mdi-react/ArrowRightThickIcon'

export const SignupPage: React.FunctionComponent<{}> = () => (
    <TrialStartFlowContainer>
        <h2 className="mt-2">Start using Sourcegraph</h2>
        <p className="text-muted mb-4">
            If your organization is already using Sourcegraph, you can <Link to="TODO">sign in</Link>.
        </p>
        <Form>
            <div className={styles.fullNameRow}>
                <Input
                    // onChange={this.onEmailFieldChange}
                    // value={this.state.email}
                    type="text"
                    name="firstName"
                    autoFocus={true}
                    spellCheck={false}
                    required={true}
                    autoComplete="given-name"
                    className="form-group"
                    placeholder="First name"
                    // disabled={this.state.submitOrError === 'loading'}
                    label={<Text className="text-left mb-1">First name</Text>}
                />
                <Input
                    // onChange={this.onEmailFieldChange}
                    // value={this.state.email}
                    type="text"
                    name="lastName"
                    spellCheck={false}
                    required={true}
                    autoComplete="family-name"
                    className="form-group"
                    placeholder="Last name"
                    // disabled={this.state.submitOrError === 'loading'}
                    label={<Text className="text-left mb-1">Last name</Text>}
                />
            </div>
            <Input
                // onChange={this.onEmailFieldChange}
                // value={this.state.email}
                type="email"
                name="email"
                spellCheck={false}
                required={true}
                autoComplete="email"
                // disabled={this.state.submitOrError === 'loading'}
                className="form-group"
                placeholder="you@company.com"
                label={<Text className="text-left mb-1">Work email</Text>}
            />
            <Input
                // onChange={this.onEmailFieldChange}
                // value={this.state.email}
                type="password"
                name="password"
                spellCheck={false}
                required={true}
                autoComplete="new-password"
                className="form-group"
                placeholder="Password"
                // disabled={this.state.submitOrError === 'loading'}
                label={<Text className="text-left mb-1">Password</Text>}
            />
            <Button
                className="mt-2"
                type="submit"
                // disabled={this.state.submitOrError === 'loading'}
                variant="primary"
                display="block"
                onClick={() => {
                    localStorage.setItem('signedIn', 'true')
                    window.location.pathname = '/new-instance'
                }}
            >
                {/* this.state.submitOrError === 'loading' ? <LoadingSpinner /> : 'Send reset password link' */}
                Sign up <ArrowRightThickIcon className="icon-inline" />
            </Button>
        </Form>
        <Button
            className="mt-3"
            type="submit"
            // disabled={this.state.submitOrError === 'loading'}
            variant="secondary"
            display="block"
        >
            <GoogleIcon className={classNames(styles.googleIcon, 'icon-inline', 'mr-1')} /> Sign up with Google
        </Button>
        <p className="text-muted small mt-3 mb-0 text-center">
            By using Sourcegraph, you agree to our <a href="#">privacy policy</a> and <a href="#">terms</a>.
        </p>
    </TrialStartFlowContainer>
)
