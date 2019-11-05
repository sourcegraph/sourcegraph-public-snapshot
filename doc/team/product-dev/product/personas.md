# Personas

> The persona is an archetype description of an imaginary but very plausible user that personifies these traits – especially their behaviors, attitudes, and goals.
>
> Personas help you prioritize what’s important. If you have decided to make “Mary” the target for this release, then if this feature is critical for “Mary” then put it in, if it’s for “Sam” then it’s out. As you can see, just as important as deciding who a release is for, is deciding who it is not for. It is an extremely common mistake for a product to try to please everyone and end up pleasing no one. This process can help prevent that.

Source: "[Personas for Product Management](https://svpg.com/personas-for-product-management/)"

### Questions:

- What are these personas designed to define? Decision makers? Evangelists that bring the product to their teams? The "purchasers" or decision makers?
- Do we want personas to define HOW the users are interacting with the product? Which features they use, what they care about?
- Should we name them - it is common to give them personalities? E.g. Dave the DevOps Engineer
- Other things to consider:
    - Influence
    - Product knowledge
    - Likes/dislikes
    - Goals
    - Skills
    - Trusts information from
- Do we have a record of interviews with users?

## Infrastructure/Platform Engineer

#### Who

An individual contributor on a team, where the team is responsible for:

- builds/manages tooling,
- sets coding standards,
- creates libraries, or
- otherwise improves the developer experience

for other developers.

#### Role

- Senior Software Engineer
- Tech Lead
- Infrastructure Engineer

(Company: 100+ developers)

#### Problems

- I personally feel we need code search for my company's code
- I personally want code search
- "I want what I had at Google/Facebook"

#### What success looks like

- Sourcegraph is deployed and proven working on their code and at their scale
- Their company decides to adopt Sourcegraph and many developers use it

#### What failure looks like

- Sourcegraph does not work and wastes their time setting it up
- They fail to get others to use Sourcegraph
- They get in trouble for setting up a paid product or one that legal/security hasn't vetted

#### Common objections

- I don't know why I need code search (for people who don't have the problems stated above)
- I and/or my team needs code search, but the rest of my organization doesn't need it
- I don't know how to spread it to the rest of my company
- The price is too high for my budget

#### Personally rewarded by

- Writing clean code
- Enabling developers to answer their own questions
- Enabling others to build and use APIs the right way

#### Examples

- https://app.hubspot.com/contacts/2762526/contact/1909869/
- https://app.hubspot.com/contacts/2762526/contact/16012101/
- https://app.hubspot.com/contacts/2762526/contact/15757651/
- https://app.hubspot.com/contacts/2762526/contact/15755551/
- https://app.hubspot.com/contacts/2762526/contact/10414951/
- https://app.hubspot.com/contacts/2762526/contact/11545601/
- https://app.hubspot.com/contacts/2762526/contact/13456355/
- https://app.hubspot.com/contacts/2762526/contact/15873101/

## DevOps/Production Engineering Engineer

#### Who

An individual contributor who is responsible for coding and scripting on processes related to build, test, packaging, deployment, monitoring, capacity planning, and observability.

#### Role

- DevOps team
- Production Engineering team (this new term is preferred by some companies)

#### Problems

- I need to stay aware of how all of our systems are deployed
- I need to help developers understand how to deploy and maintain their applications
- SREs and developers expect me to give them a way to respond to incidents

#### What success looks like

- The software delivery pipeline (build, test, package, deploy, monitor, etc.) is healthy and widely used by all of our applications' codebases
- Sourcegraph helps developers and DevOps engineers work together better
- Sourcegraph is where my DevOps team goes to understand how an application is deployed

#### What failure looks like

Nobody else at my company is using Sourcegraph, so it's not worth the maintenance burden

#### Other tools used

- CI
- Artifactory
- Datadog/LightStep/etc.
- Prometheus and Grafana

#### Common objections

- TBD

#### Personally rewarded by

Teaching best practices about deployment to engineers

#### Examples

- https://app.hubspot.com/contacts/2762526/contact/16035251/
- https://app.hubspot.com/contacts/2762526/contact/15851951/
- https://app.hubspot.com/contacts/2762526/contact/13455899/

## Site Reliability Engineer (SRE)

#### Who

The person who gets a page when the site goes down (due to a recent application change) and needs to coordinate incident response to restore the site.

#### Role

Site Reliability Engineer (SRE)

#### Problems

- When the site goes down, I need to quickly find the source of the problem
- I need to reduce the likelihood that developers build systems that will fail in production

#### What success looks like

