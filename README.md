# ChatGPT Telegram Bot

Telegram bot that provides the use ChatGPT.

## How to use

1. Clone this repo

```sh
git clone https://github.com/ivanglie/chatgpt-bot.git
```

2. Create and fill in the _.env_ file in _chatgpt-bot_ directory

```
BOT_TOKEN=YOUR_BOT_TOKEN
OPENAI_API_KEY=YOUR_OPENAI_API_KEY
```

You can add users who will have access to the bot, if needed:

```
BOT_USERS=username1,UserName2,user_name3,
```

3. Build docker image

```sh
docker buildx build --platform=linux/arm64/v8 -t my-chatgpt-bot -f Dockerfile .
```

4. Run container with _.env_

```sh
docker run --env-file .env -d -p 80:9000 my-chatgpt-bot
```

## References
* [OpenAI](https://platform.openai.com/)