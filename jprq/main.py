import argparse
import asyncio
from getpass import getuser

from .tunnel import open_tunnel


def main():
    parser = argparse.ArgumentParser(description='Live And HTTPS Localhost')
    parser.add_argument('port', type=int, help='Port Number of The Local Server')

    args = parser.parse_args()
    username = getuser()

    asyncio.get_event_loop().run_until_complete(
        open_tunnel(
            ws_uri=f'wss://open.jprq.live/_ws/?username={username}&port={args.port}',
            http_uri=f'http://127.0.0.1:{args.port}',
        )
    )


if __name__ == '__main__':
    main()
