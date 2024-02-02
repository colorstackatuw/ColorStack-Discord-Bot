# Use an official Python runtime as a parent image
FROM python:3.9.18-bullseye

# Set the working directory in the container to /app
WORKDIR /app

# Install git, required to clone the repository
RUN apt-get update && apt-get install -y git

# Clone the repository into the /app directory
#RUN git clone https://github.com/DavidSalazar123/ColorStack-Discord-Bot.git /app
COPY . . 

# Assuming requirements.txt is in the root, if it's inside src adjust the path accordingly
RUN pip install --no-cache-dir -r requirements.txt

# Change the working directory to /app/src where DiscordBot.py is located
WORKDIR /app/src

# Run DiscordBot.py when the container launches
CMD ["python", "-u", "DiscordBot.py"]