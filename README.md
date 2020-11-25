# JPRQ - Ngrok Alternative

## Get Your Localhost Online and HTTPS

<div align="center">
  <a href="https://www.youtube.com/watch?v=BXuB3cK8R0g"><img width="80%" src="https://img.youtube.com/vi/BXuB3cK8R0g/0.jpg" alt="Demo"></a>
</div>

## How JPRQ is different from Ngrok?

- JPRQ is a free and open-source Ngrok alternative to expose local servers online easily.
- It allows developers to serve unlimited requests to the local server compared to Ngrok's **40 requests/minute** limit.
- It can expose multiple ports at the same time compared to Ngrok with **1 port** limit.

---

## How to install

```bash
$ pip install jprq
```

## How to use

Replace 8000 with the port you want to expose

```
$ jprq 8000
```

Press Ctrl+C to stop it

## [NEW] Custom Subdomain

Replace `subdomain` with a subdomain you want

```
$ jprq 8000 -s=subdomain
```

## How to uninstall

```bash
$ pip uninstall jprq
```

## How JPRQ Works

<img width="100%" src="https://i.imgur.com/1kXPzyd.png">

---

### JPRQ's Server-side implementation in Golang:

<a href="https://github.com/azimjohn/jprq.live">https://github.com/azimjohn/jprq.live</a>

## Limitations

- Cannot expose WebSocket
- Doesn't work with HTTP Polling

## Troubleshooting

- With serving React, Vue or any other modern web apps, make sure you run production server or build the app and serve static files as JPRQ is not capable of exposing Websocket.
