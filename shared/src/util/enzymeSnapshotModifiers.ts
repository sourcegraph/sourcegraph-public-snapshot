// This file contains modifier functions to customize JSON serialization of snapshot tests using enzyme-to-json.
// Pass these functions to the `map` field in the `option` object enzyme-to-json's toJson.
// See https://github.com/adriantoine/enzyme-to-json#helper for documentation.

import { Json } from 'enzyme-to-json'

export function replaceHistoryObject(json: Json): Json {
    if (json.props?.history) {
        return {
            ...json,
            props: {
                ...json.props,
                history: '[History]',
            },
        } as Json
    }

    return json
}
