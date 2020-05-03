import asyncio
import json

import websockets
import ssl
import certifi

from .http import Client

ssl_context = ssl.create_default_context()
ssl_context.load_verify_locations(certifi.where())


async def open_tunnel(ws_uri: str, http_uri):
    async with websockets.connect(ws_uri, ssl=ssl_context) as websocket:
        message = json.loads(await websocket.recv())
        host, token = message["host"], message["token"]
        print(f"Online at https://{host}/")

        client = Client(http_uri, token)
        while True:
            message = json.loads(await websocket.recv())
            asyncio.ensure_future(client.process(message, websocket))
