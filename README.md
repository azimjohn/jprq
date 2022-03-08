# JPRQ - Ngrok Alternative 

# ![#f03c15](https://via.placeholder.com/15/f03c15/000000?text=+) Important Notice
<br>
I have brought JPRQ down (only HTTP tunneling) because of a phishing operation.
<br>
I am sad to see my tool serve evil purposes. I will come up with better solution and update soon.
<br>
Sorry for the inconveniences.
<br>

## Get Your Localhost Online and HTTPS


[![jprq demo](https://i.imgur.com/SEKQv7N.png)](https://www.youtube.com/watch?v=BXuB3cK8R0g "jprq demo")

## How JPRQ is different from Ngrok?

- JPRQ is a free and open-source Ngrok alternative to expose local servers online easily.
- It allows developers to serve unlimited requests to the local server compared to Ngrok's **40 requests/minute** limit.
- It can expose multiple ports at the same time compared to Ngrok with **1 port** limit.
- [NEW] it can now expose any TCP protocol, like SSH, MySQL, Redis, etc.
---

## How to install

```bash
$ pip install jprq
```

## How to use

Replace 8000 with the port you want to expose
```
$ jprq http 8000
```

For exposing SSH, WebSocket, Postgresql or any TCP servers
```
$ jprq tcp 22
```

### For windows users:
For exposing HTTP servers
```
 > py -m jprq http 8000
```

For exposing TCP servers
```
 > py -m jprq tcp 22
```

Press Ctrl+C to stop it

## [NEW] Custom Subdomain

Replace `subdomain` with a subdomain you want, works with ony http tunnels.
```
$ jprq http 8000 -s=subdomain
```

## How to uninstall

```bash
$ pip uninstall jprq
```

## How JPRQ Works

<img width="100%" src="https://i.imgur.com/1kXPzyd.png">

---

### JPRQ's Client-side implementation in Python:

<a href="https://github.com/azimjohn/jprq-python-client">https://github.com/azimjohn/jprq-python-client</a>

## Limitations

- HTTP Tunneling cannot expose WebSocket, use TCP Tunneling
- Doesn't work with HTTP Long Polling with HTTP Tunneling, Use TCP Tunneling

## Troubleshooting

- With serving React, Vue or any other modern web apps, make sure you run production server or build the app and serve static files as JPRQ is not capable of exposing Websocket.
- With serving React, Vue or any other modern web apps in development mode, you can use TCP Tunneling

## Self-Hosting 
<a href="https://github.com/robinvandernoord/jprq-server-deploy">This repository</a> offers excellent resources for self-hosting configuration and setup.
