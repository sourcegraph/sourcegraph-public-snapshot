#!/usr/bin/python

# follow tutorial at https://developers.google.com/gmail/api/quickstart/quickstart-python to get CLIENT_SECRET_FILE

############
# to get leads before we started using the new lead form, use:
# copy (select p.requestedupgradeat as date, u.login as user_login, u.name, u.company, u.location, u.homepage_url from person_settings p left join users u on u.uid=p.uid where p.requestedupgradeat is not null) to stdout with csv;

import httplib2
import base64
import re
import csv, sys

from apiclient.discovery import build
from oauth2client.client import flow_from_clientsecrets
from oauth2client.file import Storage
from oauth2client.tools import run_flow
from oauth2client import tools


# Path to the client_secret.json file downloaded from the Developer Console
CLIENT_SECRET_FILE = '/Users/sqs/Downloads/client_secret_527047051561-5ltmc8ndgbvcv1u11ctle17p5cl7oh16.apps.googleusercontent.com.json'

# Check https://developers.google.com/gmail/api/auth/scopes for all available scopes
OAUTH_SCOPE = 'https://www.googleapis.com/auth/gmail.readonly'

# Location of the credentials storage file
STORAGE = Storage('gmail.storage')

# Start the OAuth flow to retrieve credentials
flow = flow_from_clientsecrets(CLIENT_SECRET_FILE, scope=OAUTH_SCOPE)
http = httplib2.Http()

import argparse
flags = argparse.ArgumentParser(parents=[tools.argparser]).parse_args()

# Try to retrieve credentials from storage or run the flow to generate them
credentials = STORAGE.get()
if credentials is None or credentials.invalid:
  credentials = run_flow(flow, STORAGE, flags, http=http)

# Authorize the httplib2.Http object with our credentials
http = credentials.authorize(http)

# Build the Gmail service from discovery
gmail_service = build('gmail', 'v1', http=http)

# Retrieve a page of threads


pat = r"(?P<email>[^\s]*) signed-up on .*\nName: (?P<name>.*)\nPhone Number: (?P<phone>.*)\nCompany: (?P<company>.*)\nTitle: (?P<title>.*)\nTeam Size: (?P<team_size>.*)\n\nLanguages: (?P<langs>.*)\n"

csvw = csv.DictWriter(sys.stdout, fieldnames=['date', 'email', 'name', 'phone', 'company', 'title', 'team_size', 'langs'])
csvw.writeheader()

pageToken=None
while True:
  msgs = gmail_service.users().messages().list(userId='me', q='subject:"beta signup" from:notify@sourcegraph.com', pageToken=pageToken).execute()
  if not msgs['messages']:
    break
  if 'nextPageToken' in msgs:
    pageToken = msgs['nextPageToken']
  else:
    pageToken = 'DONE'
  for msgId in msgs['messages']:
    msg = gmail_service.users().messages().get(userId='me', id=msgId['id']).execute()
    p = msg['payload']
    
    dt = None
    for h in p['headers']:
      if h['name'] == 'Date':
        dt = h['value']

    d = p['body']['data']
    try:
      body0 = base64.b64decode(d.encode('utf8'))
      body = body0.decode('utf8').replace("\r", "")
      m = re.search(pat, body, re.M)
      info = m.groupdict()
      info['date'] = dt
      if m:
        csvw.writerow(info)
      else:
        sys.stderr.write("FAIL\n")
    except Exception as e:
      try:
        body0 = base64.b64decode(d.encode('utf8'))
        body = body0.decode('utf8').replace("\r", "")
        print(body)
      except:
        pass
      sys.stderr.write("FAIL b64: %r\n" % e)
  if pageToken == 'DONE':
    break
