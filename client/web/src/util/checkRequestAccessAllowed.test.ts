import { checkRequestAccessAllowed } from './checkRequestAccessAllowed'

describe('checkRequestAccessAllowed', () => {
    const defaultContext = {
        sourcegraphDotComMode: false,
        allowSignup: false,
    }

    it('should return false if dotcom mode', () => {
        expect(checkRequestAccessAllowed({ ...defaultContext, sourcegraphDotComMode: true })).toBe(false)
    })

    it('should return false if builtin signup enabled', () => {
        expect(checkRequestAccessAllowed({ ...defaultContext, allowSignup: true })).toBe(false)
    })

    it('should return false if auth access request is disabled', () => {
        expect(checkRequestAccessAllowed({ ...defaultContext, authAccessRequest: { disabled: true } })).toBe(false)
    })

    it('should return true if all conditions are met', () => {
        expect(checkRequestAccessAllowed(defaultContext)).toBe(true)
    })

    it('should return true if explicitly set disabled=false', () => {
        expect(checkRequestAccessAllowed({ ...defaultContext, authAccessRequest: { disabled: false } })).toBe(true)
    })
})
