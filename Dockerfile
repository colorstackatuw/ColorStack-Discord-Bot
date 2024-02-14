# Use an official Python runtime as a parent image
FROM python:3.9.18-bullseye

# Set the working directory in the container to /app
WORKDIR /app

# Install git, required to clone the repository
RUN apt-get update && apt-get install -y git

# Copy the current directory contents into the container at /app
COPY . .

# Install any needed packages specified in requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

# Change the working directory to /app/src where DiscordBot.py is located
WORKDIR /app/src

# Run DiscordBot.py when the container launches
CMD ["python", "-u", "DiscordBot.py"]
