# Generates the grafana dashboard for network monitoring of API calls to external services
# (ie github, bitbucker, etc). Monitoring is from the caller POV.
# Uses grafanalib (https://github.com/weaveworks/grafanalib) to generate the dashboard json.

# To use this script:
# Make sure grafanalib is installed (pip3 install grafanalib)
# PYTHONPATH=.:$PYTHONPATH  generate-dashboard -o ext_services.json ext_services.dashboard.py

from grafanalib.core import *
from dashboard_common import service_row

dashboard = Dashboard(
  title="API calls to external services (github, bitbucket, etc)",
  rows=[
    service_row('github', 'github'),
  ],
).auto_panel_ids()
