import argparse
import asyncio
from getpass import getuser
from .tunnel import open_tunnel
from . import __version__


def main():
    parser = argparse.ArgumentParser(description='Live And HTTPS Localhost')
    parser.add_argument('-p', '--port', type=int, default=False, help='Port number of the local server')
    parser.add_argument('-V', '--version', action='store_true' ,help='Version number of jprq')

    args = parser.parse_args()

    if args.version:
        print(__version__)
        return
    
    if not args.port:
        print("Please specify -p/--port argument and port.")
        return

    username = getuser()

    loop = asyncio.get_event_loop()
    try:
        loop.run_until_complete(
            open_tunnel(
                ws_uri=f'wss://open.jprq.live/_ws/?username={username}&port={args.port}',
                http_uri=f'http://127.0.0.1:{args.port}',
            )
        )
    except KeyboardInterrupt:
        print("\njprq tunnel closed")
