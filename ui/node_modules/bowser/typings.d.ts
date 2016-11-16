export const name: string
export const osversion: string
export const version: string

// grades
export const a: boolean
export const b: boolean
export const c: boolean

// engines
export const android: boolean
export const bada: boolean
export const blackberry: boolean
export const chrome: boolean
export const firefox: boolean
export const gecko: boolean
export const ios: boolean
export const msie: boolean
export const msedge: boolean
export const opera: boolean
export const phantom: boolean
export const safari: boolean
export const sailfish: boolean
export const seamonkey: boolean
export const silk: boolean
export const tizen: boolean
export const webkit: boolean
export const webos: boolean
export const mobile: boolean
export const tablet: boolean

// operating systems
export const chromeos: boolean
export const iphone: boolean
export const ipad: boolean
export const ipod: boolean
export const firefoxos: boolean
export const linux: boolean
export const mac: boolean
export const touchpad: boolean
export const windows: boolean
export const windowsphone: boolean

export function test(browserList: Flag[]): boolean
export function isUnsupportedBrowser(minVersions:Object, strictMode?:Boolean, ua?:string): boolean

export type Flag = "a" | "b" | "c" | "android" | "bada" | "blackberry"
                 | "chrome" | "firefox" | "gecko" | "ios" | "msie"
                 | "msedge" | "opera" | "phantom" | "safari"
                 | "sailfish" | "seamonkey" | "silk" | "tizen"
                 | "webkit" | "webos" | "chromeos" | "iphone"
                 | "ipad" | "ipod" | "firefoxos" | "linux" | "mac"
                 | "touchpad" | "windows" | "windowsphone"

