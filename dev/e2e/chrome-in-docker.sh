#!/bin/bash
set +e

docker run -i --rm -p=0.0.0.0:9222:9222 --name=chrome-headless alpeware/chrome-headless-trunk python <<END

import socket
import subprocess
import sys
import thread
import time

def forward(source, destination):
	string = ' '
	while string:
		string = source.recv(1024)
		if string:
			destination.sendall(string)
		else:
			source.shutdown(socket.SHUT_RD)
			destination.shutdown(socket.SHUT_WR)

print('Redirecting port 3080 to $(hostname)')
print('Starting Chrome...')
print('')
sys.stdout.flush()

subprocess.Popen(["/usr/bin/start.sh"])

dock_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
dock_socket.bind(('', 3080))
dock_socket.listen(5)
while True:
	client_socket = dock_socket.accept()[0]
	server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
	server_socket.connect(('$(hostname)', 3080))
	thread.start_new_thread(forward, (client_socket, server_socket))
	thread.start_new_thread(forward, (server_socket, client_socket))

END
