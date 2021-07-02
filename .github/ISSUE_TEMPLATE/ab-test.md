---
name: A/B test tracking issue
about: Run an A/B test on Sourcegraph.com
title: 'A/B test: <name>'
labels: 'AB-test'
assignees: ''

---

### Motivation
  
<!-- What user problem are we trying to solve with the A/B test? What uncertainty are we trying to lift? -->

### Test

<!-- What will be test? What will be the control, and what version(s) will we test? Consider linking to the issues.-->
  
### Metric and experimental design
  
**Metric:** <!-- What metric are we measuring success with? -->
**Smaller significant change:** <!-- What metric change would validate the test? If <metric> reaches at least xx (+yy%), the proposed version will be signinificantly better? -->
**Significance threshold:** 5% <!-- Default to 5% -->
**Duration/size:** How long (on how many users) do we need to run this A/B test for it to be significant? Include a link to a singificance calculator.
  
  
#### Descriptive analytics

<!-- What other analytics will you leverage to understand the test? (Note that only one metric should be used for success) -->
  
### Flag

<!-- What's the name of the feature flag for this test? Also create an `AB-test/<nameOfFlag>` label to tag releated issues -->
