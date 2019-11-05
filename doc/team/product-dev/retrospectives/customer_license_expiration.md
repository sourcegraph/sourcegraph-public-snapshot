# Customer license expiration P0 retrospective

## Incident details

- Date: Feb 21, 2019
- Customer: https://app.hubspot.com/contacts/2762526/company/419771425
- Discussion: https://sourcegraph.slack.com/archives/G9EN3TJDD/p1553176601007400

## Why did this happen?

The customer's contract originally provided them with an opt-out clause after 1 year. That 1 year mark fell on Mar 21, 2019.

I (Dan) emailed customer's A/P department on Jan 17, 2019 indicating to them that the second invoice was about to come due. They asked some questions, indicated it had been approved, and the wire went out. This occurred much faster than I expected, as compared to the first year's 60-day payment terms (the original email had been sent so the payment date would fall before the license key expiration intentionally).

After the payment was received, I didn't think to go to Sourcegraph.com to generate a new license key — I simply overlooked the second step. The deadline was still 1.5 months away, and the process I had gone through was purely financial/legal, not technical (e.g., I didn't recall that the original key was only for 1 year, rather than for the full 3 years).

Fundamentally, the process is entirely manual today. It's up to the member of the GTM team who negotiates the contract to remember to send updated license keys out.

## Action items

What are we going to change to make sure customers get the correct license keys at the correct time?

A few things. These changes all have the goal of making the process less manual/artisanal/custom.

1. Dan will add code that creates a dismissable site alert for admins starting 7 days out from license expiration.
1. Dan will add a page (or add functionality to an existing page) that shows licenses in ascending order by expiration date on Sourcegraph.com, so we can see which ones will be expiring soon.
1. To prevent fixes to similar issues in the future from becoming bottlenecked on GTM team member availability, Engineers will be permanently formally approved to create licenses for unlimited numbers of users for up to 7 days.
1. Dan will add information on dev team license key creation permissions, along with steps for creating the license key, to our development docs.
1. Dan will audit all current license keys and remove tests and expired keys
