export interface BatchChangesLicenseInfo {
    unrestricted: boolean
    maxNumChangesets: number
}

const BATCH_CHANGES_LIMITED_LICENSE: BatchChangesLicenseInfo = {
    maxNumChangesets: 10,
    unrestricted: false,
}

const BATCH_CHANGES_FULL_LICENSE: BatchChangesLicenseInfo = {
    maxNumChangesets: -1,
    unrestricted: true,
}

/**
 *
 * @param type a small subset of the batch changes license types
 * mocked here to change the window context license info for batches UI gating
 *
 * returns void
 */
export const updateJSContextBatchChangesLicense = (type: 'none' | 'limited' | 'full'): void => {
    const license =
        type === 'full' ? BATCH_CHANGES_FULL_LICENSE : type === 'limited' ? BATCH_CHANGES_LIMITED_LICENSE : undefined

    // This seems necessary to prevent Typescript from associating `window.context` with
    // Mocha.SuiteFunction and subsequently throwing an error that the property `licenseInfo`
    // doesn't exist.
    const context: any = window.context

    context.licenseInfo = context.licenseInfo
        ? {
              ...context.licenseInfo,
              batchChanges: license,
          }
        : {
              currentPlan: 'team-0',
              batchChanges: license,
          }
}
