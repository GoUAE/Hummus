# Hummus

[![Go Report](https://goreportcard.com/badge/github.com/gouae/hummus)](https://goreportcard.com/report/gouae/hummus) [![Coders(HQ) Discord](https://img.shields.io/badge/Discord-%23golang-blue.svg)](https://golang.ae/discord) 

Hummus is a GoUAE Community project that bridges the Coders(HQ) [WhatsApp](https://golang.ae/whatsapp) communication channels with [Discord](https://golang.ae/discord).

Do note that for now, we aim to just be read-only, sending messages one way (WhatsApp -> Discord). Given the constraints of WhatsApp.

## Quickstart
> [!NOTE]
> Make sure you have [go](https://go.dev) [installed](https://go.dev/dl/).

1. Clone the repository
```sh
git clone https://github.com/GoUAE/Hummus
cd Hummus
```

2. Copy the `.env.example` file to `.env`
```sh
cp .env.example .env
```

3. Fill in the necessary environment variables in the .env file. You'll need to provide values for the following variables:
```env
DISCORD_BOT_TOKENDISCORD_CHANNEL_ID=
DISCORD_WEBHOOK_ID=
DISCORD_WEBHOOK_TOKEN=
DISCORD_FALLBACK_AVATAR_URL=
# We support only 1 WhatsApp chat for now, you'll have to get its ID manually (instructions TBA).
WA_GOUAE_JID=
```

4. Install the dependencies
```sh
go mod download
```

5. Run the project
```sh
go run
```

6. Scan the QR code to log in to the WhatsApp client. 

## Maintainers
- [Gaurav-Gosain](https://github.com/gaurav-gosain)

## LICENSE
This project is licensed under the [MIT License](LICENSE).