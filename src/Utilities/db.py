"""
Database Connector Class

This class provides support to connect to the Oracle database and write the channel to the database.

Prerequisites:
- oracledb: A Python library to interact with the Oracle database.
"""

import os
from datetime import datetime

import discord
import oracledb
from dotenv import load_dotenv


class DatabaseConnector:
    def __init__(self):
        load_dotenv(override=True)
        self.username = os.getenv("DB_USERNAME")
        self.password = os.getenv("DB_PASSWORD")
        self.host = os.getenv("DB_HOST")
        self.conn = None

    def createConnection(self) -> None:
        """
        Create a connection to the Oracle database.

        Returns:
            - cx_Oracle.Connection: The connection to the Oracle database.
        """
        self.conn = oracledb.connect(user=self.username, password=self.password, dsn=self.host)

        if self.conn is None:
            raise Exception("Connection not established.")

    def writeChannel(self, guild: discord.Guild, channel: discord.TextChannel) -> None:
        """
        Write the channel to the Oracle database.

        Parameters:
            - channel: The channel to write to the database.
        """

        sql = """
                INSERT INTO DISCORD.DIM_DISCORD_TOKENS
                (server_id, join_date, server_name, channel_id)
                VALUES (:server_id, :join_date, :server_name, :channel_id)
            """
        params = {
            "server_id": str(guild.id),  
            "join_date": datetime.now(), 
            "server_name": guild.name, 
            "channel_id": str(channel.id), 
        }

        self.createConnection()
        curr = self.conn.cursor()
        curr.execute(sql, params)
        self.conn.commit()
        self.conn.close()

    def getChannels(self) -> list[int]:
        """
        Get the channels from the Oracle database.
        """
        sql = "SELECT CHANNEL_ID FROM DISCORD.DIM_DISCORD_TOKENS GROUP BY CHANNEL_ID"

        self.createConnection()
        curr = self.conn.cursor()
        curr.execute(sql)
        channel_ids = [int(row[0]) for row in curr.fetchall()]
        self.conn.close()
        return channel_ids

    def deleteServer(self, guild: discord.Guild) -> None:
        """
        Delete the channel from the Oracle database.

        Parameters:
            - guild: The guild to delete from the database.
        """
        sql = "DELETE FROM DISCORD.DIM_DISCORD_TOKENS WHERE SERVER_ID = :server_id"
        params = {"server_id": guild.id}

        self.createConnection()
        curr = self.conn.cursor()
        curr.execute(sql, params)
        self.conn.commit()
        self.conn.close()
