<p align="center">
    <img height="140" src="https://user-images.githubusercontent.com/35038240/221522083-1011e567-e2b7-424c-a018-15e965cf8df9.png#gh-light-mode-only">
    <img height="140" src="https://user-images.githubusercontent.com/35038240/221522077-5b1e3eca-ca85-4c9f-93a9-afd39cc93c88.png#gh-dark-mode-only">
</p>

## What's JPRQ?

- JPRQ is a free and open tool for exposing local servers to public network (the internet)
- it can expose TCP protocols, such as HTTP, SSH, major databases (MySQL, Postgres, Redis)

---

## How to install

```bash
$ curl -fsSL https://jprq.io/install.sh | sudo bash
```

## How to use

First obtain auth token from https://jprq.io/auth, then

```bash
$ jprq auth <your-auth-token>
```

Replace 8000 with the port you want to expose

```bash
$ jprq http 8000
```

For exposing any TCP servers, such as SSH

```bash
$ jprq tcp 22
```

Press Ctrl+C to stop it

<a href="https://www.buymeacoffee.com/azimjon" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/default-orange.png" alt="Buy Me A Coffee" height="41" width="174"></a>
