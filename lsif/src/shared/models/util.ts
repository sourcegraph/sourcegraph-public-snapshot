/**
 * Determine the table inserter batch size for an entity given the number of
 * fields inserted for that entity. We cannot perform an insert operation in
 * SQLite with more than 999 placeholder variables, so we need to flush our
 * batch before we reach that amount. If fields are added to the models, the
 * argument to this function also needs to change.
 *
 * @param numFields The number of fields for an entity.
 */
export function calcSqliteBatchSize(numFields: number): number {
    return Math.floor(999 / numFields)
}
