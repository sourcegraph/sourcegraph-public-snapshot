import { screen } from '@testing-library/react'
import { describe, expect, test } from 'vitest'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { SearchMode } from '@sourcegraph/shared/src/search'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { Toggles } from './Toggles'

describe('Toggles', () => {
    describe('Query input toggle state', () => {
        test('case toggle for case subexpressions', () => {
            renderWithBrandedContext(
                <Toggles
                    navbarSearchQuery="(case:yes foo) or (case:no bar)"
                    patternType={SearchPatternType.keyword}
                    defaultPatternType={SearchPatternType.keyword}
                    setPatternType={() => undefined}
                    caseSensitive={false}
                    setCaseSensitivity={() => undefined}
                    searchMode={SearchMode.Precise}
                    setSearchMode={() => undefined}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            )

            expect(screen.getAllByRole('checkbox', { name: 'Case sensitivity toggle' })).toMatchSnapshot()
        })

        test('case toggle for patterntype subexpressions', () => {
            renderWithBrandedContext(
                <Toggles
                    navbarSearchQuery="(foo patterntype:literal) or (bar patterntype:structural)"
                    patternType={SearchPatternType.keyword}
                    defaultPatternType={SearchPatternType.keyword}
                    setPatternType={() => undefined}
                    caseSensitive={false}
                    setCaseSensitivity={() => undefined}
                    searchMode={SearchMode.Precise}
                    setSearchMode={() => undefined}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            )
            expect(screen.getAllByRole('checkbox', { name: 'Case sensitivity toggle' })).toMatchSnapshot()
        })

        test('regexp toggle for patterntype subexpressions', () => {
            renderWithBrandedContext(
                <Toggles
                    navbarSearchQuery="(foo patterntype:literal) or (bar patterntype:structural)"
                    patternType={SearchPatternType.keyword}
                    defaultPatternType={SearchPatternType.keyword}
                    setPatternType={() => undefined}
                    caseSensitive={false}
                    setCaseSensitivity={() => undefined}
                    searchMode={SearchMode.Precise}
                    setSearchMode={() => undefined}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            )
            expect(screen.getAllByRole('checkbox', { name: 'Regular expression toggle' })).toMatchSnapshot()
        })

        test('regexp toggle with default patterntype', () => {
            renderWithBrandedContext(
                <Toggles
                    navbarSearchQuery="foo.*bar"
                    patternType={SearchPatternType.keyword}
                    defaultPatternType={SearchPatternType.standard}
                    setPatternType={() => undefined}
                    caseSensitive={false}
                    setCaseSensitivity={() => undefined}
                    searchMode={SearchMode.Precise}
                    setSearchMode={() => undefined}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            )
            expect(screen.getAllByRole('checkbox', { name: 'Regular expression toggle' })).toMatchSnapshot()
        })
    })
})
