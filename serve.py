#!/usr/bin/env python3
import http.server
import socketserver
import sys

PORT = 8888
BASE64_CONTENT = """cG9ydDogODg4OApwcm94aWVzOgogIC0gbmFtZTogdGVzdAogICAgdHlwZTogc29ja3M1CiAgICBzZXJ2ZXI6IGV4YW1wbGUuY29tCiAgICBwb3J0OiAxMDgw"""

class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-Type', 'text/plain')
        self.end_headers()
        self.wfile.write(BASE64_CONTENT.encode())
    
    def log_message(self, format, *args):
        # Suppress log messages
        pass

if __name__ == "__main__":
    with socketserver.TCPServer(("", PORT), Handler) as httpd:
        print(f"Serving on port {PORT}", file=sys.stderr)
        httpd.serve_forever()