- Reduced incident response times
- I am able to be proactive (finding defects before they take down prod), not reactive
- My company's applications have higher uptime/stability and fewer incidents
- Developers tap the SRE team's knowledge more frequently and at the right times

#### What failure looks like

Nobody else at my company is using Sourcegraph, so it's not worth the maintenance burden

#### Common objections

The defects that cause downtime for us are not related to code changes, so Sourcegraph would not help

#### Personally rewarded by

Teaching best practices about building reliable systems to engineers

#### Examples

- https://app.hubspot.com/contacts/2762526/contact/13456445/
- https://app.hubspot.com/contacts/2762526/contact/15613852/

## Engineering and DevOps Managers

#### Who

Middle manager with enough sway to get Sourcegraph widely used, but not enough to sway to allocate budget.

#### Role

- Engineering Manager (not specifically for developer infrastructure, tooling, etc.)
- Engineering Lead
- Director of DevOps (or Director of Production Engineering)
- Directory of Delivery

#### Problems

- I want to increase the velocity/quality of my team's development and code review processes
- My team needs to better cater to the internal consumers of our service's API
- I want to help my team become more independent and uplevel their skills by discovering best practices in code

#### What success looks like

- My team's velocity and/or quality improves
- My team's code review culture improves
- My team's engineers are able to learn/do more on their own without needing my help

#### What failure looks like

- My team's engineers don't end up actually using Sourcegraph
- I can't see/demonstrate the value of Sourcegraph to justify continuing usage of it
- No other teams start using Sourcegraph, so it's not worth the maintenance/education burden

#### Common objections

Sourcegraph doesn't address a burning need of mine

#### Personally rewarded by

Helping my team's engineers improve coding, planning, and communication skills

#### Examples

- https://app.hubspot.com/contacts/2762526/contact/13520551/
- https://app.hubspot.com/contacts/2762526/contact/15877401/
- https://app.hubspot.com/contacts/2762526/contact/15613852/
- https://app.hubspot.com/contacts/2762526/contact/16081201/

## IT Engineer/Manager at 250+-Engineer Companies

#### Who

System administrator (or manager) on IT or internal tools team that manages internal tools from 3rd-party vendors.

#### Role

- IT Engineer
- IT Manager
- Systems Engineer
- System Administrator

(Company: 250+ engineers)

#### Problems

I was asked to get or improve our code search tools by our engineering team

#### What success looks like

Providing a widely used code search solution that meets the needs of the engineering team with low cost and maintenance burden

#### What failure looks like

- Maintaining code search takes a lot of my time
- The developers don't end up using code search
- I can't demonstrate/understand the benefits of code search to developers or to my manager

#### Other tools used

Atlassian suite: Bitbucket, Jira, Confluence, etc.

#### Common objections

- Code search isn't one of our top priorities

#### Personally rewarded by

- TBD

#### Examples

- https://app.hubspot.com/contacts/2762526/contact/15257501/
- https://app.hubspot.com/contacts/2762526/contact/15187551/
- https://app.hubspot.com/contacts/2762526/contact/15722451/

## Dev Infrastructure Head and VP Engineering

#### Who

Person who is in charge of the development experience and tooling decisions and budget for the organization.

#### Role

- Engineering Manager for infrastructure, productivity, developer infrastructure, tools, developer effectiveness, etc. (not just any "Engineering Manager")
- Engineering Velocity
- Head of Developer Experience
- Director of Development Standards

#### Problems

- Onboarding new engineers and sharing knowledge amid engineering team hypergrowth
- Executing on large projects across teams, offices, and timezones
- Monitoring risk around security, compliance, and user data
- Providing the best tools, recruiting, and keeping up with Google/FB/etc.

#### What success looks like

Providing a widely used code search solution that meets the needs of the engineering team with low cost and maintenance burden

#### What failure looks like

- Maintaining code search takes a lot of my time
- The developers don't end up using code search
- I can't demonstrate the benefits of code search to developers or to my manager

#### Common objections

I'm not confident my developers would use this, and I don't see direct/immediate value here

#### Personally rewarded by

- TBD

#### Examples

- https://app.hubspot.com/contacts/2762526/contact/13899851/
- https://app.hubspot.com/contacts/2762526/contact/11876251/
- https://app.hubspot.com/contacts/2762526/contact/16023320/
- https://app.hubspot.com/contacts/2762526/contact/15971251/
- https://app.hubspot.com/contacts/2762526/contact/13517551/

## Others

- Early adopter/tinkerer/open source enthusiast (personality/attitude focused, rather than job/role focused)
- Security engineer
  - https://app.hubspot.com/contacts/2762526/contact/15807951
  - https://app.hubspot.com/contacts/2762526/contact/16031951
