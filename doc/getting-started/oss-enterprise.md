# Sourcegraph Enterprise vs. Sourcegraph OSS

Sourcegraph offers two versions of its product: Sourcegraph Open Source (Sourcegraph OSS) and Sourcegraph Enterprise. 

Sourcegraph Enterprise is Sourcegraph’s primary offering and includes all code intelligence platform features, like:

- Code Search
- Code navigation
- Batch Changes
- Code Insights

See the [detailed feature comparison](#detailed-feature-comparison) for a list of all features.

Sourcegraph Enterprise is the best solution for enterprises who want to use Sourcegraph with their organization’s code. Sourcegraph Enterprise is available via our [paid plans](https://about.sourcegraph.com/pricing). Teams of up to 10 developers can use a limited version of Sourcegraph Enterprise for free.

Sourcegraph OSS only includes universal code search functionality and does not include any [code intelligence platform](https://about.sourcegraph.com/blog/code-search-to-code-intelligence) or enterprise features. Sourcegraph OSS does **not** include:

- Code navigation
- Batch Changes
- Code Insights
- Code monitoring
- Notebooks
- SSO
- Browser/IDE extensions
- And [more](#detailed-feature-comparison) 

Sourcegraph OSS is built as a single container which limits scalability. 

Check out our [handbook](https://handbook.sourcegraph.com/departments/engineering/product/process/gtm/licensing/) for additional details on Sourcegraph OSS. 

## Getting started
Sourcegraph Enterprise can be run in a variety of environments, from cloud to self-hosted to your local machine. For most customers, we recommend [Sourcegraph Cloud](https://signup.sourcegraph.com/), managed entirely by Sourcegraph. Visit the [get started](https://docs.sourcegraph.com/#get-started) section of the docs for details on every deployment option.

Developers can use [Sourcegraph OSS](https://github.com/sourcegraph/sourcegraph) without agreeing to any enterprise licensing terms by building their own server image. If they do so, no code from our enterprise-licensed features will be included in their Sourcegraph deployment.

In practice, Sourcegraph OSS involves:

- Removing the enterprise directories from the repository
- Building your own docker image (you can’t just use ours)

## Detailed feature comparison

<table>
  <tr>
   <td><strong>License</strong>
   </td>
   <td><strong>Sourcegraph Open Source (Sourcegraph OSS)</strong>
   </td>
   <td colspan="3" ><strong>Sourcegraph Enterprise</strong>
   </td>
  </tr>
  <tr>
   <td><strong>Tier</strong>
   </td>
   <td>NA
   </td>
   <td>Free
   </td>
   <td>Business
   </td>
   <td>Enterprise
   </td>
  </tr>
  <tr>
   <td><strong>Price</strong>
   </td>
   <td>Free
   </td>
   <td>Free for up to 10 users
   </td>
   <td>$99 per active user/month
   </td>
   <td>Custom pricing
   </td>
  </tr>
  <tr>
   <td colspan="5" ><strong>Code intelligence platform</strong>
   </td>
  </tr>
  <tr>
   <td>Code Search
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>Code navigation (go to definition/find references)
   </td>
   <td>✗
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>Batch Changes
   </td>
   <td>✗
   </td>
   <td>10 changesets
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>Code Insights
   </td>
   <td>✗
   </td>
   <td>2 insights
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>Notebooks
   </td>
   <td>✗
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>Code monitoring
   </td>
   <td>✗
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>Comprehensive API
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td colspan="5" ><strong>Code host integrations</strong>
   </td>
  </tr>
  <tr>
   <td># of code host integrations
   </td>
   <td>Unlimited
   </td>
   <td>1
   </td>
   <td>Unlimited
   </td>
   <td>Unlimited
   </td>
  </tr>
  <tr>
   <td colspan="5" ><strong>Security, compliance, and admin</strong>
   </td>
  </tr>
  <tr>
   <td>SOC 2 Type II 
   </td>
   <td>✗
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>In-product analytics
   </td>
   <td>✗
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>User and admin roles
   </td>
   <td>✓
   </td>
   <td>✗
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>SSO/SAML
   </td>
   <td>✗
   </td>
   <td>✗
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>Standard repository permissions
   </td>
   <td>✗
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>Custom repository permissions API
   </td>
   <td>✗
   </td>
   <td>✗
   </td>
   <td>✗
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>Private instance access
   </td>
   <td>✗
   </td>
   <td>✗
   </td>
   <td>✗
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td colspan="5" ><strong>Scale and performance</strong>
   </td>
  </tr>
  <tr>
   <td>Cloud code storage
   </td>
   <td>✗
   </td>
   <td>✗
   </td>
   <td>Up to 75GB
   </td>
   <td>Over 75GB
   </td>
  </tr>
  <tr>
   <td>Executors
   </td>
   <td>✗
   </td>
   <td>✗
   </td>
   <td>2 
   </td>
   <td>4
   </td>
  </tr>
  <tr>
   <td colspan="5" ><strong>Support</strong>
   </td>
  </tr>
  <tr>
   <td>Support
   </td>
   <td>Community
   </td>
   <td>Community
   </td>
   <td>24/5 support
   </td>
   <td>24/5 support
   </td>
  </tr>
  <tr>
   <td>Technical Account Manager
   </td>
   <td>✗
   </td>
   <td>✗
   </td>
   <td>Available
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>Support SLA
   </td>
   <td>✗
   </td>
   <td>✗
   </td>
   <td>Standard
   </td>
   <td>Priority
   </td>
  </tr>
  <tr>
   <td colspan="5" ><strong>Deployment </strong>
   </td>
  </tr>
  <tr>
   <td>Single-tenant cloud deployment
   </td>
   <td>✗
   </td>
   <td>✗
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>Self-hosted deployment
   </td>
   <td>✓
   </td>
   <td>✓
   </td>
   <td>✗
   </td>
   <td>✓
   </td>
  </tr>
  <tr>
   <td>Air-gapped deployment
   </td>
   <td>✗
   </td>
   <td>✗
   </td>
   <td>✗
   </td>
   <td>Add-on
   </td>
  </tr>
  <tr>
   <td colspan="5" ><strong>Usage and billing</strong>
   </td>
  </tr>
  <tr>
   <td>Price
   </td>
   <td>Free
   </td>
   <td>Free
   </td>
   <td>$99 per active user/month
<p>
<em>Platform access fee may apply</em>
   </td>
   <td>Custom pricing
<p>
<em>Platform access fee may apply</em>
   </td>
  </tr>
  <tr>
   <td>Contract length
   </td>
   <td>NA
   </td>
   <td>NA
   </td>
   <td>Annual
   </td>
   <td>Annual
   </td>
  </tr>
  <tr>
   <td>Payment method
   </td>
   <td>NA
   </td>
   <td>NA
   </td>
   <td>Invoice
   </td>
   <td>Invoice
   </td>
  </tr>
</table>
