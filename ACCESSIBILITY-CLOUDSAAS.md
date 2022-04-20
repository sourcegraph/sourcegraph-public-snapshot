Issue link:

https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=accessibility%2Cwcag%2F2.1%2Cwcag%2F2.1%2Fauditing%2C&title=%5BAccessibility%5D%3A+

# 1

[Accessibility Audit] Cloud SaaS: Sign up

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

Note: See document for video walkthrough and additional notes.

With email:

- Open the main sourcegraph page (sourcegraph.com)
- Click on the “Log in” button on the top right
- Redirect to the login form with bunch of controls
- Click on the “Sign up” link on the bottom of the form
- Redirect to the sign up form with bunch of controls
- Click on “Continue with email” link at the bottom of the form
- Fill in email, username, password
- Click on Sign up

With GitHub:

- Same as above, but instead of signing up with email, use the buttons on the top
- Click on “Continue with GitHub” button
- Follow the GitHub prompts, if any (out of our control)
  - We do not need to audit external sites that we cannot control.

With GitLab:

- Same as above, but instead of signing up with email, use the buttons on the top
- Click on “Continue with GitLab" button
- Follow the GitLab prompts, if any (out of our control)
  - We do not need to audit external sites that we cannot control.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 2

[Accessibility Audit] Cloud SaaS: Sign in

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

With email:

- Open the main sourcegraph page (sourcegraph.com)
- Click on the “Log in” button on the top right
- You should see the login form
- Fill in your username, password and click login

With GitHub / GitLab:

- Same as above, but instead of signing in with email, click on the relevant button
- Follow the prompts, if any (out of our control)
  - We do not need to audit external sites that we cannot control.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 3

[Accessibility Audit] Cloud SaaS: Repo syncing notifications

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

- Open the main sourcegraph page (sourcegraph.com) as a logged in user.
- Click on the Cloud button in the navigation bar
- Notifications for repository syncing will be shown in a dropdown list

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 4

[Accessibility Audit] Cloud SaaS: Edit user profile

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

Edit:

- Open the main sourcegraph page (sourcegraph.com) as a logged in user
- Click on the dropdown button on the top right, next to your profile picture
- A dropdown menu should appear, click on “Settings”
- A settings page should appear, click on “Profile”
- Change information in the form.
- Click Save

Delete account:

- Follow the above instructions
- Instead of changing information on the form, find the "Contact support to delete your account" link on the page.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 5

[Accessibility Audit] Cloud SaaS: Add/Verify/Edit/Remove email

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

To access the email page:

- Open the main sourcegraph page (sourcegraph.com) as a logged in user
- Click on the dropdown button on the top right, next to your profile picture
- A dropdown menu should appear, click on “Settings”
- A settings page should appear, click on “Emails”

Add additional email address to the account:

- Type a valid email into the input under “Add email address” label (if the email address is invalid, an error will be shown)
- Click on “Add” button
- The newly added email address should appear in the list of emails on the top of the page

Verify email:

- Add a new email using the steps above.
- Click on the “Verify account” link in the email or copy the link and follow it in the browser
- Once the verification link is visited by the same user account, the email is marked as verified in the list of emails.

Change primary email address:

- Click on the “Primary email address” dropdown
- Select a different email address
- Click on “Save” button

Remove email address:

- Click on the “Remove” link on the right side, next to each email address that can be deleted.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 6

[Accessibility Audit] Cloud SaaS: User privacy - Opt out from autocomplete search results

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

- Open the main sourcegraph page (sourcegraph.com) as a logged in user.
- Click on the dropdown button on the top right, next to your profile picture
- A dropdown menu should appear, click on “Settings”
- A settings page should appear, click on “Privacy”
- Make sure the checkbox “Don’t share my profile in autocomplete search results on Sourcegraph Cloud” is checked
- Click on “Save” button to make the change permanent (otherwise changes will be lost)

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 7

[Accessibility Audit] Cloud SaaS: List your organizations

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

