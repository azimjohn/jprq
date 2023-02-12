# JPRQ - Ngrok Alternative 

<p align="center">
    <img height="140" src="https://user-images.githubusercontent.com/35038240/160110182-f38a29f8-1058-48af-8a80-b97b0103c71f.jpg">
</p>

## How JPRQ is different from Ngrok?

- JPRQ is a free and open-source Ngrok alternative to expose local servers online easily.
- It allows developers to serve unlimited requests to the local server compared to Ngrok's **40 requests/minute** limit.
- It can expose multiple ports at the same time compared to Ngrok with **1 port** limit.
- [NEW] it can now expose any TCP protocol, like SSH, MySQL, Redis, etc.
---

## How to install

```bash
$ //TODO
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

Press Ctrl+C to stop it

## How JPRQ Works

<img width="100%" src="https://i.imgur.com/1kXPzyd.png">
---
