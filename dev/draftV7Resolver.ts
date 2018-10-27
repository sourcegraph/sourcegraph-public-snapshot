import fs from 'fs';
import { ResolverOptions } from 'json-schema-ref-parser';

/**
 * If the host is offline, json-schema-ref-parser will fail to fetch
 * the v7 draft of json-schema, causing program startup to fail
 */
export const draftV7resolver: ResolverOptions = {
    order: 1,
    read: () =>
        fs.readFileSync(__dirname + '/json-schema-v7.json'),
    canRead: file =>
        file.url === 'http://json-schema.org/draft-07/schema'
}
