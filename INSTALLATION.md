# Installation

Here we will be providing steps on how to install the Discord bot on your local machine. 

1. Create your own .env file
2. Create the bot on the Discord Developer Portal (Optional)
3. Install the required dependencies

# Create your own .env file

Inorder to provide the bot with the necessary environment variables, you will need to create a .env file in the root directory of the project. 

```env
DISCORD_TOKEN=<YOUR-TOKEN>
GIT_TOKEN=<PERSONAL-GITHUB-TOKEN>
```

**Discord Token** - This is the token that you will get from the Discord Developer Portal. This token is used to authenticate the bot with the Discord API.

**GitHub Token** - This is the token that you will get from the GitHub Developer Portal. This token is used to authenticate the bot with the GitHub API.

# Create the bot on the Discord Developer Portal (Optional)
In the previous step, we mentioned that you will need a Discord token to authenticate the bot with the Discord API. You can get this token by creating a bot on the Discord Developer Portal. However, you don't have to do this as simply printing the bot to your terminal is enough to get started. 

However, if you want to create a bot, you can follow the steps below:

1. Go to the [Discord Developer Portal](https://discord.com/developers/applications)
2. Click on "New Application"
3. Give your application a name
4. Click on "Bot" in the left hand menu
5. Click on "Add Bot" and give the proper permissions
6. Click on "Copy" to copy the token to your clipboard

This is the token that you will use in the .env file. Once you have this set up, make sure to invite the bot to your server and create a channel for it to send messages to.

# Install the required dependencies
Make sure to have docker installed on your machine. You can follow the [installation guide](https://docs.docker.com/get-docker/) to get started.

Once you have followed the docker installtion guide, you want to make sure that all the libraries are installed from the imports and `requirements.txt` file.

# How to contribute
Once you have made any changes to the bot, you can follow the [contribution guide](https://github.com/colorstackatuw/ColorStack-Discord-Bot/blob/main/CONTRIBUTING.md) to get started.