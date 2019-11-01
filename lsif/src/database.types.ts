import { Id } from 'lsif-protocol'

export type DocumentId = Id
export type DocumentPath = string
export type RangeId = Id
export type DefinitionResultId = Id
export type ReferenceResultId = Id
export type DefinitionReferenceResultId = DefinitionResultId | ReferenceResultId
export type HoverResultId = Id
export type MonikerId = Id
export type PackageInformationId = Id

/**
 * A type that describes a gzipped and JSON-encoded value of type `T`.
 */
export type JSONEncoded<T> = Buffer

/**
 * A type of hashed value created by hashing a value of type `T` and performing
 * the modulus with a value of type `U`. This is to link the index of a result
 * chunk to the hashed value of the identifiers stored within it.
 */
export type HashMod<T, U> = number
