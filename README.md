# JPRQ - Ngrok Alternative

## Get Your Localhost Online and HTTPS

### Demo

<div align="center">
  <a href="https://www.youtube.com/watch?v=BXuB3cK8R0g"><img width="80%" src="https://img.youtube.com/vi/BXuB3cK8R0g/0.jpg" alt="Demo"></a>
</div>

### How JPRQ is different from Ngrok?

• JPRQ is a free and open-source Ngrok alternative to expose local servers online easily.
• It allows developers to serve unlimited requests to the local server compared to Ngrok's _40 requests/minute_ limit.
• It can expose multiple ports at the same time compared to Ngrok with 1 port limit.

### How JPRQ Works

<img width="100%" src="https://i.imgur.com/1kXPzyd.png">

### JPRQ's Server-side implementation in Golang:

<a href="https://github.com/azimjohn/jprq.live">https://github.com/azimjohn/jprq.live</a>

### How to install

```bash
$ pip install jprq
```

### How to expose local server

```
# Replace 8000 with the port you want to expose
$ jprq 8000
```

### How to uninstall
