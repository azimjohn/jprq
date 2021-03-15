import asyncio
import sys

import bson
import websockets
import ssl
import certifi

from .http import Client

ssl_context = ssl.create_default_context()
ssl_context.load_verify_locations(certifi.where())


async def open_tunnel(ws_uri: str, http_uri):
    async with websockets.connect(ws_uri, ssl=ssl_context) as websocket:
        message = bson.loads(await websocket.recv())

        if message.get("warning"):
            print(message["warning"], file=sys.stderr)

        if message.get("error"):
            print(message["error"], file=sys.stderr)
            return

        host, token = message["host"], message["token"]
        print(f"\033[32m{'Tunnel Status':<25}Online\033[00m")
        print(f"{'Forwarded':<25}{f'https://{host}/ -> {http_uri}'}\n")
        print(f"\033[2;33mVisit https://{host}/\033[00m\n")

        client = Client(http_uri, token)
        while True:
            message = bson.loads(await websocket.recv())
            asyncio.ensure_future(client.process(message, websocket))
