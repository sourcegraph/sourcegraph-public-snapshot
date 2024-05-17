# Cody Pro Client API Library

This module contains the Cody Pro REST API client library and associated types.

These are specific to the API used for managing Cody Pro subscriptions, and its associated
microservice backend. For more information, see the `sourcegraph/self-serve-cody` repo which
implements the server-side of this API.

For compatibility, the data types MUST match the Golang definitions in the `internal/api/types`
package. Care must also be taken on the Golang side to avoid breaking changes, such as not
renaming the JSON serialization of data types, only adding new fields, etc.

For any other Sourcegraph backend interactions from the frontend, that should be using the
GraphQL API.

## Usage

The API client is exposed as two React hooks: `useApiClient` and `useApiCaller`.

The `Client` provides a strongly-typed definition of the REST API exposed as a synchronous methods.
However, they just return `Call<Resp>` objects which merely _describe_ the API call to be made.
A separate `Caller` is what actually performs the operation.

⚠️ It's super important to wrap the `Call<Resp>` object's creation in `useMemo`. Otherwise, any time
the calling React comment gets repainted, the object reference passed to `useApiCaller` will change,
leading to additional HTTP requests being made unintentionally!

```ts
// Make the API call to create the Stripe Checkout session.
// Make the API call to create the Stripe Checkout session.
const call = useMemo(() => Client.createStripeCheckoutSession(req), [req.customerEmail, req.showPromoCodeField])
const { loading, error, data } = useApiCaller(call)
```