- Open the main sourcegraph page (sourcegraph.com) as a logged in user.
- Click on the dropdown button on the top right, next to your profile picture
- A dropdown menu should appear, click on “Your organizations”
- You should see a view with a list of organizations you are a member of

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 8

[Accessibility Audit] Cloud SaaS: Create a new organization

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

Note: View the above document for a video walkthrough of this journey.

- Click on the button “Create organization” in the organization list page.
- Fill in the form for joining open beta.
  - This form collects some basic information about the customer, size of the company, which code host they use and why they want to use sourcegraph (see screenshot below)
- Click “Continue” button on the bottom right
- A form for organization creation is visible.
- Fill in the input for organization name. Organization ID is created automatically for you.
  - If there is already an existing organization with the same ID, the system generates a new ID and let’s the user edit it manually.
- Accept the terms of service by making sure the checkbox is checked
  - Without this step, it is impossible to continue with organization creation
- Click “Continue” button
- You should see a getting started page for the organization

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 9

[Accessibility Audit] Cloud SaaS: Get started with organizations

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

Note: View the above document for a video walkthrough of this journey.

To access:

- Either the user is redirected to this screen after creating the organization
- Or they can access it by going to the specific organization setting:
  - Open the main sourcegraph page (sourcegraph.com)
  - Click on the dropdown button on the top right, next to your profile picture
  - A dropdown menu should appear, click on “Your organizations”
  - You should see a view with a list of organizations you are a member of
  - Pick one which was not configured yet and hence is still showing the getting started page

To use:

- Work through each step in the getting started screen

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 10

[Accessibility Audit] Cloud SaaS: Update organization settings

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

Accessing:

- Open the main sourcegraph page (sourcegraph.com) as a logged in user.
- Click on the dropdown button on the top right, next to your profile picture
- A dropdown menu should appear, click on “Your organizations”
- You should see a view with a list of organizations you are a member of
- Pick one which was not configured yet and hence is still showing the getting started page
- Go to “Settings” tab on the top and then select “Settings” in the sidebar

Using:

- Make a change to the JSON textarea shown on the page
- Updating:
  - Click the "Save" button
- Cancelling:
  - Click the "Discard" button

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 11

[Accessibility Audit] Cloud SaaS: Delete an organization

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

Accessing:

- Open the main sourcegraph page (sourcegraph.com) as a logged in user.
- Click on the dropdown button on the top right, next to your profile picture
- A dropdown menu should appear, click on “Your organizations”
- You should see a view with a list of organizations you are a member of
- Pick one which was not configured yet and hence is still showing the getting started page
- Go to “Settings” tab on the top and then select “Settings” in the sidebar

Using:

- Click a button on the bottom of the organization settings page “Delete this organization”
- A confirmation dialog is shown which asks if the user really wants to delete the organization
- Type in the organization ID
- If the strings do not match, a red input border is shown
- Click “Delete this organization” in the confirm dialog box to delete the organization and related resources

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 12

[Accessibility Audit] Cloud SaaS: Change organization display name

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

Accessing:

- Open the main sourcegraph page (sourcegraph.com) as a logged in user
- Click on the dropdown button on the top right, next to your profile picture
- A dropdown menu should appear, click on “Your organizations”
- You should see a view with a list of organizations you are a member of
- Pick one which was not configured yet and hence is still showing the getting started page
- Go to “Settings” tab on the top and then select “Profile” in the sidebar

Using:

- Change the display name of the organization in the input
- Click on “Save” button once happy

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 13

[Accessibility Audit] Cloud SaaS: View organization members

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

Accessing:

- Open the main sourcegraph page (sourcegraph.com) as a logged in user
- Click on the dropdown button on the top right, next to your profile picture
- A dropdown menu should appear, click on “Your organizations”
- You should see a view with a list of organizations you are a member of
- Pick one which was not configured yet and hence is still showing the getting started page
- Go to “Members” tab on the top and then select “Members” in the sidebar

Using:

- View list of the organization members.
- List is paginated, so if there are more than 20 members, it will show pagination at the bottom

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 14

