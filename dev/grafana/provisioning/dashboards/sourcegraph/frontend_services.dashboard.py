# Generates the grafana dashboard for network monitoring of frontend services
# (ie services used by frontend like gitserver). Monitoring is from the frontend client POV.
# Uses grafanalib (https://github.com/weaveworks/grafanalib) to generate the dashboard json.

# To use this script:
# Make sure grafanalib is installed (pip3 install grafanalib)
# PYTHONPATH=.:$PYTHONPATH generate-dashboard -o frontend_services.json frontend_services.dashboard.py

from grafanalib.core import *
from dashboard_common import service_row

dashboard = Dashboard(
  title="Frontend Services > Networking (as seen from frontend)",
  rows=[
    service_row('gitserver', 'gitserver'),
    service_row('repoupdater', 'repoupdater'),
  ],
).auto_panel_ids()
