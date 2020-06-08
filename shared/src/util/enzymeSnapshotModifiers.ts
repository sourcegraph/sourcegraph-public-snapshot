// This file contains modifier functions to customize JSON serialization of snapshot tests using enzyme-to-json.
// These are used in share/dev/enzymeSerializer.js

import { Json } from 'enzyme-to-json'

export function replaceVerboseObjects(json: Json): Json {
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