[Accessibility Audit] Cloud SaaS: (Admin) Add organization member

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

Accessing:

- Open the main sourcegraph page (sourcegraph.com) as a logged in user
- Click on the dropdown button on the top right, next to your profile picture
- A dropdown menu should appear, click on “Your organizations”
- You should see a view with a list of organizations you are a member of
- Pick one which was not configured yet and hence is still showing the getting started page
- Go to “Members” tab on the top and then select “Members” in the sidebar

Using:

- Ensure user is a site admin.
- Click on the “+ Add member” button at the top of the members list screen
- A dialog should be visible for adding members
- Type in the username of the user, case sensitivity is enforced
- Click on “Add member” button in the dialog, a new member should appear in the members list immediately

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 15

[Accessibility Audit] Cloud SaaS: Invite organization member

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

Accessing:

- Open the main sourcegraph page (sourcegraph.com) as a logged in user
- Click on the dropdown button on the top right, next to your profile picture
- A dropdown menu should appear, click on “Your organizations”
- You should see a view with a list of organizations you are a member of
- Pick one which was not configured yet and hence is still showing the getting started page
- Go to “Members” tab on the top and then select “Members” in the sidebar

Using:

- Click on the "Invite member” button at the top of the members list screen
- A dialog should be visible for inviting members
- Type in the username or email of the user, case sensitivity is enforced
- Click on "Send invite" button in the dialog
  - There is also a possibility to copy the invite link directly from the UI

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 16

[Accessibility Audit] Cloud SaaS: Remove organization member

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

- Go to the members list page
- Click on the settings cog button on the right of the member to be removed
- A dropdown should appear with one option
- Click “Remove from organization…”
- A confirm dialog from the website should appear
- If confirmed, the member should be removed from the organization

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 17

[Accessibility Audit] Cloud SaaS: Leave organization

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

- Go to the members list page
- Click on the settings cog button on the right of the member to be removed
- A dropdown should appear with one option
- Click “Leave organization”
- A confirm dialog from the website should appear
- If confirmed, the current user should be removed from the organization and redirected to user settings page

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)

# 18

[Accessibility Audit] Cloud SaaS: Pending organization invitations

### Steps to replicate journey

Taken from the [user journey list document](https://docs.google.com/document/d/1h3bh6prxPX31aOiWkNsRCvXUbmtakaaj3Kn986uZYkc/edit?usp=sharing). Use this for further context.

List invitations:

- Open the main sourcegraph page (sourcegraph.com)
- Click on the dropdown button on the top right, next to your profile picture
- A dropdown menu should appear, click on “Your organizations”
- You should see a view with a list of organizations you are a member of
- Pick one which was not configured yet and hence is still showing the getting started page
- Go to “Members” tab on the top and then select “Pending invites” in the sidebar

Copy invitation link:

- Follow above steps to list invitations
- Click the cog button on the right from an invitation
- A dropdown with options should appear
- Click the “Copy invite link” option
- A confirmation dialog should appear saying that the link is copied in the clipboard

Resend invitation:

- Follow above steps to list invitations
- Click the cog button on the right from the invitation
- A dropdown with options should appear
- Click the “Resend invite” option
- A confirmation should appear above the list of pending invitations saying that the invitation was resent.

Revoke invitation:

- Go to pending invitations screen
- Click the cog button on the right from the invitation
- A dropdown with options should appear
- Click the “Revoke invite” option
- A confirmation dialog should appear. If confirmed, the invitation is revoked and it should not appear in the list anymore.
- There is also a confirmation above the list of pending invitations saying that the invitation was revoked.

### How to audit

Follow the instructions here: [Auditing a user journey](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey)

**Note:** We are planning on arranging thorough third-party audit in the future, so our focus here is just to find the _majority_ of accessibility issues. Don't worry if you feel you haven't found 100% of potential issues, it is better to focus on the core essentials to complete the journey rather than spending lots of time going through every possible problem.

### How to raise discovered problems

Follow the instructions here: [Raising an accessibility bug](https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#raising-a-bug)
