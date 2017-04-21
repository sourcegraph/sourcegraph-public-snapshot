## Packaging this directory (for Sourcegraph employees)

* Clone this repository to your local machine and `cd` into it (`git
clone https://sourcegraph.com/sourcegraph/sourcegraph && cd
./sourcegraph/local-installer`).
* Run `./package.sh <company-id>` to create a `sourcegraph.zip` file which can be sent to the customer. If the customer has requested tracking to be disabled, then run `./package.sh ''` (otherwise usage data will be tracked using the ID specified as the first argument to the script).
* Remember to include the Google Container Registry credentials for accessing Sourcegraph containers in the email.
