import { vsCodeMocks } from './mocks'

/**
 * Mock name is required to keep Jest happy and avoid the error:
 * "The module factory of jest.mock() is not allowed to reference any out-of-scope variables"
 *
 * This function can be used to customize the default VSCode mocks in any test file.
 */
export function mockVSCodeExports(): typeof vsCodeMocks {
    return vsCodeMocks
}

/**
 * Apply the default VSCode mocks to the global scope.
 */
jest.mock('vscode', () => mockVSCodeExports())
