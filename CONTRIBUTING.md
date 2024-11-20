# Contributing to ColorStack-Discord-Bot

Thank you for investing your time in our project!

## Issues

Whether you have discovered a bug, want a new feature in ColorStack-Discord-Bot, or want to change code, [please create an issue](https://github.com/colorstackatuw/ColorStack-Discord-Bot/issues) or [start a discussion](https://github.com/colorstackatuw/ColorStack-Discord-Bot/discussions)
before any PR. We like to discuss things before implementation. We want to be focused and consider any new features carefully before committing to them. A new idea can be really relevant to you, and we understand it; that's why we try to reflect on every aspect (maintainability, optimizations, and new features).

## How Can You Help ?

- Make sure to install all the dependencies and follow the [installation guide](https://github.com/colorstackatuw/ColorStack-Discord-Bot/blob/main/INSTALLATION.md) to get started.
- If you are new to the project, you can start by looking at the `good first issue` label in the issues section.

## Pull Requests

- Fork the repository and create a new branch from `main` for a new feature or a bug fix.
- Leave a comment on the issue you are working on to let the maintainers know that you are working on it. If there are no updates after **1 week**, please reach out to the maintainers so we can assign the issue to someone else.
- All Git commits are required to be signed and reviewed by at least two maintainers. If you are not getting a response, please reach out to the maintainers after **1 week**.
- If you are creating or editing a new class, method, or function, please make sure to add the appropriate [documentation](https://github.com/colorstackatuw/ColorStack-Discord-Bot/blob/main/DOCUMENTATION.md).
- All tests must be green before merging. Our CI/CD will run [tests](https://github.com/colorstackatuw/ColorStack-Discord-Bot/actions) to ensure everything is OK.
- Before submitting the PR, please make sure to format with `pyproject.toml` to keep it consistent

## Build and Test

ColorStack-Discord-Bot is a Python-based project that uses Docker to run the service, so make sure to have Docker installed on your machine. You can follow the [installation guide](https://github.com/colorstackatuw/ColorStack-Discord-Bot/blob/main/DOCUMENTATION.md#installation) to get started. Once you have followed the installation guide, you can run the following commands to build and test the project:

To build the changes into the Docker container, you can run:

```
docker compose up -d --build
```

**Be careful, as the `DISCORD_CHANNEL` you have the bot in will send messages**

To run unit tests within Docker, you can run:

```
docker-compose run --rm main pytest /app/tests
```

If you want to debug the tests on your local machine, run your IDE debugger within `src/DiscordBot.py` to debug the bot (using print statments instead of `await.send()` would be beneficial). **Just be sure that it also works within Docker**

Although the database connector is private code hosted within the VM, what you can do instead is copy your channel ID within your test discord server and replace the following within `src/DiscordBot.py`

```python
# Get the channels to send the job postings
#db = DatabaseConnector()
channel_ids = [12345] # Your channel id
```

This will allow you send messages to your channel when running on your local machine.
