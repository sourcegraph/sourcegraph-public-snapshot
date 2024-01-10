import { useEffect, useState } from 'react'

import Cookies from 'js-cookie'

const setCookie = <T extends string | object | boolean>(
    key: string,
    newValue: T,
    opts: Cookies.CookieAttributes
): void => {
    Cookies.set(key, JSON.stringify(newValue), opts)
}

const removeCookie = (key: string, opts: Cookies.CookieAttributes): void => {
    Cookies.remove(key, opts)
}

const getCookie = <T extends string | object | boolean>(key: string): T | undefined => {
    const cookie = Cookies.get(key)
    if (cookie === undefined) {
        return
    }
    try {
        return JSON.parse(cookie) as T
    } catch (error) {
        // eslint-disable-next-line no-console
        console.error(`useCookieStorage: Error parsing cookie ${key}: ${error}`)
        return
    }
}

/**
 * A React hook to use and set state in Cookies.
 *
 * @param key The Cookie key to use.
 * @param initialValue The initial value to use when there is no value in Cookies for the key.
 * @param opts The options to use when setting the Cookie.
 * @returns A getter and setter for the value (`const [foo, setFoo] = useCookieStorage('key', 123)`, {expires: 365}).
 *
 * @example const [foo, setFoo] = useCookieStorage('key', 123, { expires: 365 })
 */
export const useCookieStorage = <T extends string | object | boolean>(
    key: string,
    initialValue?: T,
    opts: Cookies.CookieAttributes = {}
): [T | undefined, (newValue?: T) => void] => {
    const [value, setValue] = useState<T | undefined>(getCookie(key) ?? initialValue)

    useEffect(() => {
        setValue(value)
        if (value === undefined) {
            removeCookie(key, opts)
            return
        }
        setCookie(key, value, opts)
    }, [key, opts, value])

    return [value, setValue]
}
