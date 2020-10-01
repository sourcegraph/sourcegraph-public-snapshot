import React from 'react'
import { OptionFlagWithValue } from '../../shared/util/optionFlags'
import { useInputValidation, deriveInputClassName } from '../../../../shared/src/util/useInputValidation'
import { LoaderInput } from '../../../../shared/src/components/LoaderInput'
import EarthIcon from 'mdi-react/EarthIcon'
import BookOpenPageVariantIcon from 'mdi-react/BookOpenPageVariantIcon'
import { Link } from '../../../../shared/src/components/Link'
import classNames from 'classnames'
import { fromFetch } from 'rxjs/fetch'
import { SourcegraphIcon } from '../../shared/components/SourcegraphIcon'
import { fetchSite } from '../../shared/backend/server'
import { asError } from '../../../../shared/src/util/errors'
import { LinkOrButton } from '../../../../shared/src/components/LinkOrButton'
import { catchError, mapTo } from 'rxjs/operators'
import { Observable } from 'rxjs'
import { GraphQLResult } from '../../../../shared/src/graphql/graphql'
import { useMemo } from '@storybook/addons'

export interface OptionsContainerProps {
    sourcegraphURL: string
    isActivated: boolean
    ensureValidSite: (url: string) => Observable<any>
    fetchCurrentTabStatus: () => Promise<OptionsMenuProps['currentTabStatus']>
    hasPermissions: (url: string) => Promise<boolean>
    requestPermissions: (url: string) => void
    setSourcegraphURL: (url: string) => Promise<void>
    toggleExtensionDisabled: (isActivated: boolean) => Promise<void>
    onChangeOptionFlag: (key: string, value: boolean) => void
    optionFlags: OptionFlagWithValue[]
}

interface OptionsPageProps {
    version: string
    sourcegraphURL: string
    isCurrentRepositoryPrivate: boolean
    requestGraphQL: <T, V = object>(options: {
        request: string
        variables: V
        sourcegraphURL?: string
    }) => Observable<GraphQLResult<T>>
}

export const OptionsPage: React.FunctionComponent<OptionsPageProps> = ({ sourcegraphURL, requestGraphQL }) => {
    const [urlState, nextUrlFieldChange, urlInputReference] = useInputValidation(
        useMemo(
            () => ({
                initialValue: sourcegraphURL,
                asynchronousValidators: [
                    (url: string): Observable<string | undefined> =>
                        fetchSite(options => requestGraphQL({ ...options, sourcegraphURL: url })).pipe(
                            mapTo(undefined),
                            catchError(error => asError(error).message)
                        ),
                ],
            }),
            [sourcegraphURL, requestGraphQL]
        )
    )
    return (
        <div>
            <section>
                <SourcegraphIcon />
                <Toggle />
            </section>
            <section>
                <p>Get code intelligence tootlips while browsing files and reading PRs on your code host.</p>
                {/* Code host icons, with current one highlighted */}
            </section>
            <section>
                <form>
                    <label htmlFor="sourcegraph-url">Sourcegraph URL</label>
                    <LoaderInput loading={urlState.loading} className={classNames(deriveInputClassName(urlState))}>
                        <input
                            id="sourcegraph-url"
                            type="url"
                            pattern="^https://"
                            value={urlState.value}
                            onChange={nextUrlFieldChange}
                            ref={urlInputReference}
                        />
                    </LoaderInput>
                    {urlState.loading ? (
                        <small className="text-muted">Checking...</small>
                    ) : urlState.kind === 'INVALID' ? (
                        <small className="invalid-feedback">{urlState.reason}</small>
                    ) : (
                        <small className="valid-feedback">Looks good!</small>
                    )}
                </form>
            </section>
            <section>
                <LinkOrButton to="https://sourcegraph.com">
                    <EarthIcon className="icon-inline" /> Sourcegraph Cloud
                </LinkOrButton>
                <LinkOrButton to="https://sourcegraph.com">
                    <BookOpenPageVariantIcon className="icon-inline" /> Documentation
                </LinkOrButton>
            </section>
        </div>
    )
}
