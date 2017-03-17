// The stripe package handles user billing and payments. It relies on Auth0 to
// handle user information, including the Stripe customer ID. This allows us to
// avoid duplicating customer information in a separate, Sourcegraph hosted
// database.
package stripe
