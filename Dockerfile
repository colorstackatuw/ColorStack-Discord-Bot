# Use an official Python runtime as a parent image
FROM python:3.9.18-bullseye

# Copy the libraries needed for the bot
COPY requirements.txt /app/

# Set the working directory in the container to /app
WORKDIR /app

# Install all the libraries needed for the bot
RUN pip install --no-cache-dir -r requirements.txt

# Copy the entire current directory into the container at /app
COPY . .

# Run DiscordBot.py when the container launches
CMD ["python", "DiscordBot.py"]