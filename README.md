# ChatGPT Telegram Bot

Telegram bot that provides the use ChatGPT.

## How to use

1. Clone this repo

```sh
git clone https://github.com/ivanglie/chatgpt-bot.git
```

2. Put in the _.env_ file yours _BOT_TOKEN_, _OPENAI_API_KEY_

You can add Telegram users who will have access to the bot (_BOT_USERS_), if needed.

3. Build docker image

```sh
docker buildx build --platform=linux/amd64 -t my-chatgpt-bot -f Dockerfile .
```

4. Run container with _.env_

```sh
docker run --env-file .env -d -p 80:9000 my-chatgpt-bot
```

## References
* [OpenAI](https://platform.openai.com/)