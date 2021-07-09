# How to troubleshoot authentication for users whose credentials login as an admin's account on a Sourcegraph instance

This document will attempt to take you through how to troubleshoot and solve for user accounts that do not authenticate correctly and instead, when the user logs in, he or she logs in as the site admin

This document will take you step-by-step through the route to take to troubleshoot and understand why this happened and eventually solve it

## Prerequisites

* This document assumes that Sourcegraph is set up and you are trying to login via an existing username and password
* This document assumes that you are running a Kubernetes deployment
* For the step to exec into the `pgsql` pod, you will need to ensure that you have root access to your organizations deployment. A site-admin can do this for you if you are unable to

## Steps to identify and remedy

### Symptoms

1. You are attempting to login with your username and password. 

2. When authenticated, you realize that the user profile logged in is not yours, but the site admin's. 

### Steps to remedy

1. Exec into the `pgsql` pod and create a shell session by running: `kubectl exec -it <PGSQL-POD> -- psql -U sg
`
2. Check on the site admin's user_id by running: 
`select user_id from user_external_accounts where account_id = <$ADMIN_EMAIL>`

3. Note the `user_id` for the site_admin.
4. Check the user accounts that could potentially share the same `user_id` as the site admin.
5. Remove the user accounts and request them to log in again.
